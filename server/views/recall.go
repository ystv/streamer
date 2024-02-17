package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/templates"
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

		errResponse := struct {
			Error string `json:"error"`
		}{}

		unique := c.FormValue("unique")
		if len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			errResponse.Error = fmt.Sprintf("unique key invalid: %s", unique)
			return c.JSON(http.StatusOK, errResponse)
		}

		stored, err := v.store.FindStored(unique)
		if err != nil {
			log.Printf("failed to get stored: %+v, unique: %s", err, unique)
			errResponse.Error = fmt.Sprintf("failed to get stored: %+v, unique: %s", err, unique)
			return c.JSON(http.StatusOK, errResponse)
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
