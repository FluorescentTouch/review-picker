package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FluorescentTouch/review-picker/internal/bot"
	"github.com/FluorescentTouch/review-picker/internal/storage"
	"github.com/FluorescentTouch/review-picker/internal/users"
)

var token = flag.String("token", "", "tg bot API token")

func init() {
	flag.Parse()
}

func main() {
	if token == nil || *token == "" {
		log.Panic("no token provided")
	}

	s, err := storage.New()
	if err != nil {
		log.Panic(err)
	}
	defer s.Close()

	u, err := users.NewUsers(s)
	if err != nil {
		log.Panic(err)
	}

	b, err := bot.NewBot(*token)
	if err != nil {
		log.Panic(err)
	}
	defer b.Close()

	h := NewHandlers(b, u)
	go h.StartCycle()
	defer h.Stop()

	n := NewNotifier(b, u)
	n.NotifyAll("🏆️ Я поднялся! 🏆️")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Gracefully shutting down service")
	n.NotifyAll("❗️ Я упал! ❗️")
}
