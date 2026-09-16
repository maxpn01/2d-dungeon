package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"raylib-2d-sandbox/internal/protocol"
)

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("accept websocket: %v", err)
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(4096)

	ctx := context.Background()
	log.Print("websocket client connected")
	defer log.Print("websocket client disconnected")

	for {
		var request protocol.Message

		if err := wsjson.Read(ctx, conn, &request); err != nil {
			status := websocket.CloseStatus(err)
			if status != websocket.StatusNormalClosure && status != websocket.StatusGoingAway {
				log.Printf("read websocket JSON: %v", err)
			}
			return
		}

		response := protocol.Message{Message: "received: " + request.Message}
		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := wsjson.Write(writeCtx, conn, response)

		cancel()

		if err != nil {
			log.Printf("write websocket JSON: %v", err)
			return
		}
	}
}
