package websocket

import (
	"time"

	"github.com/yuukanoo/rtchat/internal/crypto"

	"github.com/gorilla/websocket"
)

type (
	client struct {
		id   string
		room string
		conn *websocket.Conn
		send chan *message
		hub  *hub
	}
)

func newClient(hub *hub, room string, conn *websocket.Conn) *client {
	return &client{
		id:   crypto.GenerateUID(64),
		room: room,
		conn: conn,
		hub:  hub,
		send: make(chan *message),
	}
}

func (c *client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var m message

		if err := c.conn.ReadJSON(&m); err != nil {
			return
		}

		// Assign some invariant metadata to the received message
		m.room = c.room
		m.From = c.id

		if m.IsAllowed() {
			c.hub.send <- &m
		}
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(15 * time.Second)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		var err error

		select {

		case <-ticker.C:
			// Writes an empty message to keep the connection alive
			err = c.conn.WriteJSON(message{})

		case m := <-c.send:
			err = c.conn.WriteJSON(m)

		}

		if err != nil {
			return
		}

	}
}
