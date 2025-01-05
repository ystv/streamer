package views

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ystv/streamer/server/helper/tx"
)

func (v *Views) HandleTXLight(url string, function tx.FunctionTX) error {
	var req *http.Request
	var resp *http.Response
	var err error
	ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	switch function {
	case tx.TransmissionOn:
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, url+tx.TransmissionOn.String(), bytes.NewReader([]byte{}))
		if err != nil {
			return fmt.Errorf("could not create request: %w", err)
		}
		resp, err = client.Do(req)
		if err != nil && !strings.Contains(err.Error(), "unexpected EOF") /*&& (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent)*/ {
			return fmt.Errorf("failed to get response from tx light transmission on: %w", err)
		} else if err != nil {
			return fmt.Errorf("failed to get response from tx light transmission on: %w", err)
		}
		defer resp.Body.Close()
	case tx.AllOff:
		if !v.ExistingStreamCheck() {
			req, err = http.NewRequestWithContext(ctx, http.MethodGet, url+tx.AllOff.String(), bytes.NewReader([]byte{}))
			if err != nil {
				return fmt.Errorf("could not create request: %w", err)
			}
			resp, err = client.Do(req)
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") /*&& (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent)*/ {
				return fmt.Errorf("failed to get response from tx light all off: %w", err)
			} else if err != nil {
				return fmt.Errorf("failed to get response from tx light all off: %w", err)
			}
			defer resp.Body.Close()
		} else if !v.SavedStreamCheck() {
			req, err = http.NewRequestWithContext(ctx, http.MethodGet, url+tx.RehearsalOn.String(), bytes.NewReader([]byte{}))
			if err != nil {
				return fmt.Errorf("could not create request: %w", err)
			}
			resp, err = client.Do(req)
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") /*&& (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent)*/ {
				return fmt.Errorf("failed to get response from tx light received on: %w", err)
			} else if err != nil {
				return fmt.Errorf("failed to get response from tx light received on: %w", err)
			}
			defer resp.Body.Close()
		}
	case tx.RehearsalOn:
		if !v.ActiveStreamCheck() {
			req, err = http.NewRequestWithContext(ctx, http.MethodGet, url+tx.RehearsalOn.String(), bytes.NewReader([]byte{}))
			if err != nil {
				return fmt.Errorf("could not create request: %w", err)
			}
			resp, err = client.Do(req)
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") /*&& (resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent)*/ {
				return fmt.Errorf("failed to get response from tx light received on: %w", err)
			} else if err != nil {
				return fmt.Errorf("failed to get response from tx light received on: %w", err)
			}
			defer resp.Body.Close()
		}
	default:
		return fmt.Errorf("unexpected function string: \"%s\"", function)
	}

	_ = req
	_ = resp

	return nil
}
