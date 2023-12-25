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

		return v.template.RenderTemplate(c.Response().Writer, nil, templates.ResumeTemplate)
	} else if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Resume POST called")
		}

		unique := c.FormValue("unique")

		stream, err := v.store.FindStream(unique)
		if err != nil {
			return fmt.Errorf("failed to get stream: %w, unique: %s", err, unique)
		}

		if stream == nil {
			log.Println("No data")
			log.Printf("rejected resume: %s", unique)
			return c.String(http.StatusOK, "REJECTED!")
		}

		log.Printf("accepted resume: %s", unique)
		return c.String(http.StatusOK, "ACCEPTED!")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
