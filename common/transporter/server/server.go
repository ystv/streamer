package server

type Server string

const (
	Forwarder Server = "forwarder"
	Recorder  Server = "recorder"
)

func (s Server) String() string {
	return string(s)
}
