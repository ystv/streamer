package views

import (
	"fmt"
	"log"
	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/templates"
	"net/http"
)

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

		unique := c.FormValue("unique")

		stored, err := v.store.FindStored(unique)
		if err != nil {
			return fmt.Errorf("failed to get stored: %w, unique: %s", err, unique)
		}

		if stored == nil {
			log.Println("No data")
			log.Printf("rejected recall: %s", unique)
			return c.String(http.StatusOK, "REJECTED!")
		}

		log.Printf("accepted recall: %s", unique)
		return c.String(http.StatusOK, "ACCEPTED!")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
