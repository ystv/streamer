package wsMessages

type WSMessage string

var (
	Acknowledged WSMessage = "ACKNOWLEDGED"
	Ping         WSMessage = "ping"
	Pong         WSMessage = "pong"
	Okay         WSMessage = "OKAY"
)

func (r WSMessage) String() string {
	return string(r)
}
