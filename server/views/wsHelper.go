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

	//b, err := json.Marshal(sendingTransporter)
	//if err != nil {
	//	return commonTransporter.ResponseTransporter{}, fmt.Errorf("failed marshaling transporter: %w", err)
	//}

	out.(chan TransporterRouter) <- sendingTransporter

	//in, valid := v.cache.Get(name.String() + internalChannelNameAppend)
	//if !valid {
	//	return commonTransporter.ResponseTransporter{}, fmt.Errorf("channel %s%s is not valid", name, internalChannelNameAppend)
	//}

	received := <-returningChannel

	var responseTransporter commonTransporter.ResponseTransporter

	err := json.Unmarshal(received, &responseTransporter)
	if err != nil {
		return commonTransporter.ResponseTransporter{}, fmt.Errorf("failed to unmarshal response: %w, recieved message: %s", err, string(received))
	}

	return responseTransporter, nil
}
