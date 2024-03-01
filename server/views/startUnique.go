package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// StartUniqueFunc is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder with a specified unique key
func (v *Views) StartUniqueFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("StartUnique POST called")
		}

		var response struct {
			Unique string `json:"unique"`
			Error  string `json:"error"`
		}

		unique := c.FormValue("unique_code")
		if len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			response.Error = fmt.Sprintf("unique key invalid: %s", unique)
			return c.JSON(http.StatusOK, response)
		}

		stored, err := v.store.FindStored(unique)
		if err != nil {
			log.Printf("unable to find unique code for startUnique: %s, %+v", unique, err)
			response.Error = fmt.Sprintf("unable to find unique code for startUnique: %s, %+v", unique, err)
			return c.JSON(http.StatusOK, response)
		}

		if stored == nil {
			log.Printf("failed to get stored as data is empty")
			response.Error = "failed to get stored as data is empty"
			return c.JSON(http.StatusOK, response)
		}

		//formValues := v.startSaveValidationHelper(c, startUniqueValidation)
		//if formValues.Error != nil {
		//	log.Printf("invalid form input: %+v", formValues.Error)
		//	response.Error = fmt.Sprintf("invalid form input: %+v", formValues.Error)
		//	return c.JSON(http.StatusOK, response)
		//}
		//
		//transporter := commonTransporter.Transporter{
		//	Action: action.startValidation,
		//	Unique: unique,
		//}
		//
		//fStart := commonTransporter.ForwarderStart{
		//	StreamIn:   formValues.Input,
		//	WebsiteOut: formValues.WebsiteOut,
		//}
		//
		//rStart := commonTransporter.RecorderStart{
		//	StreamIn: formValues.Input,
		//	PathOut:  formValues.SavePath,
		//}
		//
		//fStart.Streams = formValues.Streams
		//
		//var wg sync.WaitGroup
		//wg.Add(2)
		//var errorMessages []string
		//go func() {
		//	defer wg.Done()
		//	if formValues.RecordCheckbox {
		//		recorderTransporter := transporter
		//		recorderTransporter.Payload = rStart
		//		var wsResponse commonTransporter.ResponseTransporter
		//		wsResponse, err = v.wsHelper(server.Recorder, recorderTransporter)
		//		if err != nil {
		//			log.Printf("failed sending to Recorder for start: %+v", err)
		//			errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Recorder for start: %+v", err))
		//			return
		//		}
		//		if wsResponse.Status == wsMessages.Error {
		//			log.Printf("failed sending to Recorder for start: %s", wsResponse.Payload)
		//			errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Recorder for start: %s", wsResponse.Payload))
		//			return
		//		}
		//		if wsResponse.Status != wsMessages.Okay {
		//			log.Printf("invalid response from Recorder for start: %s", wsResponse.Status)
		//			errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Recorder for start: %s", wsResponse.Status))
		//			return
		//		}
		//	}
		//}()
		//go func() {
		//	defer wg.Done()
		//	forwarderTransporter := transporter
		//	forwarderTransporter.Payload = fStart
		//	var wsResponse commonTransporter.ResponseTransporter
		//	wsResponse, err = v.wsHelper(server.Forwarder, forwarderTransporter)
		//	if err != nil {
		//		log.Printf("failed sending to Forwarder for start: %+v", err)
		//		errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Forwarder for start: %+v", err))
		//		return
		//	}
		//	if wsResponse.Status == wsMessages.Error {
		//		log.Printf("failed sending to Forwarder for start: %s", response)
		//		errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Forwarder for start: %s", response))
		//		return
		//	}
		//	if wsResponse.Status != wsMessages.Okay {
		//		log.Printf("invalid response from Forwarder for start: %s", response)
		//		errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Forwarder for start: %s", response))
		//		return
		//	}
		//}()
		//wg.Wait()
		//
		//if len(errorMessages) == 0 {
		//	err = v.HandleTXLight(v.conf.TransmissionLight, tx.TransmissionOn)
		//	if err != nil {
		//		log.Printf("failed to turn transmission light on: %+v, ignoring and continuing", err)
		//	}
		//
		//	var s *storage.Stream
		//	s, err = v.store.AddStream(&storage.Stream{
		//		Stream:    unique,
		//		Input:     formValues.Input,
		//		Recording: rStart.PathOut,
		//		Website:   fStart.WebsiteOut,
		//		Streams:   formValues.Streams,
		//	})
		//	if err != nil {
		//		log.Printf("invalid response from Forwarder for start: %s", response)
		//		errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Forwarder for start: %s", response))
		//		response.Error = strings.Join(errorMessages, ",")
		//		return c.JSON(http.StatusOK, response)
		//	}
		//
		//	if s == nil {
		//		log.Printf("failed to add stream, data is empty")
		//		errorMessages = append(errorMessages, "failed to add stream, data is empty")
		//		response.Error = strings.Join(errorMessages, ",")
		//		return c.JSON(http.StatusOK, response)
		//	}
		//
		//	err = v.store.DeleteStored(unique)
		//	if err != nil {
		//		log.Printf("failed to delete stored: %+v, unique: %s", err, unique)
		//		response.Error = fmt.Sprintf("failed to delete stored: %+v, unique: %s", err, unique)
		//		return c.JSON(http.StatusOK, response)
		//	}
		//
		//	log.Printf("started stream: %s", unique)
		//
		//	response.Unique = unique
		//	return c.JSON(http.StatusOK, response)
		//}
		//
		//errMsg := strings.Join(errorMessages, ",")
		//log.Printf("failed to start: %+v", errMsg)
		//response.Error = fmt.Sprintf("failed to start: %+v", errMsg)
		//return c.JSON(http.StatusOK, response)

		err = v.startingWSHelper(c, unique, storedStart)
		if err != nil {
			log.Printf("failed to start: %+v", err)
			response.Error = fmt.Sprintf("failed to start: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		log.Printf("started stream: %s", unique)

		response.Unique = unique
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
