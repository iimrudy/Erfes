package commands

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log"
	"strings"
	"time"
)

type CommandHandler struct {
	Commands []CommandInfo
}

func (ch *CommandHandler) AddCommand(command CommandInfo) {
	ch.Commands = append(ch.Commands, command)
}

func (ch *CommandHandler) HandleCommands(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	if update.Message.Text == "" {
		return
	}

	// split spaces into args from update.Message.Text
	args := strings.Split(update.Message.Text, " ")
	if len(args) == 0 {
		return
	}

	// help command is injected by default
	if strings.EqualFold(args[0], "/help") {
		msg := "Available commands:\n"
		for _, command := range ch.Commands {
			params := ""
			for _, param := range command.Parameters {
				if param.Optional {
					params += "[" + param.Name + "] "
				} else {
					params += "<" + param.Name + "> "
				}
			}
			msg += command.Name + " " + params + " - " + command.Desc + "\n"
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			Text:   msg,
			ChatID: update.Message.Chat.ID,
		})
		return
	}

	for _, command := range ch.Commands {
		log.Printf("[INFO] Checking command %s\n", command.Name)
		if strings.EqualFold(args[0], command.Name) {
			start := time.Now()
			err := command.Handler(ctx, b, update, args[1:])
			if err != nil {
				log.Printf("[ERROR] Command %s failed to proccess: %s\n", command.Name, err)
			}
			log.Printf("[INFO] Command %s took %sms to proccess\n", command.Name, time.Since(start))
			return
		}
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Command not found, type /help to see all commands.",
		ChatID: update.Message.Chat.ID,
	})
}

func NewCommandHandler() *CommandHandler {
	handler := &CommandHandler{}

	// Start command
	handler.AddCommand(CommandInfo{
		Handler: StartCommand,
		Name:    "/start",
		Desc:    "Starts the bot",
	})

	// /audio command
	handler.AddCommand(CommandInfo{
		Handler: DownloadAudioCommand,
		Name:    "/audio",
		Desc:    "Downloads audio from a youtube video",
		Parameters: []Parameter{
			{
				"URL",
				false,
			},
		},
	})

	handler.AddCommand(CommandInfo{
		Handler: DownloadVideoCommand,
		Name:    "/video",
		Desc:    "Downloads video from a youtube video",
		Parameters: []Parameter{
			{
				"URL",
				false,
			},
		},
	})

	return handler
}
