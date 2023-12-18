package views

import (
	"encoding/json"
	"fmt"

	"github.com/ystv/streamer/server/helper/transporter/server"
)

func (v *Views) wsHelper(name server.Server, transporter Transporter) (string, error) {
	out, valid := v.cache.Get(name.String())
	if !valid {
		return "", fmt.Errorf("channel %s is not valid", name)
	}

	b, err := json.Marshal(transporter)
	if err != nil {
		return "", fmt.Errorf("failed marshaling transporter: %w", err)
	}

	out.(chan []byte) <- b

	in, valid := v.cache.Get(name.String() + "Internal")
	if !valid {
		return "", fmt.Errorf("channel %sInternal is not valid", name)
	}

	received := <-in.(chan []byte)

	return string(received), nil
}
