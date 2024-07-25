# jam

music server. you send it youtube ids (POST /queue) and it plays them right from the terminal where you ran the program.

run with `go run .` in the repository's root directory, then try:
```sh
curl --location 'localhost:8212/queue' \
--header 'Content-Type: application/json' \
--data '{"videoID": "FNt8xXCJplY"}'
# or whatever youtube ID you want. do it multiple times, even

curl --location 'localhost:8212/queue'
# gives you the queue of songs. note that the current song is not in the queue
```
