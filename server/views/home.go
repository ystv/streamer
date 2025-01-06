package views

import (
	"log"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/server/templates"
)

// HomeFunc is the basic html writer that provides the main page for Streamer
func (v *Views) HomeFunc(c echo.Context) error {
	if v.conf.Verbose {
		log.Println("Home called")
	}

	_, rec := v.cache.Get(server.Recorder.String())
	_, fow := v.cache.Get(server.Forwarder.String())

	var err string

	switch {
	case !rec && !fow:
		err = "No recorder or forwarder available"
	case !rec:
		err = "No recorder available"
	case !fow:
		err = "No forwarder available"
	}

	data := struct {
		ActivePage string
		Error      string
	}{
		ActivePage: "home",
		Error:      err,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.MainTemplate)
}
