package main

import (
	"context"
	"erfes/commands"
	"erfes/templates"
	"github.com/go-telegram/bot"
	"log"
	"os"
	"os/signal"
)

func main() {
	log.Default().SetFlags(log.LstdFlags | log.Lshortfile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	handler := commands.NewCommandHandler()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler.HandleCommands),
	}
	b, err := bot.New(os.Getenv("BOT_TOKEN"), opts...)
	if err != nil {
		panic(err)
	}

	// get chat of the bot
	botInfo, err := b.GetMe(ctx)
	if err != nil {
		panic(err)
	}

	log.Printf("Bot started, bot username https://t.me/%s\n", botInfo.Username)
	log.Println(templates.VideoInfo)
	b.Start(ctx)
}
