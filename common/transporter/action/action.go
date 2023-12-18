package action

type Action string

const (
	Start  Action = "start"
	Status Action = "status"
	Stop   Action = "stop"
)

func (a Action) String() string {
	return string(a)
}
