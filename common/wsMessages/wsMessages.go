package wsMessages

type WSMessage string

var (
	// Okay is the response sent back to the server when everything was okay
	Okay WSMessage = "OKAY"
	// Error is the response sent back to the server when something has gone wrong
	Error WSMessage = "ERROR"
)

// String returns the string of the web socket
func (w WSMessage) String() string {
	return string(w)
}
