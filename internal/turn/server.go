package turn

import (
	"strconv"

	"github.com/yuukanoo/rtchat/internal/logging"
	"github.com/yuukanoo/rtchat/internal/service"

	"github.com/pion/turn/v2"

	"net"
)

type (
	// Options needed by the turn server to be instantiated correclty.
	Options interface {
		// Realm used by the turn server.
		Realm() string
		// PublicIP at which the turn server will be publicly accessible.
		PublicIP() net.IP
		// Port at which the turn server will be made available.
		Port() int
	}

	// Server made available to traverse NAT.
	Server interface {
		// Close the server and stops the listener.
		Close() error
	}
)

// New instantiates a new turn server.
func New(service service.Service, logger logging.Logger, options Options) (Server, error) {
	udpListener, err := net.ListenPacket("udp4", "0.0.0.0:"+strconv.Itoa(options.Port()))

	if err != nil {
		return nil, err
	}

	s, err := turn.NewServer(turn.ServerConfig{
		Realm: options.Realm(),
		AuthHandler: func(username, realm string, srcAddr net.Addr) (key []byte, ok bool) {
			room := service.GetRoom(username)

			if room == nil {
				return nil, false
			}

			// TODO maybe cache this thing
			return turn.GenerateAuthKey(room.ID, realm, room.Credential), true
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: options.PublicIP(), // Claim that we are listening on IP passed by user (This should be your Public IP)
					Address:      "0.0.0.0",          // But actually be listening on every interface
				},
			},
		},
	})

	logger.Info(`TURN/STUN Server launched:
	Realm:		%s
	Public IP:	%s
	Port:		%d`, options.Realm(), options.PublicIP(), options.Port())

	return s, err
}
