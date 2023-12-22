package special

import "github.com/ystv/streamer/common/transporter/server"

type (
	// InitiationTransport allows for both the server type and the version to be communicated to the server
	InitiationTransport struct {
		// Server is the server type
		Server server.Server
		// Version is the string of the version
		Version string
	}
)
