package server

import (
	"context"
	"fmt"
	"time"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
)

type client struct {
	id     string
	send   chan protocol.Message
	cancel context.CancelFunc
}

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
	world := simulation.NewWorld()
	ticker := time.NewTicker(time.Second / simulation.TickRate)
	defer ticker.Stop()
	nextID := 0
	remove := func(c *client) {
		delete(clients, c)
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
					ID: state.ID, X: state.X, Y: state.Y, Size: simulation.PlayerSize, InputSeq: state.InputSeq,
					Stats: protocol.PlayerStats{
						Level: state.Stats.Level, XP: state.Stats.XP, HP: state.Stats.HP, Speed: state.Stats.Speed,
					},
				})
			}
			fruits := make([]protocol.Fruit, 0, len(snapshot.Fruits))
			for _, fruit := range snapshot.Fruits {
				fruits = append(fruits, protocol.Fruit{ID: fruit.ID, X: fruit.X, Y: fruit.Y, Size: fruit.Size})
			}
			message := protocol.Message{Type: protocol.TypeSnapshot, Tick: snapshot.Tick, Players: &players, Fruits: &fruits}
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
			tick := world.Tick()
			if tick%2 == 0 {
				broadcast()
			}
		case c := <-h.join:
			nextID++
			c.id = fmt.Sprintf("p%d", nextID)
			clients[c] = struct{}{}
			world.Join(c.id)
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
				if !world.Command(event.client.id, simulation.Direction{X: d.X, Y: d.Y}, event.message.Sequence, time.Now()) {
					remove(event.client)
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
