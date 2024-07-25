package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

func main() {

	// The song queue.
	var mtx sync.Mutex
	var queue []YTSong
	notif := make(chan struct{}, 1)

	go func() {
		var curr_song YTSong
		for range notif {
			mtx.Lock()
			for len(queue) != 0 {
				curr_song, queue = queue[0], queue[1:]
				mtx.Unlock()

				fmt.Printf("Playing song: %v\n", curr_song.VideoID)
				play(curr_song) // blocks
				mtx.Lock()
			}
			mtx.Unlock()
		}
	}()

	type PlayRequest struct {
		VideoID string `json:"videoID"`
	}
	http.HandleFunc("/queue", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodPost:
			var body PlayRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			song := YTSong{VideoID: body.VideoID, loaded: make(chan struct{})}
			go load(song)

			mtx.Lock()
			queue = append(queue, song)
			if len(notif) == 0 {
				notif <- struct{}{}
			}
			mtx.Unlock()

			w.WriteHeader(http.StatusOK)
			fmt.Printf("Added to queue: %s\n", song.VideoID)

		case http.MethodGet:
			mtx.Lock()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(queue)
			mtx.Unlock()

		default:
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		}
	})

	http.ListenAndServe(":8212", nil)
}
