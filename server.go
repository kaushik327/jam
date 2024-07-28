package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type PlayRequest struct {
	VideoID string `json:"videoID"`
}

type QueueResponse struct {
	NowPlaying YTSong   `json:"now_playing"`
	Queue      []YTSong `json:"queue"`
}

var (
	now_playing YTSong                   // Song now playing
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
		conn.WriteJSON(QueueResponse{Queue: queue, NowPlaying: now_playing})
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

			song := YTSong{VideoID: body.VideoID, loaded: make(chan struct{})}
			go load(song)

			mu.Lock()
			queue = append(queue, song)
			if len(notif) == 0 {
				notif <- struct{}{}
			}
			for conn := range conns {
				conn.WriteJSON(QueueResponse{Queue: queue, NowPlaying: now_playing})
			}
			mu.Unlock()
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":8212", nil)
}

func player_loop() {
	for {
		<-notif
		for {
			mu.Lock()
			if len(queue) == 0 {
				now_playing = YTSong{}
				for conn := range conns {
					conn.WriteJSON(QueueResponse{Queue: queue, NowPlaying: now_playing})
				}
				mu.Unlock()
				break
			}
			now_playing, queue = queue[0], queue[1:]

			for conn := range conns {
				conn.WriteJSON(QueueResponse{Queue: queue, NowPlaying: now_playing})
			}
			mu.Unlock()

			play(now_playing) // blocks
		}
	}
}
