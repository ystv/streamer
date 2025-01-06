package views

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/templates"
)

// FacebookHelpFunc is the handler for the Facebook help page
func (v *Views) FacebookHelpFunc(c echo.Context) error {
	if v.conf.Verbose {
		log.Println("facebook called")
	}

	data := struct {
		ActivePage string
	}{}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.FacebookHelpTemplate)
}
