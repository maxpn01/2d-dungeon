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
			message := protocol.SnapshotFromSimulation(world.Snapshot(), names).Message()
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
			switch event.message.Type {
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
