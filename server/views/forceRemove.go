package views

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/action"
	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/common/wsMessages"
)

func (v *Views) ForceRemoveFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		log.Printf("force remove called")

		var response struct {
			Error string `json:"error"`
		}

		unique := c.Param("unique")
		if len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			response.Error = "unique key invalid: " + unique
			return c.JSON(http.StatusOK, response)
		}

		stored := false

		_, err := v.store.FindStored(unique)
		if err != nil {
			log.Printf("force remove did not find stored with unique: %s, attempting streams", unique)
		} else {
			stored = true
			err = v.store.DeleteStored(unique)
			if err != nil {
				log.Printf("failed to delete stored: %+v, unique: %s", err, unique)
				response.Error = fmt.Sprintf("failed to delete stored: %+v, unique: %s", err, unique)
				return c.JSON(http.StatusInternalServerError, response)
			}
			log.Printf("force removed stored: %s", unique)
		}

		stream, err := v.store.FindStream(unique)
		if err != nil {
			log.Printf("force remove did not find stream with unique: %s", unique)
			if !stored {
				log.Printf("failing forced")
				response.Error = fmt.Sprintf("force did not find stream or stored with unique: %s, failing", unique)
				return c.JSON(http.StatusInternalServerError, response)
			}
		} else {
			errorString := ""

			err = v.store.DeleteStream(unique)
			if err != nil {
				log.Printf("failed to delete stream: %+v, unique: %s", err, unique)
				errorString += fmt.Sprintf("failed to delete stream: %+v, unique: %s", err, unique)
			}

			// Adding delay to ensure the watchdog has stopped checking if it did
			time.Sleep(500 * time.Millisecond)

			_, rec := v.cache.Get(server.Recorder.String())
			_, fow := v.cache.Get(server.Forwarder.String())

			transporter := commonTransporter.Transporter{
				Action: action.Stop,
				Unique: unique,
			}

			if len(stream.GetRecording()) > 0 && rec {
				recorderTransporter := transporter
				var wsResponse commonTransporter.ResponseTransporter
				wsResponse, err = v.wsHelper(server.Recorder, recorderTransporter)
				if err != nil {
					log.Printf("failed sending to Recorder for force stop: %+v", err)
					errorString += fmt.Sprintf("failed sending to Recorder for force stop: %+v", err)
				}
				if wsResponse.Status == wsMessages.Error {
					log.Printf("failed sending to Recorder for force stop: %s", wsResponse.Payload)
					errorString += fmt.Sprintf("failed sending to Recorder for force stop: %s", wsResponse.Payload)
				}
				if wsResponse.Status != wsMessages.Okay {
					log.Printf("invalid response from Recorder for force stop: %s", wsResponse.Status)
					errorString += fmt.Sprintf("invalid response from Recorder for force stop: %s", wsResponse.Status)
				}
			}
			if fow {
				forwarderTransporter := transporter

				var wsResponse commonTransporter.ResponseTransporter
				wsResponse, err = v.wsHelper(server.Forwarder, forwarderTransporter)
				if err != nil {
					log.Printf("failed sending to Forwarder for force stop: %+v", err)
					errorString += fmt.Sprintf("failed sending to Forwarder for force stop: %+v", err)
				}
				if wsResponse.Status == wsMessages.Error {
					log.Printf("failed sending to Forwarder for force stop: %s", wsResponse.Payload)
					errorString += fmt.Sprintf("failed sending to Forwarder for force stop: %s", wsResponse.Payload)
				}
				if wsResponse.Status != wsMessages.Okay {
					log.Printf("invalid response from Forwarder for force stop: %s", wsResponse.Status)
					errorString += fmt.Sprintf("invalid response from Forwarder for force stop: %s", wsResponse.Status)
				}
			}

			if len(errorString) > 0 {
				response.Error = errorString
				return c.JSON(http.StatusInternalServerError, response)
			}

			log.Printf("force removed stream: %s", unique)
		}
		return nil
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
