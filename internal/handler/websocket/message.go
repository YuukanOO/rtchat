package websocket

type (
	joinedPayload struct {
		ID string `json:"id"`
	}

	leftPayload struct {
		ID string `json:"id"`
	}

	sdpPayload struct {
		Type string `json:"type"`
		SDP  string `json:"sdp"`
	}

	icePayload struct {
		Candidate        string `json:"candidate"`
		SDPMLineIndex    int    `json:"sdpMLineIndex"`
		SDPMid           string `json:"sdpMid"`
		UsernameFragment string `json:"usernameFragment"`
	}

	// message hold every message payload that exists in the system.
	message struct {
		room string

		From string `json:"from,omitempty"`
		To   string `json:"to,omitempty"`

		// Server based messages
		Joined *joinedPayload `json:"joined,omitempty"`
		Left   *leftPayload   `json:"left,omitempty"`

		// Client messages
		Offer  *sdpPayload `json:"offer,omitempty"`
		Answer *sdpPayload `json:"answer,omitempty"`
		ICE    *icePayload `json:"ice,omitempty"`
	}
)

// IsAllowed checks if this message is allowed from client.
// It prevents malicious message sending without making the websocket stuff too
// complex.
func (m *message) IsAllowed() bool {
	return m.Joined == nil && m.Left == nil
}
