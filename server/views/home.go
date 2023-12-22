package views

import (
	"fmt"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/server/templates"
)

// HomeFunc is the basic html writer that provides the main page for Streamer
func (v *Views) HomeFunc(c echo.Context) error {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"authenticate1", http.StatusTemporaryRedirect)
		return
	}*/
	if v.conf.Verbose {
		fmt.Println("Home called")
	}

	_, rec := v.cache.Get(server.Recorder.String())
	_, fow := v.cache.Get(server.Forwarder.String())

	var err error

	if !rec && !fow {
		err = fmt.Errorf("no recorder or forwarder available")
	} else if !rec {
		err = fmt.Errorf("no recorder available")
	} else if !fow {
		err = fmt.Errorf("no forwarder available")
	}

	data := struct {
		Error error
	}{
		Error: err,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.MainTemplate)
}
