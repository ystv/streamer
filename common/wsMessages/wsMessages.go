package wsMessages

type WSMessage string

var (
	Acknowledged WSMessage = "ACKNOWLEDGED"
)

func (r WSMessage) String() string {
	return string(r)
}
