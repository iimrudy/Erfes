package main

import (
	"context"
	"erfes/commands"
	"github.com/go-telegram/bot"
	"io"
	"log"
	"os"
	"os/signal"
)

type PassThru struct {
	io.Reader
	TransferredBytes int64 // Total # of bytes transferred
	CallBack         func(int64)
}

func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.TransferredBytes += int64(n)
	if pt.CallBack != nil {
		pt.CallBack(pt.TransferredBytes)
	}
	return n, err
}

func main() {
	log.Default().SetFlags(log.LstdFlags | log.Lshortfile)
	/*client := youtube.Client{}
	video, err := client.GetVideo("https://www.youtube.com/watch?v=k7jkoIEkr9I")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Video %+v\n", video.Title)
	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := client.GetStream(video, &formats[0])
	f := &formats[0]
	// calc file size from bitrate
	duration := video.Duration
	fileSize := int64(float64(f.Bitrate) * duration.Seconds() / 8)

	aStream := PassThru{stream, 0, func(bytes int64) {
		// download progress
		fmt.Printf("Downloaded %d bytes of %d bytes (%d%%)\r", bytes, fileSize, int64(float64(bytes)/float64(fileSize)*100))
	}}

	if err != nil {
		panic(err)
	}

	file, err := os.Create("video.mp4")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, &aStream)
	if err != nil {
		panic(err)
	}*/

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
	b.Start(ctx)

}
