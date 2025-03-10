package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/templates"
)

// ResumeFunc is used if the user decides to return at a later date then they can, by inputting the unique code that they were given then they can go to the resume page and enter the code
func (v *Views) ResumeFunc(c echo.Context) error {
	if c.Request().Method == "GET" {
		if v.conf.Verbose {
			log.Println("Resume GET called")
		}

		data := struct {
			ActivePage string
		}{
			ActivePage: "resume",
		}

		return v.template.RenderTemplate(c.Response().Writer, data, templates.ResumeTemplate)
	} else if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Resume POST called")
		}

		unique := c.FormValue("unique")

		var response ResumeResponse

		stream, err := v.store.FindStream(unique)
		if err != nil {
			log.Printf("failed to get stream: %+v, unique: %s", err, unique)
			response.Error = fmt.Sprintf("failed to get stream: %+v, unique: %s", err, unique)
			return c.JSON(http.StatusOK, response)
		}

		if stream == nil {
			log.Println("No data")
			log.Printf("rejected resume: %s", unique)
			response.Error = "REJECTED!"
			return c.JSON(http.StatusOK, response)
		}

		log.Printf("accepted resume: %s", unique)

		response.Response = "ACCEPTED!"
		response.Website = len(stream.GetWebsite()) > 0
		response.Recording = len(stream.GetRecording()) > 0
		response.Streams = uint64(len(stream.GetStreams()))
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
