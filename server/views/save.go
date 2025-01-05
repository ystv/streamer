package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/helper/tx"
	"github.com/ystv/streamer/server/storage"
	"github.com/ystv/streamer/server/templates"
)

// SaveFunc allows for the functionality of saving a stream's details for later in order to make things easier for massive operations where you have multiple streams at once
func (v *Views) SaveFunc(c echo.Context) error {
	if c.Request().Method == "GET" {
		if v.conf.Verbose {
			log.Println("Save GET called")
		}

		data := struct {
			ActivePage string
		}{
			ActivePage: "save",
		}

		return v.template.RenderTemplate(c.Response().Writer, data, templates.SaveTemplate)
	} else if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Save POST called")
		}

		var response struct {
			Unique string `json:"unique"`
			Error  string `json:"error"`
		}

		formValues := v.startSaveValidationHelper(c, saveValidation)
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

		stored := &storage.Stored{
			Stream:    unique,
			Input:     formValues.Input,
			Recording: formValues.SavePath,
			Website:   formValues.WebsiteOut,
			Streams:   formValues.Streams,
		}

		s, err := v.store.AddStored(stored)
		if err != nil {
			log.Printf("failed to add stored: %+v, unique: %s", err, unique)
			response.Error = fmt.Sprintf("failed to add stored: %+v, unique: %s", err, unique)
			return c.JSON(http.StatusOK, response)
		}

		if s == nil {
			log.Printf("failed to add stored, stored is nil")
			response.Error = "failed to add stored, stored is nil"
			return c.JSON(http.StatusOK, response)
		}

		err = v.HandleTXLight(v.conf.TransmissionLight, tx.RehearsalOn)
		if err != nil {
			log.Printf("failed to turn transmission light on: %+v, ignoring and continuing", err)
		}

		log.Printf("saved stream: %s", unique)
		response.Unique = unique
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
