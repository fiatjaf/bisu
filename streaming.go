package main

import (
	"net/http"

	"github.com/fasthttp/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var streamingConns = make([]*websocket.Conn, 0, 10)

func streamingHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn().Err(err).Msg("failed to upgrade websocket connection on streaming handler")
		return
	}

	streamingConns = append(streamingConns, conn)
	topic := r.URL.Query().Get("stream")
	switch topic {
	case "public":
	case "public:local":
	case "hashtag":
		_ = r.URL.Query().Get("tag")
	case "hashtag:local":
		_ = r.URL.Query().Get("tag")
	case "user":
	case "user:notification":
	case "list":
	case "direct":
	}
}
