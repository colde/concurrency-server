package main

import (
	"bytes"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is an middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// current video playing
	media string

	// position of video
	pos int

	// user
	user int

	// Is the user an admin
	admin bool
}

// Handle new playback sessions starting
func handleNewSessionCommand(c *Client, command string) {
	parts := strings.Split(command, " ")
	if len(parts) != 3 {
		// Command didn't follow syntax of NEW <media> <user>
		log.Print("Invalid new command " + command)
	}

	c.media = parts[1]
	log.Printf("Set media to %s for client %s", c.media, c.conn.RemoteAddr())

	i, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Print("Unable to decode user id ", parts[2])
	} else {
		c.user = i
		log.Printf("Set user to %d for client %s", c.user, c.conn.RemoteAddr())
	}
}

// Update position for clients
func handlePosUpdateCommand(c *Client, command string) {
	p := command[4:]
	i, err := strconv.Atoi(p)
	if err != nil {
		log.Print("Unable to decode position ", p)
	} else {
		c.pos = i
		log.Printf("Set position to %d for client %s", c.pos, c.conn.RemoteAddr())
	}
}

// Set status to admin if pw is correct
func handleAdminCommand(c *Client, command string) {
	pw := command[6:]
	if pw == adminpw {
		c.admin = true
		log.Print("Client ", c.conn.RemoteAddr(), " marked as admin")
		c.send <- []byte("ADMIN OK")
	} else {
		log.Print(" --- WARNING --- WRONG ADMIN PW FROM ", c.conn.RemoteAddr())
	}
}

// Parse received websocket messages
func handleMessage(c *Client, message []byte) {
	s := string(message[:])

	if strings.HasPrefix(s, "NEW") {
		handleNewSessionCommand(c, s)
	} else if strings.HasPrefix(s, "POS") {
		handlePosUpdateCommand(c, s)
	} else if strings.HasPrefix(s, "ADMIN") {
		handleAdminCommand(c, s)
	} else if strings.HasPrefix(s, "STATUS") && c.admin {
		log.Print("Admin ", c.conn.RemoteAddr(), " asked for status")
		c.hub.status <- true
	} else {
		c.send <- []byte("INVALID COMMAND")
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		handleMessage(c, message)
	}
}

// write writes a message with the given message type and payload.
func (c *Client) write(mt int, payload []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client
	go client.writePump()
	client.readPump()
}
