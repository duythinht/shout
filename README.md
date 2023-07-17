## Radio Server for Wibu

### Prerequisite

* Golang 1.20++
* ffmpeg: `brew install ffmpeg`
* A Slack token that associcated with slack app that have permission to read channel history, listing/write bookmarks

### Easy to start

```
go run cmd/music-station/main.go
```

### Get title via websocket
```
const socket = new WebSocket("wss://radio.0x97a.com/now-playing");
socket.addEventListener("message", (event) => {
  console.log("Message from server ", event.data);
});
```

### Known issues

* So much, but I don't know :)