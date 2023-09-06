package views

import (
	"errors"
	"github.com/ystv/streamer/server/templates"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func (v *Views) CustomHTTPErrorHandler(err error, c echo.Context) {
	log.Print(err)
	var he *echo.HTTPError
	var status int
	if errors.As(err, &he) {
		status = he.Code
	} else {
		status = 500
	}
	c.Response().WriteHeader(status)
	data := struct {
		Code  int
		Error any
	}{
		Code:  status,
		Error: he.Message,
	}
	err1 := v.template.RenderTemplate(c.Response().Writer, data, templates.ErrorTemplate)
	if err1 != nil {
		log.Printf("failed to render error page: %+v", err1)
	}
}

func (v *Views) Error404(c echo.Context) error {
	return v.template.RenderTemplate(c.Response().Writer, nil, templates.NotFound404Template)
}