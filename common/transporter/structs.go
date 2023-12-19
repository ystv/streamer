package transporter

import "github.com/ystv/streamer/common/transporter/action"

type (
	Transporter struct {
		Action  action.Action `json:"action"`
		Unique  string        `json:"unique"`
		Payload interface{}   `json:"payload"`
	}

	ForwarderStart struct {
		StreamIn   string   `json:"streamIn"`
		WebsiteOut string   `json:"websiteOut"`
		Streams    []string `json:"streams"`
	}

	ForwarderStatus struct {
		Website bool `json:"website"`
		Streams int  `json:"streams"`
	}

	ForwarderStatusResponse struct {
		Website string            `json:"website"`
		Streams map[uint64]string `json:"streams"`
	}

	RecorderStart struct {
		StreamIn string `json:"streamIn"`
		PathOut  string `json:"pathOut"`
	}
)

// ResponseSeparator is a random string separator between the response status and body
const ResponseSeparator = "+~+"
