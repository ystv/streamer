package views

import (
	"fmt"
)

func (v *Views) generateUnique() (string, error) {
	var b []byte

	loop := true

	for loop {
		b = make([]byte, 10)
		for i := range b {
			b[i] = charset[seededRand.Intn(len(charset))]
		}

		streams1, err := v.store.GetStreams()
		if err != nil {
			return "", fmt.Errorf("failed to get streams: %+v", err)
		}

		if len(streams1) == 0 {
			break
		}

		for _, s := range streams1 {
			if s.Stream == string(b) {
				loop = true
				break
			}
			loop = false
		}

		stored, err := v.store.GetStored()
		if err != nil {
			return "", fmt.Errorf("failed to get stored: %+v", err)
		}

		if len(stored) == 0 {
			break
		}

		for _, s := range stored {
			if s.Stream == string(b) {
				loop = true
				break
			}
			loop = false
		}
	}

	return string(b), nil
}
