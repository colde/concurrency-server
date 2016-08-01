package main

import (
	"fmt"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Send status to all clients
	status chan bool
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		status:     make(chan bool),
	}
}

func sendToAllAdmins(h *Hub, message []byte) {
	for client := range h.clients {
		if client.admin {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}

func sendToAll(h *Hub, message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func sendStatus(h *Hub) {
	for client := range h.clients {
		if client.admin {
			continue
		}
		m := fmt.Sprintf("CLIENT %s %d %s %d", client.conn.RemoteAddr(), client.user, client.media, client.pos)
		sendToAllAdmins(h, []byte(m))
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Print("Client joined ", client.conn.RemoteAddr())
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			log.Print("Client left ", client.conn.RemoteAddr())
		case message := <-h.broadcast:
			sendToAll(h, message)
		case _ = <-h.status:
			sendToAllAdmins(h, []byte("STATUS"))
			sendStatus(h)
			sendToAllAdmins(h, []byte("END OF STATUS"))
		}
	}
}
