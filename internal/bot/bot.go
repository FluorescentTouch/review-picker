package bot

import (
	"context"
	"log"

	"github.com/FluorescentTouch/review-picker/internal/statuses"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	ctx    context.Context
	cancel context.CancelFunc
}

type Message struct {
	Text    string
	Command bool
	ReplyID int
	Author  string
	ChatID  int64
}

func NewBot(token string) (Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return Bot{}, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return Bot{api: bot, ctx: ctx, cancel: cancel}, nil
}

func (b Bot) Close() {
	b.cancel()
}

func (b Bot) Messages() chan Message {
	msg := make(chan Message, 100)
	go func() {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates := b.api.GetUpdatesChan(u)
		for {
			select {
			case <-b.ctx.Done():
				return
			case update, ok := <-updates:
				if !ok {
					return
				}
				if update.Message != nil {
					// receive and process command
					msg <- Message{
						Text:    update.Message.Text,
						Command: update.Message.IsCommand(),
						ReplyID: update.Message.MessageID,
						Author:  update.Message.From.UserName,
						ChatID:  update.FromChat().ID,
					}
				} else if update.CallbackQuery != nil {
					// Respond to the callback query, telling Telegram to show the user
					// a message with the data received.
					callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
					if _, err := b.api.Request(callback); err != nil {
						log.Printf("Bot.api.Request error: %v\n", err)
						continue
					}

					// edit original message
					replaceMessage := statuses.ReplaceStatusInText(update.CallbackQuery.Message.Text, statuses.Status(update.CallbackQuery.Data))
					editConfig := tgbotapi.NewEditMessageText(
						update.CallbackQuery.Message.Chat.ID,
						update.CallbackQuery.Message.MessageID,
						replaceMessage,
					)
					editConfig.DisableWebPagePreview = true
					keyboard, ok := genKeyboardByStatus(statuses.Status(update.CallbackQuery.Data))
					if ok {
						editConfig.ReplyMarkup = &keyboard
					}

					_, err := b.api.Send(editConfig)
					if err != nil {
						log.Printf("Bot.api.Send for callback error: %v\n", err)
						continue
					}
				}
			}
		}
	}()
	return msg
}

func (b Bot) ReplyWithKeyboardTo(chatID int64, msgID int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = msgID
	msg.DisableWebPagePreview = true
	keyboard, ok := genKeyboardByStatus(statuses.StatusWait)
	if ok {
		msg.ReplyMarkup = keyboard
	}

	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Bot.ReplyWithKeyboardTo error: %v\n", err)
	}
}

func (b Bot) ReplyTo(chatID int64, msgID int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = msgID
	msg.DisableWebPagePreview = true

	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Bot.ReplyTo error: %v\n", err)
	}
}

func (b Bot) Send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true

	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Bot.Send error: %v\n", err)
	}
}

func genKeyboardByStatus(s statuses.Status) (tgbotapi.InlineKeyboardMarkup, bool) {
	if s.IsFinal() {
		return tgbotapi.InlineKeyboardMarkup{}, false
	}

	buttons := make([]tgbotapi.InlineKeyboardButton, 0, 3)
	for _, goal := range statuses.StatusTransitions[s] {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(goal.Description(), goal.String()))
	}

	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...)), true
}
