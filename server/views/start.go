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
	"github.com/ystv/streamer/server/storage"
)

// StartFunc is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder
func (v *Views) StartFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Start POST called")
		}

		var response struct {
			Unique string `json:"unique"`
			Error  string `json:"error"`
		}

		formValues := v.startSaveValidationHelper(c, Start)
		if formValues.Error != nil {
			log.Printf("invalid form input: %+v", formValues.Error)
			response.Error = fmt.Sprintf("invalid form input: %+v", formValues.Error)
			return c.JSON(http.StatusOK, response)
		}

		unique, err := v.generateUnique()
		if err != nil {
			log.Printf("failed to get unique: %+v", err)
			response.Error = fmt.Sprintf("failed to get unique: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		transporter := commonTransporter.Transporter{
			Action: action.Start,
			Unique: unique,
		}

		fStart := commonTransporter.ForwarderStart{
			StreamIn:   formValues.Input,
			WebsiteOut: formValues.WebsiteOut,
		}

		rStart := commonTransporter.RecorderStart{
			StreamIn: formValues.Input,
			PathOut:  formValues.SavePath,
		}

		fStart.Streams = formValues.Streams

		var wg sync.WaitGroup
		wg.Add(2)
		var errorMessages []string
		go func() {
			defer wg.Done()
			if formValues.RecordCheckbox {
				recorderTransporter := transporter
				recorderTransporter.Payload = rStart
				var wsResponse commonTransporter.ResponseTransporter
				wsResponse, err = v.wsHelper(server.Recorder, recorderTransporter)
				if err != nil {
					log.Printf("failed sending to Recorder for start: %+v", err)
					errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Recorder for start: %+v", err))
					return
				}
				if wsResponse.Status == wsMessages.Error {
					log.Printf("failed sending to Recorder for start: %s", wsResponse.Payload)
					errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Recorder for start: %s", wsResponse.Payload))
					return
				}
				if wsResponse.Status != wsMessages.Okay {
					log.Printf("invalid response from Recorder for start: %s", wsResponse.Status)
					errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Recorder for start: %s", wsResponse.Status))
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			forwarderTransporter := transporter
			forwarderTransporter.Payload = fStart
			var wsResponse commonTransporter.ResponseTransporter
			wsResponse, err = v.wsHelper(server.Forwarder, forwarderTransporter)
			if err != nil {
				log.Printf("failed sending to Forwarder for start: %+v", err)
				errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Forwarder for start: %+v", err))
				return
			}
			if wsResponse.Status == wsMessages.Error {
				log.Printf("failed sending to Forwarder for start: %s", response)
				errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Forwarder for start: %s", response))
				return
			}
			if wsResponse.Status != wsMessages.Okay {
				log.Printf("invalid response from Forwarder for start: %s", response)
				errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Forwarder for start: %s", response))
				return
			}
		}()
		wg.Wait()

		if len(errorMessages) == 0 {
			err = v.HandleTXLight(v.conf.TransmissionLight, tx.TransmissionOn)
			if err != nil {
				log.Printf("failed to turn transmission light on: %+v, ignoring and continuing", err)
			}

			var s *storage.Stream
			s, err = v.store.AddStream(&storage.Stream{
				Stream:    unique,
				Input:     formValues.Input,
				Recording: rStart.PathOut,
				Website:   fStart.WebsiteOut,
				Streams:   formValues.Streams,
			})
			if err != nil {
				log.Printf("invalid response from Forwarder for start: %s", response)
				errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Forwarder for start: %s", response))
				response.Error = strings.Join(errorMessages, ",")
				return c.JSON(http.StatusOK, response)
			}

			if s == nil {
				log.Printf("failed to add stream, data is empty")
				errorMessages = append(errorMessages, "failed to add stream, data is empty")
				response.Error = strings.Join(errorMessages, ",")
				return c.JSON(http.StatusOK, response)
			}

			log.Printf("started stream: %s", unique)

			response.Unique = unique
			return c.JSON(http.StatusOK, response)
		}

		errMsg := strings.Join(errorMessages, ",")
		log.Printf("failed to start: %+v", errMsg)
		response.Error = fmt.Sprintf("failed to start: %+v", errMsg)
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
