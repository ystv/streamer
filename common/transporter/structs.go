package transporter

import (
	"github.com/ystv/streamer/common/transporter/action"
	"github.com/ystv/streamer/common/wsMessages"
)

type (
	// TransporterUnique is sent with an ID in order to be returned and sorted
	TransporterUnique struct {
		// ID is the unique code used to identify a request
		ID string `json:"id" mapstructure:"id"`
		// Payload is used for either Transporter or a ping
		Payload interface{} `json:"payload" mapstructure:"payload"`
	}

	// Transporter is the parent struct to send to a recipient and must always be used except with a ping
	Transporter struct {
		// Action is the action that is to be performed by the recipient
		Action action.Action `json:"action" mapstructure:"action"`
		// Unique is the unique code for each stream
		Unique string `json:"unique" mapstructure:"unique"`
		// Payload is the data package to send to the recipient, start or status information but can be left blank
		Payload interface{} `json:"payload,omitempty" mapstructure:"payload"`
	}

	// ForwarderStart is the payload in the Transporter for starting stream forwarding
	ForwarderStart struct {
		// StreamIn is the incoming stream to forward
		StreamIn string `json:"streamIn" mapstructure:"streamIn"`
		// WebsiteOut indicates the endpoint to send the stream to
		WebsiteOut string `json:"websiteOut,omitempty" mapstructure:"websiteOut"`
		// Streams is the list of all other stream endpoints to send the stream to
		Streams []string `json:"streams" mapstructure:"streams"`
	}

	// ForwarderStatus is the payload in the Transporter for getting the status forwarding
	ForwarderStatus struct {
		// Website indicates if the website needs to be accounted for in the log collection
		Website bool `json:"website" mapstructure:"website"`
		// Streams is the number of forwarded streams are to be collected
		Streams int `json:"streams" mapstructure:"website"`
	}

	// RecorderStart is the payload in the Transporter for starting stream recording
	RecorderStart struct {
		// StreamIn is the incoming stream to forward
		StreamIn string `json:"streamIn" mapstructure:"streamIn"`
		// PathOut is the requested path for the stream to be saved to
		PathOut string `json:"pathOut" mapstructure:"pathOut"`
	}

	// ResponseTransporter is the parent struct to send to the server and must always be used except with a pong
	ResponseTransporter struct {
		// Status is the status of the response, either okay or error and indicates the success of the request
		Status wsMessages.WSMessage `json:"status" mapstructure:"status"`
		// Payload is the data package to send to the server, status information but can be left blank
		Payload interface{} `json:"payload,omitempty" mapstructure:"payload"`
	}

	// ForwarderStatusResponse is the payload in the ResponseTransporter for starting stream forwarding
	ForwarderStatusResponse struct {
		// Website contains the log data of the website stream
		Website string `json:"website" mapstructure:"website"`
		// Streams contain the log data of all the forwarded streams
		Streams map[string]string `json:"streams" mapstructure:"streams"`
	}
)
