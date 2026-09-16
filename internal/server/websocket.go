package server

import (
	"context"
	"log"
	"math"
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
	c := &client{name: protocol.NormalizePlayerName(r.URL.Query().Get("name")), send: make(chan protocol.Message, 16), cancel: cancel}
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

	const messagesPerSecond = 120.0
	tokens, lastMessage := messagesPerSecond, time.Now()
	for {
		kind, data, err := conn.Read(ctx)
		if err != nil {
			status := websocket.CloseStatus(err)
			if ctx.Err() == nil && status != websocket.StatusNormalClosure && status != websocket.StatusGoingAway {
				log.Printf("read websocket JSON: %v", err)
			}
			return
		}
		now := time.Now()
		tokens = math.Min(messagesPerSecond, tokens+now.Sub(lastMessage).Seconds()*messagesPerSecond)
		lastMessage = now
		if tokens < 1 {
			conn.Close(websocket.StatusPolicyViolation, "message rate exceeded")
			return
		}
		tokens--
		if kind != websocket.MessageText {
			conn.Close(websocket.StatusUnsupportedData, "expected JSON text")
			return
		}
		request, err := protocol.DecodeClientMessage(data)
		if err != nil {
			conn.Close(websocket.StatusPolicyViolation, "invalid client message")
			return
		}
		select {
		case h.messages <- clientMessage{client: c, message: request}:
		case <-ctx.Done():
			return
		case <-h.done:
			return
		}
	}
}
