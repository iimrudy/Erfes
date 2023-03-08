package templates

import (
	"github.com/kkdai/youtube/v2"
	"html/template"
	"strings"
)

var VideoInfo *template.Template = nil

func init() {

	VideoInfo = template.Must(template.ParseFiles("templates/video_info.template"))
}

func GetVideoInfoTemplate(video *youtube.Video, prefix string, suffix string) string {
	buffer := new(strings.Builder)
	VideoInfo.Execute(buffer, &struct {
		Prefix   string
		Title    string
		Author   string
		Duration string
		ID       string
		Suffix   string
	}{
		Prefix:   prefix,
		Title:    video.Title,
		Author:   video.Author,
		Duration: video.Duration.String(),
		ID:       video.ID,
		Suffix:   suffix,
	})
	return buffer.String()
}
