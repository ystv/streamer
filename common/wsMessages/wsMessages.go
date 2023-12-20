package wsMessages

type WSMessage string

var (
	Okay  WSMessage = "OKAY"
	Error WSMessage = "ERROR"
)

func (w WSMessage) String() string {
	return string(w)
}
