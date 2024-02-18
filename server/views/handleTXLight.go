package views

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ystv/streamer/server/helper/tx"
)

func (v *Views) HandleTXLight(url string, function tx.FunctionTX) error {
	var resp *http.Response
	var err error
	switch function {
	case tx.TransmissionOn:
		resp, err = http.Get(url + tx.TransmissionOn.String())
		if err != nil && !strings.Contains(err.Error(), "unexpected EOF") && (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent) {
			return fmt.Errorf("failed to get response from tx light transmission on: %w", err)
		}
	case tx.AllOff:
		if !v.ExistingStreamCheck() {
			resp, err = http.Get(url + tx.AllOff.String())
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") && (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent) {
				return fmt.Errorf("failed to get response from tx light all off: %w", err)
			}
		} else if !v.SavedStreamCheck() {
			resp, err = http.Get(url + tx.RehearsalOn.String())
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") && (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent) {
				return fmt.Errorf("failed to get response from tx light rehearsal on: %w", err)
			}
		}
	case tx.RehearsalOn:
		if !v.ActiveStreamCheck() {
			resp, err = http.Get(url + tx.RehearsalOn.String())
			log.Printf("response: %#v, error: %#v", resp, err)
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") && (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent) {
				return fmt.Errorf("failed to get response from tx light rehearsal on: %w", err)
			}
		}
	default:
		return fmt.Errorf("unexpected function string: \"%s\"", function)
	}

	return nil
}
