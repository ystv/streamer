package wsMessages

type WSMessage string

var (
	Acknowledged WSMessage = "ACKNOWLEDGED"
	Ping         WSMessage = "PING"
	Pong         WSMessage = "PONG"
	Okay         WSMessage = "OKAY"
	Error        WSMessage = "ERROR"
)

func (w WSMessage) String() string {
	return string(w)
}
