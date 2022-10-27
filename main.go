package main

import (
	"flag"
	"log"

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
	log.Println("starting listening cycle...")
	h.StartCycle()
}
