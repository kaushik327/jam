# jam

Music server that reads youtube ids and plays them right from the terminal where you ran the program.

Run with `go run .` in the repository's root directory, then send jsons through websocket

```
send a video id
<- {"type": "Add", "videoID": "FNt8xXCJplY"}

you're sent the whole queue whenever it changes
-> {
    "paused": false,
    "now_playing": {"videoID": "FNt8xXCJplY"},
    "queue": [
        {"videoID": "FNt8xXCJplY"},
        {"videoID": "FNt8xXCJplY"},
        {"videoID": "FNt8xXCJplY"}
    ]
}

pause the song
<- {"type": "Toggle"}

-> {
    "paused": true,
    "now_playing": {"videoID": "FNt8xXCJplY"},
    "queue": [
        {"videoID": "FNt8xXCJplY"},
        {"videoID": "FNt8xXCJplY"},
        {"videoID": "FNt8xXCJplY"}
    ]
}
```