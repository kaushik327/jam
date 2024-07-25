package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	mp3 "github.com/gopxl/beep/mp3"
	speaker "github.com/gopxl/beep/speaker"
	youtube "github.com/kkdai/youtube/v2"
)

type YTSong struct {
	VideoID string        `json:"videoID"` // 11-character ID in YouTube URL
	loaded  chan struct{} // closed once .mp3 file is downloaded
}

func load(s YTSong) {

	if _, err := os.Stat(fmt.Sprintf("songs/%s.mp3", s.VideoID)); err == nil {
		close(s.loaded)
		return
	}

	client := youtube.Client{}

	video, err := client.GetVideo(s.VideoID)
	if err != nil {
		log.Fatalf("Failed to get video: %v", err)
	}

	formats := video.Formats.Type("audio/mp4")
	if len(formats) == 0 {
		panic("no audio/mp4 format available")
	}

	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		log.Fatalf("Failed to get m4a stream: %v", err)
	}
	defer stream.Close()

	if err = os.MkdirAll("songs", os.ModePerm); err != nil {
		log.Fatalf("Failed to create songs directory: %v", err)
	}

	cmd := exec.Command("ffmpeg", "-y", "-i", "pipe:0", "-q:a", "0", fmt.Sprintf("songs/%s.mp3", s.VideoID))
	cmd.Stdin = stream
	if err = cmd.Run(); err != nil {
		log.Fatalf("Failed to run ffmpeg: %v", err)
	}

	close(s.loaded)
}

func play(s YTSong) {
	<-s.loaded

	f, err := os.Open(fmt.Sprintf("songs/%s.mp3", s.VideoID))
	if err != nil {
		log.Fatalf("Failed to open new file: %v", err)
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatalf("Failed to play mp3: %v", err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.PlayAndWait(streamer)
}
