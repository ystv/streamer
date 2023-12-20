package wsMessages

type SpecialWSMessage string

var (
	Acknowledged SpecialWSMessage = "ACKNOWLEDGED"
	Ping         SpecialWSMessage = "PING"
	Pong         SpecialWSMessage = "PONG"
)

func (s SpecialWSMessage) String() string {
	return string(s)
}
