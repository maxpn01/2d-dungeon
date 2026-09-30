package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"sync"
	"time"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/terrain"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Update struct {
	SafeZones  []simulation.SafeZone
	Boss       simulation.Boss
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

const maxChatEntries = parameters.MaxChatEntries

type sentInput struct {
	sequence  uint64
	direction protocol.Direction
}

type Session struct {
	safeZones       []simulation.SafeZone
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
	return ConnectWithCharacter(endpoint, name, simulation.Square)
}

func ConnectWithCharacter(endpoint, name string, character simulation.Character) *Session {
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
	if s.direction.Dash || s.direction.Eject {
		direction.AimX, direction.AimY = s.direction.AimX, s.direction.AimY
	}
	direction.Dash = direction.Dash || s.direction.Dash
	direction.Eject = direction.Eject || s.direction.Eject
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
	p := s.predicted
	p.Split = p.Split.Clone()
	return p, s.predictionReady
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
	s.direction.Eject = false
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
		if err := validateBoss(message.Boss); err != nil {
			return err
		}
		if err := validateAIs(*message.AIs); err != nil {
			return err
		}
		if len(message.SafeZones) > simulation.SafeZoneCount {
			return errors.New("server sent too many refuges")
		}
		for _, zone := range message.SafeZones {
			for _, v := range []float64{zone.X, zone.Y, zone.Radius} {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					return errors.New("server sent invalid refuge")
				}
			}
			if zone.Radius <= 0 || zone.Radius > 100 || zone.X-zone.Radius < simulation.MapMin || zone.Y-zone.Radius < simulation.MapMin || zone.X+zone.Radius > simulation.MapMax || zone.Y+zone.Radius > simulation.MapMax {
				return errors.New("server sent invalid refuge bounds")
			}
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
		s.publish(Update{SafeZones: message.SafeZones, Boss: message.Boss, PlayerID: welcome.PlayerID, Players: *message.Players, Tick: message.Tick, Ready: true, Local: local, ReceivedAt: time.Now(), Fruits: *message.Fruits, AIs: *message.AIs})
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
		if player.Character != "" && simulation.NormalizeCharacter(player.Character) != player.Character {
			return errors.New("server sent invalid character")
		}
		if player.Stats.HP+player.Split.Mass() > simulation.MaxHPForLevel(player.Stats.Level)+1e-9 {
			return errors.New("server sent invalid total mass")
		}
		for _, cell := range player.Split {
			if !cell.Active {
				continue
			}
			if player.Character != simulation.Circle || !validStats(protocol.PlayerStats{Level: cell.Stats.Level, XP: cell.Stats.XP, HP: cell.Stats.HP, Speed: cell.Stats.Speed}, simulation.PlayerSpeed) || cell.Stats.HP <= 0 || cell.Stats.Level != player.Stats.Level || cell.Stats.XP != player.Stats.XP {
				return errors.New("server sent invalid split stats")
			}
			for _, v := range []float64{cell.X, cell.Y, cell.Remaining, cell.Launch} {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					return errors.New("server sent invalid split cell")
				}
			}
			if cell.Remaining < 0 || cell.Remaining > simulation.SplitDuration || cell.Launch < 0 || cell.Launch > simulation.SplitLaunchDuration || !cell.Direction.Valid() || !terrain.Walkable(cell.X, cell.Y, simulation.PlayerSize, simulation.PlayerSize) {
				return errors.New("server sent invalid split position or timer")
			}
		}
		if math.IsNaN(player.EjectCooldown) || math.IsInf(player.EjectCooldown, 0) || player.EjectCooldown < 0 || player.EjectCooldown > simulation.EjectInterval {
			return errors.New("server sent invalid eject cooldown")
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
		if math.IsNaN(fruit.MP) || math.IsInf(fruit.MP, 0) || fruit.MP < 0 || fruit.MP > simulation.MaxMP || fruit.MP > 0 && (fruit.HP != 0 || fruit.Source != "") {
			return errors.New("server sent invalid mana pickup")
		}
		for _, v := range []float64{fruit.X, fruit.Y, fruit.Size} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return errors.New("server sent invalid fruit coordinates")
			}
		}
		maxDropHP := 1.0
		if fruit.Source == "boss" {
			maxDropHP = simulation.BossPartHP
		}
		if fruit.Source != "" && (math.IsNaN(fruit.HP) || fruit.HP <= 0 || fruit.HP > maxDropHP) {
			return errors.New("server sent invalid mass pickup")
		}
		if fruit.Size <= 0 || !terrain.Walkable(fruit.X, fruit.Y, fruit.Size, fruit.Size) {
			return errors.New("server sent fruit outside the map")
		}
	}
	return nil
}

