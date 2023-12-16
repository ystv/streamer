package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/templates"
	"net/http"
)

// RecallFunc can pull back up stream details from the save function and allows you to start a stored stream
func (v *Views) RecallFunc(c echo.Context) error {
	if c.Request().Method == "GET" {
		if v.conf.Verbose {
			fmt.Println("Recall GET called")
		}

		return v.template.RenderTemplate(c.Response().Writer, nil, templates.RecallTemplate)
	} else if c.Request().Method == "POST" {
		if v.conf.Verbose {
			fmt.Println("Recall POST called")
		}

		stored, err := v.store.FindStored(c.FormValue("unique"))
		if err != nil {
			return err
		}

		if stored == nil {
			fmt.Println("No data")
			fmt.Println("REJECTED!")
			return c.String(http.StatusOK, "REJECTED!")
		}

		fmt.Println("ACCEPTED!")
		return c.String(http.StatusOK, "ACCEPTED!")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
