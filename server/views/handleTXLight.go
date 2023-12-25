package views

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ystv/streamer/server/helper/tx"
)

func (v *Views) HandleTXLight(url string, function tx.FunctionTX) error {
	switch function {
	case tx.TransmissionOn:
		_, err := http.Get(url + tx.TransmissionOn.String())
		if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
			return fmt.Errorf("failed to get response from tx light transmission on: %w", err)
		}
		return nil
	case tx.AllOff:
		if !v.ExistingStreamCheck() {
			_, err := http.Get(url + tx.AllOff.String()) // Output is ignored as it returns a 204 status
			// and there's a unique bug with no content
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
				return fmt.Errorf("failed to get response from tx light all off: %w", err)
			}
		} else if !v.SavedStreamCheck() {
			_, err := http.Get(url + tx.RehearsalOn.String())
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
				return fmt.Errorf("failed to get response from tx light rehearsal on: %w", err)
			}
		}
		return nil
	case tx.RehearsalOn:
		if !v.ActiveStreamCheck() {
			_, err := http.Get(url + tx.RehearsalOn.String())
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
				return fmt.Errorf("failed to get response from tx light rehearsal on: %w", err)
			}
		}
		return nil
	default:
		return fmt.Errorf("unexpected function string: \"%s\"", function)
	}
}
