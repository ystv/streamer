package views

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/templates"
)

type RecallStream struct {
	StreamServer string `json:"streamServer"`
	StreamKey    string `json:"streamKey"`
}

// RecallFunc can pull back up stream details from the save function and allows you to start a stored stream
func (v *Views) RecallFunc(c echo.Context) error {
	if c.Request().Method == "GET" {
		if v.conf.Verbose {
			log.Println("Recall GET called")
		}

		return v.template.RenderTemplate(c.Response().Writer, nil, templates.RecallTemplate)
	} else if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Recall POST called")
		}

		response := struct {
			Unique        string         `json:"unique"`
			InputEndpoint string         `json:"inputEndpoint"`
			InputStream   string         `json:"inputStream"`
			RecordingPath string         `json:"recordingPath,omitempty"`
			WebsiteStream string         `json:"websiteStream,omitempty"`
			Streams       []RecallStream `json:"streams"`
			Error         string         `json:"error"`
		}{}

		unique := c.FormValue("unique")
		if len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			response.Error = fmt.Sprintf("unique key invalid: %s", unique)
			return c.JSON(http.StatusOK, response)
		}

		stored, err := v.store.FindStored(unique)
		if err != nil {
			log.Printf("failed to get stored: %+v, unique: %s", err, unique)
			response.Error = fmt.Sprintf("failed to get stored: %+v, unique: %s", err, unique)
			return c.JSON(http.StatusOK, response)
		}

		if stored == nil {
			log.Printf("failed to get stored as data is empty")
			response.Error = "failed to get stored as data is empty"
			return c.JSON(http.StatusOK, response)
		}

		response.Unique = stored.Stream

		if len(stored.Recording) > 0 {
			response.RecordingPath = stored.Recording
		}

		if len(stored.Website) > 0 {
			response.WebsiteStream = stored.Website
		}

		inputPart := strings.Split(stored.Input, "/")
		if len(inputPart) != 2 {
			log.Printf("failed to get input stream string, invalid array size: %d, %+v", len(inputPart), inputPart)
			response.Error = fmt.Sprintf("failed to get input stream string, invalid array size: %d, %+v", len(inputPart), inputPart)
			return c.JSON(http.StatusOK, response)
		}
		response.InputEndpoint = inputPart[0]
		response.InputStream = inputPart[1]

		response.Streams = []RecallStream{}
		for _, stream := range stored.Streams {
			var recallStream RecallStream
			splitStream := strings.Split(stream, "|")
			if len(splitStream) != 2 {
				log.Printf("failed to get output stream string, invalid array size: %d, %+v", len(splitStream), splitStream)
				response.Error = fmt.Sprintf("failed to get output stream string, invalid array size: %d, %+v", len(splitStream), splitStream)
				return c.JSON(http.StatusOK, response)
			}
			recallStream.StreamServer = splitStream[0]
			recallStream.StreamKey = splitStream[1]
			response.Streams = append(response.Streams, recallStream)
		}

		log.Printf("accepted recall: %s", unique)
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
