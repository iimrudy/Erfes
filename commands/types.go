package commands

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Command func(ctx context.Context, b *bot.Bot, update *models.Update, args []string) error

type CommandInfo struct {
	Handler    Command
	Name       string
	Desc       string
	Parameters []Parameter
}

type Parameter struct {
	Name     string
	Optional bool
}
