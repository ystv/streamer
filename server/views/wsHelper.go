package views

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/server"
)

func (v *Views) wsHelper(name server.Server, transporter commonTransporter.Transporter) (commonTransporter.ResponseTransporter, error) {
	out, valid := v.cache.Get(name.String())
	if !valid {
		return commonTransporter.ResponseTransporter{}, fmt.Errorf("channel %s is not valid", name)
	}

	returningChannel := make(chan []byte)

	sendingTransporter := TransporterRouter{
		TransporterUnique: commonTransporter.TransporterUnique{
			ID:      uuid.NewString(),
			Payload: transporter,
		},
		ReturningChannel: returningChannel,
	}

	log.Printf("sending message to %s: %#v", name, transporter)

	out.(chan TransporterRouter) <- sendingTransporter

	received := <-returningChannel

	var responseTransporter commonTransporter.ResponseTransporter

	err := json.Unmarshal(received, &responseTransporter)
	if err != nil {
		return commonTransporter.ResponseTransporter{}, fmt.Errorf("failed to unmarshal response: %w, received message: %s", err, string(received))
	}

	return responseTransporter, nil
}
