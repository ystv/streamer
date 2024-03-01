package views

import (
	"log"

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
		log.Println("Home called")
	}

	_, rec := v.cache.Get(server.Recorder.String())
	_, fow := v.cache.Get(server.Forwarder.String())

	var err string

	if !rec && !fow {
		err = "No recorder or forwarder available"
	} else if !rec {
		err = "No recorder available"
	} else if !fow {
		err = "No forwarder available"
	}

	data := struct {
		Error string
	}{
		Error: err,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.MainTemplate)
}
