package commands

import (
	"context"
	"erfes/templates"
	"erfes/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kkdai/youtube/v2"
	"log"
	"os"
	"strconv"
)

var spinner = []rune("↖↗↘↙")

type result struct {
	FileName string
	Err      error
	Video    *youtube.Video
}

func DownloadAudioCommand(ctx context.Context, b *bot.Bot, update *models.Update, args []string) error {

	if len(args) != 1 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			Text:   "Invalid arguments. Usage: /audio <url>",
			ChatID: update.Message.Chat.ID,
		})
		return nil
	}

	msg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Fetching video information from " + args[0],
		ChatID: update.Message.Chat.ID,
	})

	resultChan := make(chan result)

	go func() {
		video, err := utils.FetchVideoInfo(ctx, args[0], b, msg)
		if err != nil {
			log.Println("[ERROR]", err)
			resultChan <- result{Err: err}
			return
		}
		fileName, err := utils.DownloadAudio(ctx, b, msg, video)
		if err != nil {
			resultChan <- result{Err: err}
		}
		resultChan <- result{FileName: fileName, Video: video}

	}()

	res := <-resultChan
	if res.Err != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    msg.Chat.ID,
			MessageID: msg.ID,
			Text:      "Error: " + res.Err.Error(),
		})
		return res.Err
	}

	return uploadAudio(ctx, b, msg, res)
}

func uploadAudio(ctx context.Context, b *bot.Bot, msg *models.Message, res result) error {

	// calc file2 size
	file2Info, err := os.Stat(res.FileName)
	if err != nil {
		return err
	}

	fileSize := file2Info.Size()

	fileData, errReadFile := os.Open("./" + res.FileName)
	if errReadFile != nil {
		return errReadFile
	}

	defer func() {
		fileData.Close()
		os.Remove(res.FileName)
	}()

	oldPercentage := int64(0)
	rotate := 0

	file2 := &utils.PassThru{
		Reader:           fileData,
		TransferredBytes: 0,
		CallBack: func(bytes int64) {
			percentage := int64(float64(bytes) / float64(fileSize) * 100)
			if percentage != oldPercentage {
				percentageStr := "Status \\( `" + strconv.FormatInt(percentage, 10) + "` \\) % "
				status := string(spinner[rotate%len(spinner)])
				txt := templates.GetVideoInfoTemplate(res.Video, "Downloading", percentageStr+status)

				b.EditMessageText(ctx, &bot.EditMessageTextParams{
					ChatID:                msg.Chat.ID,
					MessageID:             msg.ID,
					Text:                  txt,
					ParseMode:             models.ParseModeMarkdown,
					DisableWebPagePreview: true,
				})
				oldPercentage = percentage
				rotate++
			}

		},
	}

	params := &bot.SendAudioParams{
		ChatID:    msg.Chat.ID,
		Audio:     &models.InputFileUpload{Filename: res.Video.Title + ".mp3", Data: file2},
		Caption:   templates.GetVideoInfoTemplate(res.Video, "", ""),
		ParseMode: models.ParseModeMarkdown,
	}

	_, err = b.SendAudio(ctx, params)
	if err != nil {
		return err
	}
	_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if err != nil {
		return err
	}
	return nil
}
