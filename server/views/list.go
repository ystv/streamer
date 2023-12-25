package views

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/templates"
)

// ListFunc lists all current streams that are registered in the database
func (v *Views) ListFunc(c echo.Context) error {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"list", http.StatusTemporaryRedirect)
		return
	}*/
	if c.Request().Method == "GET" {
		if v.conf.Verbose {
			log.Println("Stop GET called")
		}

		return v.template.RenderTemplate(c.Response().Writer, nil, templates.ListTemplate)
	} else if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Stop POST called")
		}

		streams, err := v.store.GetStreams()
		if err != nil {
			return fmt.Errorf("failed to get streams: %w", err)
		}

		var streamsSlice []string

		data := false

		for _, s := range streams {
			data = true
			streamsSlice = append(streamsSlice, "Active", "-", s.Stream, "-", s.Input)
			streamsSlice = append(streamsSlice, "<br>")
		}

		stored, err := v.store.GetStored()
		if err != nil {
			return fmt.Errorf("failed to get stored: %w", err)
		}

		for _, s := range stored {
			data = true
			streamsSlice = append(streamsSlice, "Saved", "-", s.Stream, "-", s.Input)
			streamsSlice = append(streamsSlice, "<br>")
		}

		if !data {
			return c.String(http.StatusOK, "No current streams")
		} else {
			stringByte := strings.Join(streamsSlice, "\x20")
			return c.String(http.StatusOK, stringByte)
		}
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
