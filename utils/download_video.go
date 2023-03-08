package utils

import (
	"context"
	"erfes/templates"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/kkdai/youtube/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"os"
	"strconv"
)

var client = youtube.Client{Debug: os.Getenv("DEBUG") == "true"}
var spinner = []rune("↖↗↘↙")

func FetchVideoInfo(ctx context.Context, url string, b *bot.Bot, msg *models.Message) (video *youtube.Video, err error) {
	video, err = client.GetVideoContext(ctx, url)
	if err != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    msg.Chat.ID,
			MessageID: msg.ID,
			Text:      "Error: " + err.Error(),
		})
	}
	return
}

func DownloadVideo(ctx context.Context, b *bot.Bot, msg *models.Message, video *youtube.Video) (string, error) {
	_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:                msg.Chat.ID,
		MessageID:             msg.ID,
		Text:                  templates.GetVideoInfoTemplate(video, "Downloading the video", ""),
		ParseMode:             models.ParseModeMarkdown,
		DisableWebPagePreview: true,
	})

	if err != nil {
		return "", err
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := client.GetStreamContext(ctx, video, &formats[0])
	if err != nil {
		b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    msg.Chat.ID,
			MessageID: msg.ID,
			Text:      "Error: " + err.Error(),
		})
		return "", err
	}
	duration := video.Duration
	fileSize := int64(float64(formats[0].Bitrate) * duration.Seconds() / 8)
	rotate := 0
	oldPercentage := int64(0)

	stream1 := &PassThru{
		Reader:           stream,
		TransferredBytes: 0,
		CallBack: func(bytes int64) {
			percentage := int64(float64(bytes) / float64(fileSize) * 100)
			if percentage != oldPercentage {
				percentageStr := "Status \\( `" + strconv.FormatInt(percentage, 10) + "` \\) % "
				status := string(spinner[rotate%len(spinner)])
				txt := templates.GetVideoInfoTemplate(video, "Downloading", percentageStr+status)

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

	fileName := "./temp/" + uuid.New().String()

	file, err := os.Create(fileName + ".mp4")
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, stream1)
	if err != nil {
		return "", err
	}
	return fileName + ".mp4", nil
}

func DownloadAudio(ctx context.Context, b *bot.Bot, msg *models.Message, video *youtube.Video) (string, error) {
	fileName, err := DownloadVideo(ctx, b, msg, video)
	if err != nil {
		return "", err
	}
	defer os.Remove(fileName)
	_, err = b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:                msg.Chat.ID,
		MessageID:             msg.ID,
		Text:                  templates.GetVideoInfoTemplate(video, "Converting to audio", ""),
		ParseMode:             models.ParseModeMarkdown,
		DisableWebPagePreview: true,
	})

	if err != nil {
		return "", err
	}

	// remove .mp4
	fileNameMp3 := fileName[:len(fileName)-4] + ".mp3"

	// Create a new FFmpeg command
	cmd := ffmpeg.Input(fileName).Output(fileNameMp3).GlobalArgs("-y")

	// Run the command and check for errors
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return fileNameMp3, nil
}
