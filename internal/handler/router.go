package handler

import (
	"html/template"
	"net/http"

	"github.com/yuukanoo/rtchat/internal/handler/websocket"
	"github.com/yuukanoo/rtchat/internal/logging"
	"github.com/yuukanoo/rtchat/internal/service"

	"github.com/go-chi/chi"
)

type (
	// Router configured and ready to be hosted.
	Router interface {
		// Handler which process incoming connections.
		Handler() http.Handler
		// Close the current handler router and release needed resources.
		Close() error
	}

	// Options holds needed configuration for the router.
	Options interface {
		// TurnURL represents the turn server which should be used for the
		// communication.
		TurnURL() string
		// StunURL represents the stun server used to establish connections.
		StunURL() string
	}

	router struct {
		options Options
		service service.Service
		ws      websocket.Server
		*chi.Mux

		// Templates
		homeTpl *template.Template
		roomTpl *template.Template
	}
)

// New instantiates a new http handler ready to be used with an http server.
func New(service service.Service, logger logging.Logger, options Options) (Router, error) {
	r := &router{
		options: options,
		service: service,
		ws:      websocket.New(service, logger, chi.URLParam),
		Mux:     chi.NewRouter(),

		homeTpl: template.Must(template.ParseFiles("templates/index.html")),
		roomTpl: template.Must(template.ParseFiles("templates/room.html")),
	}

	r.Get("/ws/{id}", r.ws.Handle)
	r.Post("/rooms", r.CreateRoom)
	r.Get("/rooms/{id}", r.ShowRoom)
	r.Get("/", r.ShowHome)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Launch the websocket server!
	go r.ws.Run()

	return r, nil
}

func (r *router) ShowHome(w http.ResponseWriter, req *http.Request) {
	r.homeTpl.Execute(w, nil)
}

func (r *router) CreateRoom(w http.ResponseWriter, req *http.Request) {
	id := r.service.CreateRoom()

	// Check that the room has at least one user in it in a while to prevent
	// empty rooms for staying forever.
	r.ws.CheckEmptiness(id)

	http.Redirect(w, req, "/rooms/"+id, http.StatusSeeOther)
}

func (r *router) ShowRoom(w http.ResponseWriter, req *http.Request) {
	room := r.service.GetRoom(chi.URLParam(req, "id"))

	if room == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	r.roomTpl.Execute(w, struct {
		RoomID         string
		RoomCredential string
		TurnURL        string
		StunURL        string
		Username       string
		Credential     string
	}{
		RoomID:         room.ID,
		RoomCredential: room.Credential,
		TurnURL:        r.options.TurnURL(),
		StunURL:        r.options.StunURL(),
		Username:       room.ID,
		Credential:     room.Credential,
	})
}

func (r *router) Close() error {
	return r.ws.Close()
}

func (r *router) Handler() http.Handler {
	return r
}
