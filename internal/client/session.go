package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Update struct {
	protocol.Snapshot
	ReceivedAt time.Time
	Local      protocol.Player
	PlayerID   string
	Ready      bool
	Err        error
}

type sentInput struct {
	sequence  uint64
	direction protocol.Direction
}

type Session struct {
	outgoingControls []protocol.ClientMessage
	safeZones        []simulation.SafeZone
	dead             bool
	predicted        protocol.Player
	predictionReady  bool
	sequence         uint64
	pending          []sentInput
	updates          chan Update
	done             chan struct{}
	cancel           context.CancelFunc
	inputMu          sync.Mutex
	direction        protocol.Direction
	inputAt          time.Time
	chatMu           sync.Mutex
	chat             []protocol.ChatEntry
	outgoingChat     []string
}

func Connect(endpoint string) *Session {
	return ConnectWithName(endpoint, "Player")
}

func ConnectWithName(endpoint, name string) *Session {
	return ConnectWithCharacter(endpoint, name, simulation.Square)
}

func ConnectWithCharacter(endpoint, name string, character simulation.Character) *Session {
	worldmap.SetMultiplayer(true)
	ctx, cancel := context.WithCancel(context.Background())
	s := &Session{updates: make(chan Update, 1), done: make(chan struct{}), cancel: cancel}
	go func() {
		defer close(s.done)
		defer close(s.updates)
		defer cancel()
		address, err := url.Parse(endpoint)
		if err != nil {
			s.publish(Update{Err: err})
			return
		}
		query := address.Query()
		query.Set("character", string(simulation.NormalizeCharacter(character)))
		query.Set("name", protocol.NormalizePlayerName(name))
		address.RawQuery = query.Encode()
		if err := s.run(ctx, address.String()); err != nil && ctx.Err() == nil {
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
	if s.direction.Shoot || s.direction.Dash || s.direction.Interact {
		direction.AimX, direction.AimY = s.direction.AimX, s.direction.AimY
	}
	direction.Shoot = direction.Shoot || s.direction.Shoot
	direction.Dash = direction.Dash || s.direction.Dash
	direction.Interact = direction.Interact || s.direction.Interact
	s.direction, s.inputAt = direction, time.Now()
	s.inputMu.Unlock()
}

func (s *Session) sendControl(message protocol.ClientMessage) {
	s.inputMu.Lock()
	defer s.inputMu.Unlock()
	if len(s.outgoingControls) < 8 {
		s.outgoingControls = append(s.outgoingControls, message)
	}
}

func (s *Session) ChoosePerk(statueID string, cycle uint64, card int) {
	s.sendControl(protocol.ClientMessage{Type: protocol.TypePerkChoice, Perk: &protocol.PerkSelection{StatueID: statueID, Cycle: cycle, Card: card}})
}

func (s *Session) CloseOffer() { s.sendControl(protocol.ClientMessage{Type: protocol.TypeCloseOffer}) }
func (s *Session) Respawn()    { s.sendControl(protocol.ClientMessage{Type: protocol.TypeRespawn}) }

func (s *Session) input() protocol.Direction {
	s.inputMu.Lock()
	defer s.inputMu.Unlock()
	if time.Since(s.inputAt) > 150*time.Millisecond {
		s.direction = protocol.Direction{}
		return protocol.Direction{}
	}
	direction := s.direction
	s.direction.Shoot = false
	s.direction.Dash = false
	s.direction.Interact = false
	return direction
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
	// Snapshot size grows with the number of split bodies.
	conn.SetReadLimit(-1)

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
				s.inputMu.Lock()
				controls := s.outgoingControls
				s.outgoingControls = nil
				s.inputMu.Unlock()
				for _, message := range controls {
					writeCtx, finish := context.WithTimeout(ioCtx, 3*time.Second)
					err := wsjson.Write(writeCtx, conn, message)
					finish()
					if err != nil {
						writerError <- err
						return
					}
				}
				failed := false
				for _, text := range s.takeOutgoingChat() {
					writeCtx, finish := context.WithTimeout(ioCtx, 3*time.Second)
					err := wsjson.Write(writeCtx, conn, protocol.ClientMessage{Type: protocol.TypeChat, Message: text})
					finish()
					if err != nil {
						writerError <- err
						failed = true
						break
					}
				}
				if failed {
					return
				}
				direction := s.input()
				s.inputMu.Lock()
				if s.dead {
					s.inputMu.Unlock()
					continue
				}
				if len(s.pending) >= 128 {
					s.inputMu.Unlock()
					writerError <- errors.New("server stopped acknowledging input")
					return
				}
				s.sequence++
				sequence := s.sequence
				s.pending = append(s.pending, sentInput{sequence, direction})
				if s.predictionReady {
					predictPlayer(&s.predicted, direction, s.safeZones)
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
		if message.Type == protocol.TypeChat {
			if text := protocol.NormalizeChatMessage(message.Message); text != "" {
				s.appendChat(protocol.ChatEntry{
					At: time.Now(), Name: protocol.NormalizePlayerName(message.Name),
					Text: text, Mine: message.PlayerID == welcome.PlayerID,
				})
			}
			continue
		}
		if message.Type != protocol.TypeSnapshot || message.Players == nil || message.Fruits == nil || message.AIs == nil {
			return errors.New("server sent an unexpected message")
		}
		if message.Tick < lastTick {
			return errors.New("server tick moved backwards")
		}
		if err := protocol.ValidateSnapshot(message.Snapshot(), welcome.PlayerID); err != nil {
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
		s.safeZones = message.SafeZones
		s.dead = local.Dead
		if s.dead {
			s.pending = nil
		}
		for _, input := range s.pending {
			predictPlayer(&local, input.direction, message.SafeZones)
		}
		s.predicted, s.predictionReady = local, !s.dead
		s.predicted.Split = local.Split.Clone()
		s.inputMu.Unlock()
		s.publish(Update{Snapshot: message.Snapshot(), PlayerID: welcome.PlayerID, Ready: true, Local: local, ReceivedAt: time.Now()})
	}
}
