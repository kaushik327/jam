package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gopxl/beep/speaker"
	"github.com/gorilla/websocket"
)

type PlayRequest struct {
	Type    string `json:"type"`
	VideoID string `json:"videoID,omitempty"`
}

type QueueResponse struct {
	Paused     bool     `json:"paused"`
	NowPlaying YTSong   `json:"now_playing"`
	Queue      []YTSong `json:"queue,omitempty"`
}

var (
	now_playing YTSong                   // Song now playing
	paused      = false                  // Whether the current song is paused
	queue       []YTSong                 // The song queue.
	conns       map[*websocket.Conn]bool // Set of Websocket connections to notify with queue updates
	notif       chan struct{}            // Notifies player to play again after it empties
	mu          sync.Mutex               // mutex for good health

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	conns = make(map[*websocket.Conn]bool)
	notif = make(chan struct{}, 1)
	defer close(notif)

	go player_loop()

	http.HandleFunc("/queue", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
			return
		}

		mu.Lock()
		conns[conn] = true
		conn.WriteJSON(QueueResponse{paused, now_playing, queue})
		mu.Unlock()
		defer func() {
			mu.Lock()
			delete(conns, conn)
			mu.Unlock()
			conn.Close()
		}()

		for {
			var body PlayRequest
			if err := conn.ReadJSON(&body); err != nil {
				break
			}
			mu.Lock()
			switch body.Type {
			case "Toggle":
				if paused {
					err = speaker.Resume()
				} else {
					err = speaker.Suspend()
				}
				if err != nil {
					http.Error(w, "Failed to toggle speaker", http.StatusBadRequest)
					return
				}
				paused = !paused
			case "Add":
				song := YTSong{VideoID: body.VideoID, loaded: make(chan struct{})}
				go load(song)

				queue = append(queue, song)
				if len(notif) == 0 {
					notif <- struct{}{}
				}
			}
			for conn := range conns {
				conn.WriteJSON(QueueResponse{paused, now_playing, queue})
			}
			mu.Unlock()
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	if err := http.ListenAndServe(":8212", nil); err != nil {
		log.Fatalf("Failed to start backend: %v", err)
	}
}

func player_loop() {
	for {
		<-notif
		for {
			mu.Lock()
			if len(queue) == 0 {
				now_playing = YTSong{}
				for conn := range conns {
					conn.WriteJSON(QueueResponse{paused, now_playing, queue})
				}
				mu.Unlock()
				break
			}
			now_playing, queue = queue[0], queue[1:]

			for conn := range conns {
				conn.WriteJSON(QueueResponse{paused, now_playing, queue})
			}
			mu.Unlock()

			play(now_playing) // blocks
		}
	}
}
