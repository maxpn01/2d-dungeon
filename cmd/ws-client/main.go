package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"raylib-2d-sandbox/internal/protocol"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws://localhost:8080/ws", nil)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.CloseNow()

	for _, text := range []string{"hello", "still connected", "goodbye"} {
		if err := wsjson.Write(ctx, conn, protocol.Message{Message: text}); err != nil {
			return fmt.Errorf("send: %w", err)
		}
		var reply protocol.Message
		if err := wsjson.Read(ctx, conn, &reply); err != nil {
			return fmt.Errorf("receive: %w", err)
		}
		fmt.Println(reply.Message)
	}

	return conn.Close(websocket.StatusNormalClosure, "demo finished")
}
