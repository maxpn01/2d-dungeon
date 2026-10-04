package server

import (
	"context"
	"log"
	"math"
	"net/http"
	"slices"
	"time"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (h *Hub) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{CompressionMode: websocket.CompressionContextTakeover})
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(4096)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := &client{joined: make(chan bool, 1), character: simulation.NormalizeCharacter(simulation.Character(r.URL.Query().Get("character"))), name: protocol.NormalizePlayerName(r.URL.Query().Get("name")), send: make(chan protocol.Message, 16), snapshots: make(chan protocol.Message, 1), cancel: cancel}
	select {
	case h.join <- c:
	case <-h.done:
		return
	}
	select {
	case accepted := <-c.joined:
		if !accepted {
			conn.Close(websocket.StatusPolicyViolation, "dungeon is full (2 players)")
			return
		}
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
		var sentFruits []protocol.Fruit
		var sentRound uint64
		write := func(message protocol.Message) error {
			fruits := message.Fruits
			if message.Type == protocol.TypeSnapshot && message.RoundID == sentRound && fruits != nil && slices.Equal(*fruits, sentFruits) {
				message.Fruits = nil
			}
			writeCtx, stop := context.WithTimeout(ctx, 5*time.Second)
			defer stop()
			err := wsjson.Write(writeCtx, conn, message)
			if err == nil && message.Type == protocol.TypeSnapshot && fruits != nil {
				sentFruits, sentRound = *fruits, message.RoundID
			}
			return err
		}
		for {
			select {
			case message := <-c.send:
				if err := write(message); err != nil {
					return
				}
				continue
			default:
			}
			select {
			case <-ctx.Done():
				return
			case message := <-c.send:
				if err := write(message); err != nil {
					return
				}
			case message := <-c.snapshots:
				select {
				case control := <-c.send:
					if err := write(control); err != nil {
						return
					}
				default:
				}
				if err := write(message); err != nil {
					return
				}
			}
		}
	}()
	defer func() { cancel(); <-writerDone }()

	const messagesPerSecond = parameters.MessagesPerSecond
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
