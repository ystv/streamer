package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
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

	return v.template.RenderTemplate(c.Response().Writer, nil, templates.MainTemplate)
}
