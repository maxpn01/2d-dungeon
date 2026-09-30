package server

import (
	"context"
	"fmt"
	"time"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
)

type client struct {
	character simulation.Character
	name      string
	id        string
	send      chan protocol.Message
	cancel    context.CancelFunc
	lastChat  time.Time
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
	defer close(h.done)
	clients := make(map[*client]struct{})
	names := make(map[string]string)
	world := simulation.NewWorld()
	ticker := time.NewTicker(time.Second / simulation.TickRate)
	defer ticker.Stop()
	nextID := 0
	remove := func(c *client) {
		delete(clients, c)
		delete(names, c.id)
		world.Leave(c.id)
		c.cancel()
	}
	defer func() {
		for c := range clients {
			remove(c)
		}
	}()

	broadcast := func() {
		for {
			snapshot := world.Snapshot()
			players := make([]protocol.Player, 0, len(snapshot.Players))
			for _, state := range snapshot.Players {
				players = append(players, protocol.Player{
					EjectCooldown: state.EjectCooldown, Character: state.Character, Split: state.Split, GatesClosed: state.GatesClosed, InArena: state.InArena, TeleportCharge: state.TeleportCharge, Combat: state.Combat, Name: names[state.ID], Dead: state.Dead, KilledBy: state.KilledBy, ID: state.ID, X: state.X, Y: state.Y, Size: simulation.PlayerSize, InputSeq: state.InputSeq,
					Stats: protocol.PlayerStats{
						Level: state.Stats.Level, XP: state.Stats.XP, HP: state.Stats.HP, MP: state.Stats.MP, Speed: state.Stats.Speed,
					},
				})
			}
			fruits := make([]protocol.Fruit, 0, len(snapshot.Fruits))
			for _, fruit := range snapshot.Fruits {
				fruits = append(fruits, protocol.Fruit{Source: fruit.Source, HP: fruit.HP, MP: fruit.MP, ID: fruit.ID, X: fruit.X, Y: fruit.Y, Size: fruit.Size})
			}
			ais := make([]protocol.AI, 0, len(snapshot.AIs))
			for _, ai := range snapshot.AIs {
				ais = append(ais, protocol.AI{
					Combat: ai.Combat, Behavior: ai.Behavior, Character: ai.Character, Split: ai.Split, ID: ai.ID, X: ai.X, Y: ai.Y, Size: simulation.AISize,
					Stats: protocol.PlayerStats{Level: ai.Stats.Level, XP: ai.Stats.XP, HP: ai.Stats.HP, MP: ai.Stats.MP, Speed: ai.Stats.Speed},
				})
			}
			message := protocol.Message{SafeZones: snapshot.SafeZones, Boss: snapshot.Boss, Type: protocol.TypeSnapshot, Tick: snapshot.Tick, Players: &players, Fruits: &fruits, AIs: &ais}
			dropped := false
			for c := range clients {
				select {
				case c.send <- message:
				default:
					remove(c)
					dropped = true
				}
			}
			if !dropped {
				return
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			world.Step(time.Now())
			broadcast()
		case c := <-h.join:
			nextID++
			c.id = fmt.Sprintf("p%d", nextID)
			clients[c] = struct{}{}
			names[c.id] = c.name
			world.Join(c.id)
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
			if event.message.Type == protocol.TypeInput {
				d := event.message.Direction
				if !world.Command(event.client.id, simulation.Direction{X: d.X, Y: d.Y, Eject: d.Eject, Boost: d.Boost, Dash: d.Dash, AimX: d.AimX, AimY: d.AimY}, event.message.Sequence, time.Now()) {
					remove(event.client)
					broadcast()
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
