package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"raylib-2d-sandbox/internal/protocol"
)

func main() {
	url := flag.String("url", "ws://localhost:8080/ws", "WebSocket server URL")
	flag.Parse()
	if err := run(*url); err != nil {
		log.Fatal(err)
	}
}

func run(url string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	conn, _, err := websocket.Dial(dialCtx, url, nil)
	cancel() // Only connecting has a five-second deadline.
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.CloseNow()

	// Close gracefully on Ctrl+C while the read loop is still processing frames.
	finished := make(chan struct{})
	defer close(finished)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "client exiting")
		case <-finished:
		}
	}()
	fmt.Println("Connected. Open another client to see updates. Ctrl+C to leave.")
	for {
		var message protocol.Message
		// Broadcasts arrive independently; we don't send a request before each read.
		if err := wsjson.Read(context.Background(), conn, &message); err != nil {
			if ctx.Err() != nil || websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				return nil
			}
			return fmt.Errorf("receive: %w", err)
		}
		switch message.Type {
		case protocol.TypeWelcome:
			fmt.Printf("Your player ID: %s\n", message.PlayerID)
		case protocol.TypeSnapshot:
			if message.Players != nil {
				fmt.Printf("Connected players: %v\n", *message.Players)
			}
		case protocol.TypeEcho:
			fmt.Println(message.Message)
		}
	}
}
