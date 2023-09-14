package views

import (
	"encoding/json"
	"fmt"
)

func (v *Views) wsHelper(name string, transporter Transporter) (string, error) {
	out, valid := v.cache.Get(name)
	if !valid {
		return "", fmt.Errorf("channel %s is not valid", name)
	}

	b, err := json.Marshal(transporter)
	if err != nil {
		return "", fmt.Errorf("failed marshaling transporter: %w", err)
	}

	out.(chan []byte) <- b

	in, valid := v.cache.Get(name + "Internal")
	if !valid {
		return "", fmt.Errorf("channel %sInternal is not valid", name)
	}

	received := <-in.(chan []byte)

	return string(received), nil
}
