package server

type Server string

const (
	// Forwarder is the type denoted for the forwarder server
	Forwarder Server = "forwarder"
	// Recorder is the type denoted for the recorder server
	Recorder Server = "recorder"
)

// String returns the string of the server
func (s Server) String() string {
	return string(s)
}
