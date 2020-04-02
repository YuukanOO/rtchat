package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/yuukanoo/rtchat/internal/logging"
	"github.com/yuukanoo/rtchat/internal/service"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const checkDelay = 15 * time.Second

type (
	// Server represents a Websocket server.
	Server interface {
		// Handle should be used in a router to process an incoming ws request.
		Handle(http.ResponseWriter, *http.Request)
		// CheckEmptiness inform the websocket server to check the given room for
		// emptiness in some time.
		// This will prevent the creation of empty rooms which are not desirable
		// for this tiny server.
		CheckEmptiness(string)
		// Run the realtime server.
		Run() error
		// Close the current server and all connections.
		Close() error
	}

	// GetRouteParamFunc is needed to extract a route parameter from a request
	// no matter which router has been chosen.
	GetRouteParamFunc func(*http.Request, string) string

	hub struct {
		service       service.Service
		logger        logging.Logger
		isRunning     bool
		getRouteParam GetRouteParamFunc
		register      chan *client
		unregister    chan *client
		quit          chan bool
		clients       map[string]*client
		rooms         map[string][]*client
		check         chan string
		send          chan *message
	}
)

// New instantiates a new websocket server to process realtime requests.
// It expects the route to have an url param named "id" which represents the room
// identifier.
func New(service service.Service, logger logging.Logger, fn GetRouteParamFunc) Server {
	return &hub{
		logger:        logger,
		service:       service,
		getRouteParam: fn,
		register:      make(chan *client),
		unregister:    make(chan *client),
		quit:          make(chan bool),
		clients:       make(map[string]*client),
		rooms:         make(map[string][]*client),
		check:         make(chan string),
		send:          make(chan *message),
	}
}

func (h *hub) CheckEmptiness(id string) {
	go func() {
		<-time.After(checkDelay)
		h.check <- id
	}()
}

func (h *hub) Handle(w http.ResponseWriter, r *http.Request) {
	cred := r.Header.Get("Sec-WebSocket-Protocol")

	// Checks if the room exists first, if not, just returns a 404
	room := h.service.GetRoom(h.getRouteParam(r, "id"))

	if room == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Checks if credentials are good to prevent anyone to access the websocket and
	// received messages for this room.
	if room.Credential != cred {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.register <- newClient(h, room.ID, conn)
}

func (h *hub) Run() error {
	if h.isRunning {
		return fmt.Errorf("server is already running")
	}

	h.isRunning = true

	for {
		select {

		case c := <-h.register:
			clients := h.rooms[c.room]

			h.rooms[c.room] = append(clients, c)
			h.clients[c.id] = c

			go c.readPump()
			go c.writePump()

			h.logger.Debug(`joined:
	user: %s
	room: %s`, c.id, c.room)

			// Notify every other user in the same room that a new user has joined
			go func() {
				h.send <- &message{
					room: c.room,
					From: c.id,
					Joined: &joinedPayload{
						ID: c.id,
					},
				}
			}()

		case c := <-h.unregister:
			h.clients[c.id] = nil
			delete(h.clients, c.id)

			clients := h.rooms[c.room]

			for i, cli := range clients {
				if cli == c {
					clients[i] = nil
					h.rooms[c.room] = append(clients[:i], clients[i+1:]...)
					break
				}
			}

			h.logger.Debug(`left:
	user: %s
	room: %s`, c.id, c.room)

			// Trigger a check for emptiness in a while
			h.CheckEmptiness(c.room)

			// Notify every other user in the same room that a user has left
			go func() {
				h.send <- &message{
					room: c.room,
					From: c.id,
					Left: &leftPayload{
						ID: c.id,
					},
				}
			}()

		case id := <-h.check:
			// Delete the room if there is no more user
			if len(h.rooms[id]) == 0 {
				h.rooms[id] = nil
				delete(h.rooms, id)
				h.logger.Debug("No users left in %s, deleting", id)

				go h.service.DeleteRoom(id)
			}

		case m := <-h.send:
			if m.To != "" {
				// If it should be sent to one client in particular
				c := h.clients[m.To]

				// Check the origin
				if c.room == m.room {
					c.send <- m
				}
			} else {
				// Else broadcast the message to everyone in the same room
				clients := h.rooms[m.room]

				for _, c := range clients {
					if c.id != m.From {
						c.send <- m
					}
				}
			}

		case <-h.quit:
			for _, c := range h.clients {
				c.conn.Close()
			}
			return nil
		}
	}
}

func (h *hub) Close() error {
	if !h.isRunning {
		return nil
	}

	h.quit <- true

	return nil
}
