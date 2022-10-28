package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/FluorescentTouch/review-picker/internal/bot"
	"github.com/FluorescentTouch/review-picker/internal/statuses"
	"github.com/FluorescentTouch/review-picker/internal/users"
)

const (
	commandAdd            = "/add"
	commandAddDescription = "Add new user to list of reviewers"

	commandVacationStart            = "/vac_start"
	commandVacationStartDescription = "Deactivate user for review picks"

	commandVacationEnd            = "/vac_end"
	commandVacationEndDescription = "Activate user for review picks"

	commandPickReviewer            = "/pick"
	commandPickReviewerDescription = "Pick a random reviewer from active pool. Message has to contain PR link"

	commandPickTwoReviewer            = "/pick2"
	commandPickTwoReviewerDescription = "Pick two random reviewers from active pool. Message has to contain PR link"

	commandHelp            = "/help"
	commandHelpDescription = "Show this help"

	commandStart            = "/start"
	commandStartDescription = "Show this help"

	oneApproveHeader       = "[1] approve required for pr's:\n"
	twoApproveHeader       = "[2] approves required for pr's:\n"
	commentNothingToReview = "nothing to review"
)

var errUnknownCommand = errors.New("unknown command")

type Handlers struct {
	bot   bot.Bot
	users users.Users

	ctx    context.Context
	cancel context.CancelFunc

	mapping map[string]handler
}

type handler struct {
	h func(message bot.Message) // handler
	d string                    // description
}

func NewHandlers(bot bot.Bot, usersService users.Users) *Handlers {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Handlers{
		bot:    bot,
		users:  usersService,
		ctx:    ctx,
		cancel: cancel,
	}
	h.mapping = map[string]handler{
		commandAdd:             {h: h.AddUserToList, d: commandAddDescription},
		commandVacationStart:   {h: h.StartVacationToUser, d: commandVacationStartDescription},
		commandVacationEnd:     {h: h.StopVacationToUser, d: commandVacationEndDescription},
		commandPickReviewer:    {h: h.PickReviewer, d: commandPickReviewerDescription},
		commandPickTwoReviewer: {h: h.PickTwoReviewers, d: commandPickTwoReviewerDescription},
		commandStart:           {h: h.ShowHelp, d: commandStartDescription},
		commandHelp:            {h: h.ShowHelp, d: commandHelpDescription},
	}
	return h
}

func (h *Handlers) Stop() {
	h.cancel()
}

func (h *Handlers) StartCycle() {
	log.Println("starting listening cycle...")

	for {
		select {
		case <-h.ctx.Done():
			return
		case msg, ok := <-h.bot.Messages():
			if !ok {
				return
			}

			if !msg.Command {
				continue
			}

			// casual usecase
			command := strings.Split(msg.Text, " ")[0]
			handler, ok := h.mapping[command]
			if ok {
				handler.h(msg)
				continue
			}

			// menu usecase
			command = strings.Split(msg.Text, "@")[0]
			handler, ok = h.mapping[command]
			if ok {
				handler.h(msg)
				continue
			}

			// unknown command
			h.bot.ReplyTo(msg.ChatID, msg.ReplyID, errUnknownCommand.Error())
			continue
		}
	}
}

func (h *Handlers) ShowHelp(msg bot.Message) {
	b := strings.Builder{}
	b.WriteString("Commands to use:\n")
	for cmd, h := range h.mapping {
		b.WriteString(fmt.Sprintf("%s: %s\n", cmd, h.d))
	}
	h.bot.ReplyTo(msg.ChatID, msg.ReplyID, b.String())
}

func (h *Handlers) PickReviewer(msg bot.Message) {
	u, err := h.users.Rand(msg.ChatID, msg.Author)
	if err != nil {
		h.bot.ReplyTo(msg.ChatID, msg.ReplyID, err.Error())
		return
	}
	urls := findUrls(msg.Text)
	if len(urls) == 0 {
		h.bot.ReplyTo(msg.ChatID, msg.ReplyID, "nothing to review")
		return
	}
	builder := strings.Builder{}
	builder.WriteString(statuses.StatusWait.String() + " " + oneApproveHeader)
	for i := range urls {
		builder.WriteString(urls[i] + "\n")
	}
	builder.WriteString(fmt.Sprintf("approver: @%s", u.Name))
	h.bot.ReplyWithKeyboardTo(
		msg.ChatID, msg.ReplyID, builder.String(),
	)
}

func (h *Handlers) PickTwoReviewers(msg bot.Message) {
	randomUsers, err := h.users.Rand2(msg.ChatID, msg.Author)
	if err != nil {
		h.bot.ReplyTo(msg.ChatID, msg.ReplyID, err.Error())
		return
	}
	urls := findUrls(msg.Text)
	if len(urls) == 0 {
		h.bot.ReplyTo(msg.ChatID, msg.ReplyID, commentNothingToReview)
		return
	}
	builder := strings.Builder{}
	builder.WriteString(statuses.StatusWait.String() + " " + twoApproveHeader)
	for i := range urls {
		builder.WriteString(urls[i] + "\n")
	}
	h.bot.ReplyWithKeyboardTo(msg.ChatID, msg.ReplyID, builder.String()+fmt.Sprintf("approver: @%s", randomUsers[0].Name))
	h.bot.ReplyWithKeyboardTo(msg.ChatID, msg.ReplyID, builder.String()+fmt.Sprintf("approver: @%s", randomUsers[1].Name))
}

func (h *Handlers) AddUserToList(msg bot.Message) {
	err := h.users.AddUser(msg.ChatID, users.User{Name: msg.Author})
	if err != nil {
		h.bot.ReplyTo(msg.ChatID, msg.ReplyID, err.Error())
		return
	}
	h.bot.Send(msg.ChatID, fmt.Sprintf("user @%s added to reviewers", msg.Author))
}

func (h *Handlers) StartVacationToUser(msg bot.Message) {
	err := h.users.AddUser(msg.ChatID, users.User{Name: msg.Author, Vacation: true})
	if err != nil {
		h.bot.ReplyTo(msg.ChatID, msg.ReplyID, err.Error())
		return
	}
	h.bot.Send(msg.ChatID, fmt.Sprintf("user @%s уехал отдыхать", msg.Author))
}

func (h *Handlers) StopVacationToUser(msg bot.Message) {
	err := h.users.AddUser(msg.ChatID, users.User{Name: msg.Author, Vacation: false})
	if err != nil {
		h.bot.ReplyTo(msg.ChatID, msg.ReplyID, err.Error())
		return
	}
	h.bot.Send(msg.ChatID, fmt.Sprintf("user @%s вернулся из отпуска", msg.Author))
}

func findUrls(text string) []string {
	out := make([]string, 0, 2)
	entries := strings.Split(text, " ")
	for _, entry := range entries {
		if strings.HasPrefix(entry, "http") {
			out = append(out, entry)
		}
	}
	return out
}
