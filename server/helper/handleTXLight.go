package helper

import (
	"fmt"
	"github.com/ystv/streamer/server/helper/tx"
	"net/http"
	"strings"
)

func HandleTXLight(url string, function tx.FunctionTX, verbose bool) (err error) {
	switch function {
	case tx.TransmissionOn:
		_, err = http.Get(url + tx.TransmissionOn.String())
		if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
			return
		}
		break
	case tx.AllOff:
		if !ExistingStreamCheck(verbose) {
			_, err = http.Get(url + tx.AllOff.String()) // Output is ignored as it returns a 204 status and there's a weird bug with no content
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
				return
			}
		} else if !SavedStreamCheck(verbose) {
			_, err = http.Get(url + tx.RehearsalOn.String())
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
				return
			}
		}
		break
	case tx.RehearsalOn:
		if !ActiveStreamCheck(verbose) {
			_, err = http.Get(url + tx.RehearsalOn.String())
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
				return
			}
		}
		break
	default:
		err = fmt.Errorf("unexpected function string: \"%s\"", function)
	}
	return
}
