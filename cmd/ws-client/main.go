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

	"raylib-2d-sandbox/internal/protocol"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func main() {
	url := flag.String("url", "ws://localhost:8080/ws", "WebSocket server URL")
	move := flag.String("move", "none", "scripted direction: none, right, left, up, down, diagonal")
	duration := flag.Duration("duration", time.Second, "how long to send movement before stopping")
	flag.Parse()
	direction, err := parseDirection(*move)
	if err != nil {
		log.Fatal(err)
	}
	if *duration <= 0 {
		log.Fatal("duration must be positive")
	}
	if err := run(*url, direction, *duration); err != nil {
		log.Fatal(err)
	}
}

func parseDirection(name string) (*protocol.Direction, error) {
	switch name {
	case "none":
		return nil, nil
	case "right":
		return &protocol.Direction{X: 1}, nil
	case "left":
		return &protocol.Direction{X: -1}, nil
	case "up":
		return &protocol.Direction{Y: -1}, nil
	case "down":
		return &protocol.Direction{Y: 1}, nil
	case "diagonal":
		return &protocol.Direction{X: 1, Y: 1}, nil
	default:
		return nil, fmt.Errorf("unknown direction %q", name)
	}
}

func run(url string, direction *protocol.Direction, duration time.Duration) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	conn, _, err := websocket.Dial(dialCtx, url, nil)
	cancel()
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1024 * 1024)

	finished := make(chan struct{})
	defer close(finished)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "client exiting")
		case <-finished:
		}
	}()

	sendCtx, stopSending := context.WithCancel(ctx)
	ready := make(chan struct{})
	senderDone := make(chan struct{})
	senderError := make(chan error, 1)
	go func() {
		defer close(senderDone)
		select {
		case <-ready:
		case <-sendCtx.Done():
			return
		}
		if direction == nil {
			return
		}
		if err := sendMovement(sendCtx, conn, *direction, duration); err != nil && sendCtx.Err() == nil {
			senderError <- err
			conn.CloseNow()
		}
	}()
	defer func() { stopSending(); <-senderDone }()

	fmt.Println("Connected. Ctrl+C to leave. Snapshots show server-owned positions.")
	welcomed := false
	var lastPrinted time.Time
	for {
		var message protocol.Message
		if err := wsjson.Read(context.Background(), conn, &message); err != nil {
			select {
			case sendErr := <-senderError:
				return fmt.Errorf("send movement: %w", sendErr)
			default:
			}
			if ctx.Err() != nil || websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				return nil
			}
			return fmt.Errorf("receive: %w", err)
		}
		switch message.Type {
		case protocol.TypeWelcome:
			fmt.Printf("Your player ID: %s\n", message.PlayerID)
			if !welcomed {
				close(ready)
				welcomed = true
			}
		case protocol.TypeSnapshot:
			if message.Players != nil && time.Since(lastPrinted) >= 500*time.Millisecond {
				fmt.Printf("tick=%d", message.Tick)
				if message.Fruits != nil {
					fmt.Printf(" fruits=%d", len(*message.Fruits))
				}
				for _, player := range *message.Players {
					fmt.Printf(" %s=(%.1f,%.1f) lvl=%d xp=%.2f", player.ID, player.X, player.Y, player.Stats.Level, player.Stats.XP)
				}
				fmt.Println()
				lastPrinted = time.Now()
			}
		case protocol.TypeEcho:
			fmt.Println(message.Message)
		}
	}
}

func sendMovement(ctx context.Context, conn *websocket.Conn, direction protocol.Direction, duration time.Duration) error {
	write := func(d protocol.Direction) error {
		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return wsjson.Write(writeCtx, conn, protocol.ClientMessage{Type: protocol.TypeInput, Direction: &d})
	}
	if err := write(direction); err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()
	deadline := time.Now().Add(duration)
	timer := time.NewTimer(duration)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := write(protocol.Direction{}); err != nil {
				return err
			}
			fmt.Println("Movement stopped; continuing to receive snapshots.")
			return nil
		case <-ticker.C:
			if time.Now().Before(deadline) {
				if err := write(direction); err != nil {
					return err
				}
			}
		}
	}
}
