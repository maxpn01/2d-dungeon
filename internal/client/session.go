// Package client keeps networking independent from raylib's drawing thread.
package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Update struct {
	Fruits     []protocol.Fruit
	ReceivedAt time.Time
	Local      protocol.Player
	PlayerID   string
	Players    []protocol.Player
	Tick       uint64
	Ready      bool
	Err        error
}

type sentInput struct {
	sequence  uint64
	direction protocol.Direction
}

type Session struct {
	predicted       protocol.Player
	predictionReady bool
	sequence        uint64
	pending         []sentInput
	updates         chan Update
	done            chan struct{}
	cancel          context.CancelFunc
	inputMu         sync.Mutex
	direction       protocol.Direction
	inputAt         time.Time
}

func Connect(url string) *Session {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Session{updates: make(chan Update, 1), done: make(chan struct{}), cancel: cancel}
	go func() {
		defer close(s.done)
		defer close(s.updates)
		defer cancel()
		if err := s.run(ctx, url); err != nil && ctx.Err() == nil {
			s.publish(Update{Err: err})
		}
	}()
	return s
}

func (s *Session) Updates() <-chan Update { return s.updates }
func (s *Session) Done() <-chan struct{}  { return s.done }
func (s *Session) Close()                 { s.cancel() }

func (s *Session) SetInput(direction protocol.Direction) {
	s.inputMu.Lock()
	s.direction, s.inputAt = direction, time.Now()
	s.inputMu.Unlock()
}

func (s *Session) PredictedPlayer() (protocol.Player, bool) {
	s.inputMu.Lock()
	defer s.inputMu.Unlock()
	return s.predicted, s.predictionReady
}

func (s *Session) input() protocol.Direction {
	s.inputMu.Lock()
	defer s.inputMu.Unlock()
	if time.Since(s.inputAt) > 150*time.Millisecond {
		return protocol.Direction{}
	}
	return s.direction
}

func (s *Session) publish(update Update) {
	select {
	case <-s.updates:
	default:
	}
	s.updates <- update
}

func (s *Session) run(ctx context.Context, url string) error {
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	conn, _, err := websocket.Dial(dialCtx, url, nil)
	cancel()
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	connectionCtx, cancelConnection := context.WithCancel(context.Background())
	finished, closerDone := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(closerDone)
		select {
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "leaving multiplayer")
			cancelConnection()
		case <-finished:
		}
	}()
	defer func() {
		close(finished)
		conn.CloseNow()
		cancelConnection()
		<-closerDone
	}()
	conn.SetReadLimit(1024 * 1024)

	read := func(readCtx context.Context) (protocol.Message, error) {
		timeoutCtx, stop := context.WithTimeout(readCtx, 5*time.Second)
		defer stop()
		var message protocol.Message
		err := wsjson.Read(timeoutCtx, conn, &message)
		return message, err
	}
	welcome, err := read(connectionCtx)
	if err != nil {
		return fmt.Errorf("receive welcome: %w", err)
	}
	if welcome.Type != protocol.TypeWelcome || welcome.PlayerID == "" {
		return errors.New("server did not send a valid welcome")
	}

	ioCtx, stop := context.WithCancel(connectionCtx)
	writerDone := make(chan struct{})
	writerError := make(chan error, 1)
	go func() {
		defer close(writerDone)
		defer stop()
		ticker := time.NewTicker(time.Second / 30)
		defer ticker.Stop()
		for {
			select {
			case <-ioCtx.Done():
				return
			case <-ticker.C:
				direction := s.input()
				s.inputMu.Lock()
				if len(s.pending) >= 128 {
					s.inputMu.Unlock()
					writerError <- errors.New("server stopped acknowledging input")
					return
				}
				s.sequence++
				sequence := s.sequence
				s.pending = append(s.pending, sentInput{sequence, direction})
				if s.predictionReady {
					s.predicted.X, s.predicted.Y = simulation.Move(s.predicted.X, s.predicted.Y, simulation.Direction{X: direction.X, Y: direction.Y}, simulation.StepSeconds)
				}
				s.inputMu.Unlock()
				writeCtx, finish := context.WithTimeout(ioCtx, 3*time.Second)
				err := wsjson.Write(writeCtx, conn, protocol.ClientMessage{Type: protocol.TypeInput, Direction: &direction, Sequence: sequence})
				finish()
				if err != nil {
					writerError <- err
					return
				}
			}
		}
	}()
	defer func() { stop(); <-writerDone }()

	var lastTick, lastAck uint64
	for {
		message, err := read(ioCtx)
		if err != nil {
			select {
			case err = <-writerError:
			default:
			}
			return fmt.Errorf("connection lost: %w", err)
		}
		if message.Type != protocol.TypeSnapshot || message.Players == nil || message.Fruits == nil {
			return errors.New("server sent an unexpected message")
		}
		if message.Tick < lastTick {
			return errors.New("server tick moved backwards")
		}
		if err := validatePlayers(*message.Players, welcome.PlayerID); err != nil {
			return err
		}
		if err := validateFruits(*message.Fruits); err != nil {
			return err
		}
		lastTick = message.Tick
		var local protocol.Player
		for _, player := range *message.Players {
			if player.ID == welcome.PlayerID {
				local = player
				break
			}
		}
		s.inputMu.Lock()
		if local.InputSeq > s.sequence || local.InputSeq < lastAck {
			s.inputMu.Unlock()
			return errors.New("server acknowledged an unknown input")
		}
		lastAck = local.InputSeq
		first := 0
		for first < len(s.pending) && s.pending[first].sequence <= local.InputSeq {
			first++
		}
		s.pending = s.pending[first:]
		for _, input := range s.pending {
			local.X, local.Y = simulation.Move(local.X, local.Y, simulation.Direction{X: input.direction.X, Y: input.direction.Y}, simulation.StepSeconds)
		}
		s.predicted, s.predictionReady = local, true
		s.inputMu.Unlock()
		s.publish(Update{PlayerID: welcome.PlayerID, Players: *message.Players, Tick: message.Tick, Ready: true, Local: local, ReceivedAt: time.Now(), Fruits: *message.Fruits})
	}
}

func validatePlayers(players []protocol.Player, localID string) error {
	ids := make(map[string]bool, len(players))
	for _, player := range players {
		if player.ID == "" || ids[player.ID] {
			return errors.New("server sent invalid player IDs")
		}
		ids[player.ID] = true
		for _, value := range []float64{player.X, player.Y, player.Size} {
			if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > math.MaxFloat32 {
				return errors.New("server sent an invalid player position or size")
			}
		}
		if player.Size <= 0 {
			return errors.New("server sent an invalid player size")
		}
	}
	if !ids[localID] {
		return errors.New("local player is missing from the world")
	}
	return nil
}

func validateFruits(fruits []protocol.Fruit) error {
	if len(fruits) > simulation.MaxFruits {
		return errors.New("server sent too many fruits")
	}
	ids := make(map[string]bool, len(fruits))
	for _, fruit := range fruits {
		if fruit.ID == "" || ids[fruit.ID] {
			return errors.New("server sent invalid fruit IDs")
		}
		ids[fruit.ID] = true
		for _, v := range []float64{fruit.X, fruit.Y, fruit.Size} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return errors.New("server sent invalid fruit coordinates")
			}
		}
		if fruit.Size <= 0 || fruit.X < simulation.MapMin || fruit.Y < simulation.MapMin ||
			fruit.X+fruit.Size > simulation.MapMax || fruit.Y+fruit.Size > simulation.MapMax {
			return errors.New("server sent fruit outside the map")
		}
	}
	return nil
}
