package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"raylib-2d-sandbox/internal/protocol"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (h *Hub) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(4096)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := &client{send: make(chan protocol.Message, 16), cancel: cancel}
	select {
	case h.join <- c:
	case <-h.done:
		return
	}
	defer func() {
		select {
		case h.leave <- c:
		case <-h.done:
		}
	}()

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case message := <-c.send:
				writeCtx, stop := context.WithTimeout(ctx, 5*time.Second)
				err := wsjson.Write(writeCtx, conn, message)
				stop()
				if err != nil {
					return
				}
			}
		}
	}()
	defer func() { cancel(); <-writerDone }()

	for {
		var request protocol.Message
		if err := wsjson.Read(ctx, conn, &request); err != nil {
			status := websocket.CloseStatus(err)
			if ctx.Err() == nil && status != websocket.StatusNormalClosure && status != websocket.StatusGoingAway {
				log.Printf("read websocket JSON: %v", err)
			}
			return
		}
		if request.Type != "" && request.Type != protocol.TypeEcho {
			conn.Close(websocket.StatusPolicyViolation, "unsupported message type")
			return
		}
		reply := protocol.Message{Type: protocol.TypeEcho, Message: "received: " + request.Message}
		select {
		case h.messages <- clientMessage{client: c, message: reply}:
		case <-ctx.Done():
			return
		case <-h.done:
			return
		}
	}
}
