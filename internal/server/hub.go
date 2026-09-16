package server

import (
	"context"
	"fmt"
	"sort"

	"raylib-2d-sandbox/internal/protocol"
)

type client struct {
	id     string
	send   chan protocol.Message
	cancel context.CancelFunc
}

type clientMessage struct {
	client  *client
	message protocol.Message
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
	nextID := 0
	remove := func(c *client) {
		delete(clients, c)
		c.cancel()
	}
	defer func() {
		for c := range clients {
			remove(c)
		}
	}()

	broadcast := func() {
		for {
			players := make([]protocol.Player, 0, len(clients))
			for c := range clients {
				players = append(players, protocol.Player{ID: c.id})
			}
			sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
			message := protocol.Message{Type: protocol.TypeSnapshot, Players: &players}
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
		case c := <-h.join:
			nextID++
			c.id = fmt.Sprintf("p%d", nextID)
			clients[c] = struct{}{}
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
			select {
			case event.client.send <- event.message:
			default:
				remove(event.client)
				broadcast()
			}
		}
	}
}
