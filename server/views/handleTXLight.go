package views

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ystv/streamer/server/helper/tx"
)

func (v *Views) HandleTXLight(url string, function tx.FunctionTX) error {
	var resp *http.Response
	var err error
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	switch function {
	case tx.TransmissionOn:
		resp, err = client.Get(url + tx.TransmissionOn.String())
		if err != nil && !strings.Contains(err.Error(), "unexpected EOF") /*&& (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent)*/ {
			return fmt.Errorf("failed to get response from tx light transmission on: %w", err)
		}
	case tx.AllOff:
		if !v.ExistingStreamCheck() {
			resp, err = client.Get(url + tx.AllOff.String())
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") /*&& (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent)*/ {
				return fmt.Errorf("failed to get response from tx light all off: %w", err)
			}
		} else if !v.SavedStreamCheck() {
			resp, err = client.Get(url + tx.RehearsalOn.String())
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") /*&& (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent)*/ {
				return fmt.Errorf("failed to get response from tx light rehearsal on: %w", err)
			}
		}
	case tx.RehearsalOn:
		if !v.ActiveStreamCheck() {
			resp, err = client.Get(url + tx.RehearsalOn.String())
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") /*&& (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent)*/ {
				return fmt.Errorf("failed to get response from tx light rehearsal on: %w", err)
			}
		}
	default:
		return fmt.Errorf("unexpected function string: \"%s\"", function)
	}

	_ = resp

	return nil
}