func validStats(stats protocol.PlayerStats, baseSpeed float64) bool {
	if math.IsNaN(stats.MP) || math.IsInf(stats.MP, 0) || stats.MP < 0 || stats.MP > simulation.MaxMP {
		return false
	}
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
		if simulation.NormalizeCharacter(ai.Character) != ai.Character {
			return errors.New("server sent an invalid AI character")
		}
		if !validCombat(ai.Combat) {
			return errors.New("server sent invalid AI combat state")
		}
		if !validStats(ai.Stats, simulation.AISpeed) {
			return errors.New("server sent invalid AI stats")
		}
		if ai.Stats.HP+ai.Split.Mass() > simulation.MaxHPForLevel(ai.Stats.Level)+1e-9 {
			return errors.New("server sent invalid AI total mass")
		}
		for _, cell := range ai.Split {
			if !cell.Active {
				continue
			}
			if ai.Character != simulation.Circle || !validStats(protocol.PlayerStats{Level: cell.Stats.Level, XP: cell.Stats.XP, HP: cell.Stats.HP, Speed: cell.Stats.Speed}, simulation.AISpeed) || cell.Stats.HP <= 0 || cell.Stats.Level != ai.Stats.Level || cell.Stats.XP != ai.Stats.XP {
				return errors.New("server sent invalid AI split stats")
			}
			for _, v := range []float64{cell.X, cell.Y, cell.Remaining, cell.Launch} {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					return errors.New("server sent invalid AI split cell")
				}
			}
			if cell.Remaining < 0 || cell.Remaining > simulation.SplitDuration || cell.Launch < 0 || cell.Launch > simulation.SplitLaunchDuration || !cell.Direction.Valid() || !terrain.Walkable(cell.X, cell.Y, simulation.AISize, simulation.AISize) {
				return errors.New("server sent invalid AI split position or timer")
			}
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

func predictPlayer(player *protocol.Player, input protocol.Direction, zones []simulation.SafeZone) {
	p := simulation.Player{EjectCooldown: player.EjectCooldown, Character: player.Character, Split: player.Split.Clone(), GatesClosed: player.GatesClosed, InArena: player.InArena, X: player.X, Y: player.Y, Dead: player.Dead, Combat: player.Combat, Stats: simulation.Stats{Level: player.Stats.Level, XP: player.Stats.XP, HP: player.Stats.HP, MP: player.Stats.MP, Speed: player.Stats.Speed}}
	d := simulation.Direction{X: input.X, Y: input.Y, Boost: input.Boost, Dash: input.Dash, Eject: input.Eject, AimX: input.AimX, AimY: input.AimY}
	simulation.EjectFood(&p, d, simulation.StepSeconds)
	simulation.AdvancePlayer(&p, d, simulation.StepSeconds, zones)
	player.X, player.Y, player.Combat, player.Stats.HP = p.X, p.Y, p.Combat, p.Stats.HP
	player.Stats.MP = p.Stats.MP
	player.Split = p.Split
	player.EjectCooldown = p.EjectCooldown
}

func validCombat(c simulation.CombatState) bool {
	for _, v := range []float64{c.RecoilX, c.RecoilY} {
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > simulation.BossRecoilDistance+.001 {
			return false
		}
	}
	if math.IsNaN(c.HitGrace) || math.IsInf(c.HitGrace, 0) || c.HitGrace < 0 || c.HitGrace > simulation.BossHitGrace {
		return false
	}
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
	case simulation.DashRecoiling:
		return c.Remaining <= simulation.BossRecoilDuration
	case simulation.DashBumped:
		return c.Remaining <= simulation.BumpStun
	default:
		return false
	}
}

func validateBoss(b simulation.Boss) error {
	if b.Phase == "" {
		return nil
	}
	switch b.Phase {
	case "waiting", "gathering", "fighting", "defeated":
	default:
		return errors.New("server sent invalid boss phase")
	}
	switch b.Behavior {
	case "", "hunting", "winding", "rushing", "recovering", "feeding", "patrolling", "fleeing":
	default:
		return errors.New("server sent invalid boss behavior")
	}
	if b.XP < 0 || (b.Level == simulation.MaxLevel && b.XP != 0) || (b.Level < simulation.MaxLevel && b.XP >= simulation.XPForNextLevel(b.Level)) {
		return errors.New("server sent invalid boss XP")
	}
	if b.Chain < 0 || b.Chain > 3 {
		return errors.New("server sent invalid boss combo")
	}
	if len(b.Tail) > simulation.BossMaxTail || b.Level < 1 || b.Level > simulation.MaxLevel || b.HP < 0 || b.HP > b.MaxHP || b.MaxHP > simulation.BossMaxHP || b.MaxHP <= 0 || b.Countdown < 0 || b.Countdown > simulation.BossCountdown || b.SpawnDuration < 0 || b.SpawnTicks < 0 || b.SpawnTicks > b.SpawnDuration || b.Defeats < 0 {
		return errors.New("server sent invalid boss stats")
	}
	for _, v := range []float64{b.X, b.Y, b.HP, b.MaxHP, b.Countdown, b.XP, b.AttackTarget.X, b.AttackTarget.Y} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errors.New("server sent invalid boss values")
		}
	}
	if b.X < simulation.MapMin || b.Y < simulation.MapMin || b.X+simulation.BossSize > simulation.MapMax || b.Y+simulation.BossSize > simulation.MapMax {
		return errors.New("server sent boss outside surface")
	}
	for _, segment := range b.Tail {
		if math.IsNaN(segment.X) || math.IsNaN(segment.Y) || math.IsInf(segment.X, 0) || math.IsInf(segment.Y, 0) || segment.X < simulation.MapMin || segment.Y < simulation.MapMin || segment.X+simulation.BossSize > simulation.MapMax || segment.Y+simulation.BossSize > simulation.MapMax {
			return errors.New("server sent invalid boss tail")
		}
	}
	return nil
}
