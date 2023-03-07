package commands

import (
	"context"
	"erfes/utils"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/kkdai/youtube/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"log"
	"os"
	"strconv"
)

var client = youtube.Client{Debug: true}
var spinner = []rune("↖↗↘↙")

func DownloadAudioCommand(ctx context.Context, b *bot.Bot, update *models.Update, args []string) error {

	if len(args) != 1 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			Text:   "Invalid arguments. Usage: /audio <url>",
			ChatID: update.Message.Chat.ID,
		})
		return nil
	}

	msg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Fetching url information from " + args[0],
		ChatID: update.Message.Chat.ID,
	})

	video, err := client.GetVideo(args[0])
	if err != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    update.Message.Chat.ID,
			MessageID: msg.ID,
			Text:      "Error: " + err.Error(),
		})
		return err
	}

	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.Message.Chat.ID,
		MessageID: msg.ID,
		Text:      genMsg(video, "Video Info", "Downloading \\.\\.\\."),
		ParseMode: models.ParseModeMarkdown,
	})

	if err != nil {
		return err
	}

	go handleDownloadAudio(ctx, b, msg, video)

	return nil
}

func handleDownloadAudio(ctx context.Context, b *bot.Bot, msg *models.Message, video *youtube.Video) {
	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    msg.Chat.ID,
			MessageID: msg.ID,
			Text:      "Error: " + err.Error(),
		})
		return
	}
	duration := video.Duration
	fileSize := int64(float64(formats[0].Bitrate) * duration.Seconds() / 8)
	rotate := 0
	oldPercentage := int64(0)

	stream1 := &utils.PassThru{
		Reader:           stream,
		TransferredBytes: 0,
		CallBack: func(bytes int64) {
			percentage := int64(float64(bytes) / float64(fileSize) * 100)
			if percentage != oldPercentage {
				b.EditMessageText(ctx, &bot.EditMessageTextParams{
					ChatID:    msg.Chat.ID,
					MessageID: msg.ID,
					Text:      genMsg(video, "Downloading", "Status \\( `"+strconv.FormatInt(percentage, 10)+"` \\) % "+string(spinner[rotate%len(spinner)])),
					ParseMode: models.ParseModeMarkdown,
				})
				oldPercentage = percentage
				rotate++
			}
			/*b.SendAudio(ctx, &bot.SendAudioParams{
				ChatID:    msg.Chat.ID,
				Audio:     stream,
			})*/

		},
	}
	fileName := "./temp/" + uuid.New().String()

	file, err := os.Create(fileName + ".mp4")
	if err != nil {
		log.Printf("Error creating file: %v\n", err)
	}
	defer file.Close()
	defer os.Remove(fileName + ".mp4")

	_, err = io.Copy(file, stream1)
	if err != nil {
		log.Printf("Error copying file: %v\n", err)
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
		Text:      genMsg(video, "Converting To Audio", ""),
		ParseMode: models.ParseModeMarkdown,
	})

	// Create a new FFmpeg command
	cmd := ffmpeg.Input(fileName + ".mp4").Output(fileName + ".mp3").GlobalArgs("-y")

	// Run the command and check for errors
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error converting file: %v\n", err)
		return
	}
	// calc file2 size
	file2Info, err := os.Stat(fileName + ".mp3")
	if err != nil {
		log.Println(err)
		return
	}
	fileSize = file2Info.Size()

	fileData, errReadFile := os.Open("./" + fileName + ".mp3")
	if errReadFile != nil {
		fmt.Printf("error read file, %v\n", errReadFile)
		return
	}
	defer fileData.Close()
	defer os.Remove(fileName + ".mp3")

	file2 := utils.PassThru{
		Reader:           fileData,
		TransferredBytes: 0,
		CallBack: func(bytes int64) {
			percentage := int64(float64(bytes) / float64(fileSize) * 100)
			if percentage != oldPercentage {
				b.EditMessageText(ctx, &bot.EditMessageTextParams{
					ChatID:    msg.Chat.ID,
					MessageID: msg.ID,
					Text:      genMsg(video, "Uploading", "Status \\( `"+strconv.FormatInt(percentage, 10)+"` \\) % "+string(spinner[rotate%len(spinner)])),
					ParseMode: models.ParseModeMarkdown,
				})
				oldPercentage = percentage
				rotate++
			}

		},
	}

	params := &bot.SendAudioParams{
		ChatID:    msg.Chat.ID,
		Audio:     &models.InputFileUpload{Filename: video.Title + ".mp3", Data: &file2},
		Caption:   genMsg(video, "", ""),
		ParseMode: models.ParseModeMarkdown,
	}
	b.SendAudio(ctx, params)
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})

	fmt.Println("File conversion successful")
}

func genMsg(video *youtube.Video, prefix, suffix string) string {
	return prefix + "\n\n🎫 **Title** \n`" + video.Title + "`\n\n🫂 **Author** \n`" + video.Author + "`\n\n🕦 **Duration** \n`" + video.Duration.String() + "`\n\n🔗 [URL](https://youtu.be/" + video.ID + ")\n\n" + suffix
}
