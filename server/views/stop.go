package views

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/action"
	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/common/wsMessages"
	"github.com/ystv/streamer/server/helper/tx"
)

// StopFunc is used when the stream is finished,
// then you can stop the stream by pressing the stop button, and that would kill all the ffmpeg commands
func (v *Views) StopFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Stop POST called")
		}

		var response struct {
			Stopped bool   `json:"stopped"`
			Error   string `json:"error"`
		}

		unique := c.FormValue("unique_code")
		if len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			response.Error = fmt.Sprintf("unique key invalid: %s", unique)
			return c.JSON(http.StatusOK, response)
		}

		stream, err := v.store.FindStream(unique)
		if err != nil {
			log.Printf("unable to find unique code for stop: %s, %+v", unique, err)
			response.Error = fmt.Sprintf("unable to find unique code for stop: %s, %+v", unique, err)
			return c.JSON(http.StatusOK, response)
		}

		if stream == nil {
			log.Printf("failed to get stream as data is empty")
			response.Error = "failed to get stream as data is empty"
			return c.JSON(http.StatusOK, response)
		}

		transporter := commonTransporter.Transporter{
			Action: action.Stop,
			Unique: unique,
		}

		_, rec := v.cache.Get(server.Recorder.String())
		_, fow := v.cache.Get(server.Forwarder.String())

		if (!rec && len(stream.Recording) > 0) && !fow {
			err = fmt.Errorf("no recorder or forwarder available")
		} else if !rec && len(stream.Recording) > 0 {
			err = fmt.Errorf("no recorder available")
		} else if !fow {
			err = fmt.Errorf("no forwarder available")
		}

		var wg sync.WaitGroup
		var errorMessages []string
		wg.Add(2)
		go func() {
			defer wg.Done()
			if len(stream.Recording) > 0 {
				recorderTransporter := transporter
				var wsResponse commonTransporter.ResponseTransporter
				wsResponse, err = v.wsHelper(server.Recorder, recorderTransporter)
				if err != nil {
					log.Printf("failed sending to Recorder for stop: %+v", err)
					errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Recorder for stop: %+v", err))
					return
				}
				if wsResponse.Status == wsMessages.Error {
					log.Printf("failed sending to Recorder for stop: %s", wsResponse.Payload)
					errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Recorder for stop: %s", wsResponse.Payload))
					return
				}
				if wsResponse.Status != wsMessages.Okay {
					log.Printf("invalid response from Recorder for stop: %s", wsResponse.Status)
					errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Recorder for start: %s", wsResponse.Status))
					return
				}

				log.Println("recorder stop success")
			}
		}()
		go func() {
			defer wg.Done()
			forwarderTransporter := transporter

			var wsResponse commonTransporter.ResponseTransporter
			wsResponse, err = v.wsHelper(server.Forwarder, forwarderTransporter)
			if err != nil {
				log.Printf("failed sending to Forwarder for stop: %+v", err)
				errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Forwarder for stop: %+v", err))
				return
			}
			if wsResponse.Status == wsMessages.Error {
				log.Printf("failed sending to Forwarder for stop: %s", wsResponse.Payload)
				errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Forwarder for stop: %s", wsResponse.Payload))
				return
			}
			if wsResponse.Status != wsMessages.Okay {
				log.Printf("invalid response from Forwarder for stop: %s", wsResponse.Status)
				errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Forwarder for stop: %s", wsResponse.Status))
				return
			}

			log.Println("forwarder stop success")
		}()
		wg.Wait()

		if len(errorMessages) == 0 {
			err = v.HandleTXLight(v.conf.TransmissionLight, tx.AllOff)
			if err != nil {
				log.Printf("failed to turn transmission light off: %+v, ignoring and continuing", err)
			}

			err = v.store.DeleteStream(unique)
			if err != nil {
				log.Printf("failed to delete stream: %+v, unique: %s", err, unique)
				errorMessages = append(errorMessages, fmt.Sprintf("failed to delete stream: %+v, unique: %s", err, unique))
				response.Error = strings.Join(errorMessages, ",")
				return c.JSON(http.StatusOK, response)
			}

			log.Printf("stopped stream: %s", unique)

			response.Stopped = true
			return c.JSON(http.StatusOK, response)
		}

		response.Error = strings.Join(errorMessages, ",")
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
