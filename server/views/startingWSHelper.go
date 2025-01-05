package views

import (
	"errors"
	"fmt"
	"log"
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

func (v *Views) startingWSHelper(c echo.Context, unique string, startingType StartingType) error {
	var valType ValidationType

	switch startingType {
	case nonStoredStart:
		valType = startValidation
	case storedStart:
		valType = startUniqueValidation
	default:
		return fmt.Errorf("invalid starting type: %d", startingType)
	}

	formValues := v.startSaveValidationHelper(c, valType)
	if formValues.Error != nil {
		return fmt.Errorf("invalid form input: %w", formValues.Error)
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

	var err error

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
			log.Printf("failed sending to Forwarder for start: %s", wsResponse.Payload)
			errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Forwarder for start: %s", wsResponse.Payload))
			return
		}
		if wsResponse.Status != wsMessages.Okay {
			log.Printf("invalid response from Forwarder for start: %s", wsResponse.Status)
			errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Forwarder for start: %s", wsResponse.Status))
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
			return fmt.Errorf("failed to add stream: %w", err)
		}

		if s == nil {
			return errors.New("failed to add stream, data is empty")
		}

		if startingType == storedStart {
			err = v.store.DeleteStored(unique)
			if err != nil {
				return fmt.Errorf("failed to delete stored: %w, unique: %s", err, unique)
			}
		}

		return nil
	}
	errMsg := strings.Join(errorMessages, ",")
	return fmt.Errorf("failed to start: %+v", errMsg)
}
