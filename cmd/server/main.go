package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/yuukanoo/rtchat/internal/handler"
	"github.com/yuukanoo/rtchat/internal/logging"
	"github.com/yuukanoo/rtchat/internal/service"
	"github.com/yuukanoo/rtchat/internal/turn"
)

func main() {
	e := flags{
		debug: flag.Bool("debug", false, "Should we launch in the debug mode?"),
		turn: turnFlags{
			realm:    flag.String("realm", "rtchat.io", "Realm used by the turn server."),
			publicIP: flag.String("turn-ip", "192.168.0.14", "IP Address that TURN can be contacted on. Should be publicly available."),
			port:     flag.Int("turn-port", 3478, "Listening port for the TURN/STUN endpoint."),
		},
		web: webFlags{
			port: flag.Int("http-port", 5000, "Web server listening port."),
		},
	}

	flag.Parse()

	logger := logging.New(*e.debug)

	if *e.debug {
		logger.Info("launched in debug mode, extra output is expected")
	}

	// Instantiates the service that creates rooms
	service := service.New()

	// Instantiate and launch the turn server
	turnServer, err := turn.New(service, logger, &e.turn)

	if err != nil {
		panic(err)
	}

	defer turnServer.Close()

	// Instantiate the application router
	r, err := handler.New(service, logger, &e.turn)

	if err != nil {
		panic(err)
	}

	defer r.Close()

	// Launch the HTTP server!
	s := &http.Server{Handler: r.Handler(), Addr: e.web.Address()}

	go func() {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()

	defer s.Close()

	logger.Info(`HTTP server launched:
	Listening:	%s`, e.web.Address())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	logger.Info("Shutting down, goodbye 👋")
}

// flags represents options which can be passed to internal packages.
type flags struct {
	debug *bool
	turn  turnFlags
	web   webFlags
}

// turnFlags contains turn server related configuration.
type turnFlags struct {
	realm    *string
	publicIP *string
	port     *int
}

// webFlags contains web specific flags.
type webFlags struct {
	port *int
}

func (f *turnFlags) Realm() string    { return *f.realm }
func (f *turnFlags) PublicIP() net.IP { return net.ParseIP(*f.publicIP) }
func (f *turnFlags) Port() int        { return *f.port }
func (f *turnFlags) TurnURL() string  { return fmt.Sprintf("turn:%s:%d", *f.publicIP, *f.port) }
func (f *turnFlags) StunURL() string  { return fmt.Sprintf("stun:%s:%d", *f.publicIP, *f.port) }

func (f *webFlags) Address() string { return fmt.Sprintf(":%d", *f.port) }
