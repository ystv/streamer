package views

import (
	"encoding/json"
	"fmt"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/server"
)

func (v *Views) wsHelper(name server.Server, transporter commonTransporter.Transporter) (commonTransporter.ResponseTransporter, error) {
	out, valid := v.cache.Get(name.String())
	if !valid {
		return commonTransporter.ResponseTransporter{}, fmt.Errorf("channel %s is not valid", name)
	}

	b, err := json.Marshal(transporter)
	if err != nil {
		return commonTransporter.ResponseTransporter{}, fmt.Errorf("failed marshaling transporter: %w", err)
	}

	out.(chan []byte) <- b

	in, valid := v.cache.Get(name.String() + internalChannelNameAppend)
	if !valid {
		return commonTransporter.ResponseTransporter{}, fmt.Errorf("channel %s%s is not valid", name, internalChannelNameAppend)
	}

	received := <-in.(chan []byte)

	var responseTransporter commonTransporter.ResponseTransporter

	err = json.Unmarshal(received, &responseTransporter)
	if err != nil {
		return commonTransporter.ResponseTransporter{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return responseTransporter, nil
}
