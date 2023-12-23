package action

type Action string

const (
	// Start is the action sent to the recorder and forwarder when a stream starts
	Start Action = "start"
	// Status is the action sent to the recorder and forwarder when the ffmpeg logs are required
	Status Action = "status"
	// Stop is the action sent to the recorder and forwarder when the server is meant to be stopped
	Stop Action = "stop"
)

// String returns the string of the action
func (a Action) String() string {
	return string(a)
}
