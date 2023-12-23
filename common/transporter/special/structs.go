package special

import "github.com/ystv/streamer/common/transporter/server"

type (
	// InitiationTransporter allows for both the server type and the version to be communicated to the server
	InitiationTransporter struct {
		// Server is the server type
		Server server.Server `json:"server"`
		// Version is the string of the version
		Version string `json:"version"`
	}
)
