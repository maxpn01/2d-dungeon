package server

import (
	"context"
	"fmt"
	"math"
	"time"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"
)

type client struct {
	joined       chan bool
	ready        bool
	briefingDone bool
	character    simulation.Character
	name         string
	id           string
	send         chan protocol.Message
	snapshots    chan protocol.Message
	cancel       context.CancelFunc
	lastChat     time.Time
}

const chatCooldown = parameters.ChatCooldown

type clientMessage struct {
	client  *client
	message protocol.ClientMessage
}

type Hub struct {
	join     chan *client
	leave    chan *client
	messages chan clientMessage
	done     chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		join: make(chan *client), leave: make(chan *client),
		messages: make(chan clientMessage), done: make(chan struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	worldmap.SetMultiplayer(true)
	defer close(h.done)
	clients := make(map[*client]struct{})
	names := make(map[string]string)
	world := simulation.NewWorld()
	matchStarted := false
	briefingActive := false
	roundID := uint64(1)
	var startAt time.Time
	ticker := time.NewTicker(time.Second / simulation.TickRate)
	defer ticker.Stop()
	nextID := 0
	remove := func(c *client) {
		delete(clients, c)
		delete(names, c.id)
		world.Leave(c.id)
		if !matchStarted || world.Finished() {
			briefingActive = false
			startAt = time.Time{}
			if world.Finished() && len(clients) > 0 {
				matchStarted = true
			}
			for remaining := range clients {
				remaining.ready = false
				if !world.Finished() {
					remaining.briefingDone = false
				}
			}
		}
		if len(clients) == 0 {
			matchStarted = false
			startAt = time.Time{}
			world = simulation.NewWorld()
			roundID++
		}
		c.cancel()
	}
	defer func() {
		for c := range clients {
			remove(c)
		}
	}()

	var lastBroadcast time.Time
	broadcast := func() {
		lastBroadcast = time.Now()
		if len(clients) == 0 {
			return
		}
		snapshot := protocol.SnapshotFromSimulation(world.Snapshot(), names)
		snapshot.RoundID = roundID
		snapshot.MatchStarted = matchStarted
		snapshot.BriefingActive = briefingActive
		for i := range snapshot.Players {
			for c := range clients {
				if c.id == snapshot.Players[i].ID {
					snapshot.Players[i].LobbyReady = c.ready
				}
			}
		}
		if !startAt.IsZero() && !matchStarted {
			snapshot.StartCountdown = max(1, int(math.Ceil(time.Until(startAt).Seconds())))
		}
		for c := range clients {
			personalized := snapshot
			personalized.Statues = world.StatuesForPlayer(c.id)
			message := personalized.Message()
			select {
			case <-c.snapshots:
			default:
			}
			c.snapshots <- message
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			if !matchStarted && !startAt.IsZero() && !now.Before(startAt) {
				if world.Finished() {
					world.RestartRound()
					roundID++
				}
				world.DiscardInputs()
				startAt = time.Time{}
				matchStarted = true
			}
			if matchStarted {
				wasFinished := world.Finished()
				world.Step(now)
				if !wasFinished && world.Finished() {
					for c := range clients {
						c.ready = false
					}
				}
			}
			if matchStarted && world.Tick()%(simulation.TickRate/parameters.SnapshotRate) == 0 || !matchStarted && now.Sub(lastBroadcast) >= parameters.LobbySnapshotInterval {
				broadcast()
			}
		case c := <-h.join:
			nextID++
			c.id = fmt.Sprintf("p%d", nextID)
			if !world.Join(c.id) {
				c.joined <- false
				continue
			}
			c.joined <- true
			clients[c] = struct{}{}
			names[c.id] = c.name
			world.SetCharacter(c.id, c.character)
			c.send <- protocol.Message{Type: protocol.TypeWelcome, PlayerID: c.id}
			broadcast()
		case c := <-h.leave:
			if _, exists := clients[c]; exists {
				remove(c)
				broadcast()
			}
		case event := <-h.messages:
			if _, exists := clients[event.client]; !exists {
				continue
			}
			if !matchStarted && (event.message.Type == protocol.TypePerkChoice || event.message.Type == protocol.TypeCloseOffer || event.message.Type == protocol.TypeRespawn) {
				continue
			}
			switch event.message.Type {
			case protocol.TypeStartGame, protocol.TypeRematch:
				rematch := event.message.Type == protocol.TypeRematch
				if !briefingActive && startAt.IsZero() && (rematch && world.Finished() || !rematch && !matchStarted && !world.Finished()) {
					event.client.ready = true
					allReady := len(clients) == simulation.MaxPlayers
					for c := range clients {
						allReady = allReady && c.ready
					}
					if allReady {
						matchStarted = false
						allBriefed := true
						for c := range clients {
							allBriefed = allBriefed && c.briefingDone
						}
						if rematch || allBriefed {
							startAt = time.Now().Add(parameters.MatchCountdownSeconds * time.Second)
						} else {
							briefingActive = true
						}
					}
					broadcast()
				}
				continue
			case protocol.TypeBriefingDone:
				if briefingActive {
					event.client.briefingDone = true
					allBriefed := len(clients) == simulation.MaxPlayers
					for c := range clients {
						allBriefed = allBriefed && c.ready && c.briefingDone
					}
					if allBriefed {
						briefingActive = false
						startAt = time.Now().Add(parameters.MatchCountdownSeconds * time.Second)
					}
					broadcast()
				}
				continue
			case protocol.TypePerkChoice:
				choice := event.message.Perk
				world.ChoosePerk(event.client.id, choice.StatueID, choice.Cycle, choice.Card)
				broadcast()
				continue
			case protocol.TypeCloseOffer:
				world.CloseOffer(event.client.id)
				broadcast()
				continue
			case protocol.TypeRespawn:
				world.Respawn(event.client.id)
				broadcast()
				continue
			}
			if event.message.Type == protocol.TypeInput {
				d := event.message.Direction
				if !world.Command(event.client.id, d.Simulation(), event.message.Sequence, time.Now()) {
					remove(event.client)
					broadcast()
				}
				if !matchStarted {
					world.DiscardInputs()
				}
				continue
			}
			if event.message.Type == protocol.TypeChat {
				text := protocol.NormalizeChatMessage(event.message.Message)
				now := time.Now()
				if text == "" || now.Sub(event.client.lastChat) < chatCooldown {
					continue
				}
				event.client.lastChat = now
				chat := protocol.Message{
					Type: protocol.TypeChat, Tick: world.Tick(),
					PlayerID: event.client.id, Name: names[event.client.id], Message: text,
				}
				dropped := false
				for c := range clients {
					select {
					case c.send <- chat:
					default:
						remove(c)
						dropped = true
					}
				}
				if dropped {
					broadcast()
				}
				continue
			}
			reply := protocol.Message{Type: protocol.TypeEcho, Message: "received: " + event.message.Message}
			select {
			case event.client.send <- reply:
			default:
				remove(event.client)
				broadcast()
			}
		}
	}
}
