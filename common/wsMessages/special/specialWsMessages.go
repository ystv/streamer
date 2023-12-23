package wsMessages

type SpecialWSMessage string

var (
	// Acknowledged is the special web socket message sent during the connection phase
	Acknowledged SpecialWSMessage = "ACKNOWLEDGED"
	// Ping is sent from the server to the forwarder and recorder to validate the connection
	Ping SpecialWSMessage = "PING"
	// Pong is sent from the forwarder or recorder to the server to validate the connection in response to the Ping
	Pong SpecialWSMessage = "PONG"
)

// String returns the string of the special websocket message
func (s SpecialWSMessage) String() string {
	return string(s)
}
