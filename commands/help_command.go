package commands

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func StartCommand(ctx context.Context, b *bot.Bot, update *models.Update, _ []string) error {
	b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Hello, " + update.Message.From.FirstName + "!\n\nType /help to see available commands.",
		ChatID: update.Message.Chat.ID,
	})
	return nil
}
