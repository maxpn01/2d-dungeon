package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"sync"
	"time"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/terrain"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Update struct {
	AIs        []protocol.AI
	Fruits     []protocol.Fruit
	ReceivedAt time.Time
	Local      protocol.Player
	PlayerID   string
	Players    []protocol.Player
	Tick       uint64
	Ready      bool
	Err        error
}

type ChatEntry struct {
	At   time.Time
	Name string
	Text string
	Mine bool
}

const maxChatEntries = 50

type sentInput struct {
	sequence  uint64
	direction protocol.Direction
}

type Session struct {
	dead            bool
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
	chatMu          sync.Mutex
	chat            []ChatEntry
	outgoingChat    []string
}

func Connect(endpoint string) *Session {
	return ConnectWithName(endpoint, "Player")
}

func ConnectWithName(endpoint, name string) *Session {
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
	direction.Dash = direction.Dash || s.direction.Dash
	s.direction, s.inputAt = direction, time.Now()
	s.inputMu.Unlock()
}

func (s *Session) SendChat(text string) {
	text = protocol.NormalizeChatMessage(text)
	if text == "" {
		return
	}
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	if len(s.outgoingChat) < 8 {
		s.outgoingChat = append(s.outgoingChat, text)
	}
}

func (s *Session) Chat() []ChatEntry {
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	return append([]ChatEntry(nil), s.chat...)
}

func (s *Session) RestoreChat(entries []ChatEntry) {
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	history := append([]ChatEntry(nil), entries...)
	s.chat = append(history, s.chat...)
	if len(s.chat) > maxChatEntries {
		s.chat = append([]ChatEntry(nil), s.chat[len(s.chat)-maxChatEntries:]...)
	}
}

func (s *Session) appendChat(entry ChatEntry) {
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	s.chat = append(s.chat, entry)
	if len(s.chat) > maxChatEntries {
		s.chat = append([]ChatEntry(nil), s.chat[len(s.chat)-maxChatEntries:]...)
	}
}

func (s *Session) takeOutgoingChat() []string {
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	pending := s.outgoingChat
	s.outgoingChat = nil
	return pending
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
		s.direction = protocol.Direction{}
		return protocol.Direction{}
	}
	direction := s.direction
	s.direction.Dash = false
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
					predictPlayer(&s.predicted, direction)
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
				s.appendChat(ChatEntry{
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
		if err := validatePlayers(*message.Players, welcome.PlayerID); err != nil {
			return err
		}
		if err := validateFruits(*message.Fruits); err != nil {
			return err
		}
		if err := validateAIs(*message.AIs); err != nil {
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
		s.dead = local.Dead
		if s.dead {
			s.pending = nil
		}
		for _, input := range s.pending {
			predictPlayer(&local, input.direction)
		}
		s.predicted, s.predictionReady = local, !s.dead
		s.inputMu.Unlock()
		s.publish(Update{PlayerID: welcome.PlayerID, Players: *message.Players, Tick: message.Tick, Ready: true, Local: local, ReceivedAt: time.Now(), Fruits: *message.Fruits, AIs: *message.AIs})
	}
}

func validatePlayers(players []protocol.Player, localID string) error {
	ids := make(map[string]bool, len(players))
	for _, player := range players {
		if player.ID == "" || ids[player.ID] {
			return errors.New("server sent invalid player IDs")
		}
		ids[player.ID] = true
		if !validStats(player.Stats, simulation.PlayerSpeed) {
			return errors.New("server sent invalid player stats")
		}
		for _, value := range []float64{player.X, player.Y, player.Size} {
			if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > math.MaxFloat32 {
				return errors.New("server sent an invalid player position or size")
			}
		}
		if !validCombat(player.Combat) {
			return errors.New("server sent invalid combat state")
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
	if len(fruits) > simulation.MaxFruits+simulation.MaxMassDrops {
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
		if fruit.Source != "" && (math.IsNaN(fruit.HP) || fruit.HP <= 0 || fruit.HP > 1) {
			return errors.New("server sent invalid mass pickup")
		}
		if fruit.Size <= 0 || !terrain.Walkable(fruit.X, fruit.Y, fruit.Size, fruit.Size) {
			return errors.New("server sent fruit outside the map")
		}
	}
	return nil
}

func validStats(stats protocol.PlayerStats, baseSpeed float64) bool {
	if stats.Level < 1 || stats.Level > simulation.MaxLevel || math.IsNaN(stats.XP) || math.IsInf(stats.XP, 0) ||
		stats.XP < 0 || stats.HP < 0 || stats.HP > simulation.MaxHPForLevel(stats.Level) || math.IsNaN(stats.HP) ||
		stats.Speed != baseSpeed {
		return false
	}
	if stats.Level == simulation.MaxLevel {
		return stats.XP == 0
	}
	return stats.XP < simulation.XPForNextLevel(stats.Level)
}

func validateAIs(ais []protocol.AI) error {
	if len(ais) > simulation.InitialAIs {
		return errors.New("server sent too many AIs")
	}
	ids := make(map[string]bool, len(ais))
	for _, ai := range ais {
		if ai.ID == "" || ids[ai.ID] {
			return errors.New("server sent invalid AI IDs")
		}
		ids[ai.ID] = true
		if !validStats(ai.Stats, simulation.AISpeed) {
			return errors.New("server sent invalid AI stats")
		}
		for _, v := range []float64{ai.X, ai.Y, ai.Size} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return errors.New("server sent invalid AI coordinates")
			}
		}
		if ai.Size <= 0 || !terrain.Walkable(ai.X, ai.Y, ai.Size, ai.Size) {
			return errors.New("server sent an AI outside the map")
		}
	}
	return nil
}

func predictPlayer(player *protocol.Player, input protocol.Direction) {
	p := simulation.Player{X: player.X, Y: player.Y, Dead: player.Dead, Combat: player.Combat, Stats: simulation.Stats{Level: player.Stats.Level, XP: player.Stats.XP, HP: player.Stats.HP, Speed: player.Stats.Speed}}
	simulation.AdvancePlayer(&p, simulation.Direction{X: input.X, Y: input.Y, Dash: input.Dash}, simulation.StepSeconds)
	player.X, player.Y, player.Combat, player.Stats.HP = p.X, p.Y, p.Combat, p.Stats.HP
}

func validCombat(c simulation.CombatState) bool {
	if !c.Facing.Valid() || !c.Direction.Valid() || c.Facing.Dash || c.Direction.Dash || math.IsNaN(c.Remaining) || math.IsInf(c.Remaining, 0) || c.Remaining < 0 {
		return false
	}
	switch c.Phase {
	case simulation.DashIdle:
		return c.Remaining == 0
	case simulation.DashWinding:
		return c.Remaining <= simulation.DashWindup
	case simulation.DashRushing:
		return c.Remaining <= simulation.DashBurst
	case simulation.DashRecovering:
		return c.Remaining <= simulation.DashRecovery
	case simulation.DashBumped:
		return c.Remaining <= simulation.BumpStun
	default:
		return false
	}
}
