package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/templates"
)

// FacebookHelpFunc is the handler for the Facebook help page
func (v *Views) FacebookHelpFunc(c echo.Context) error {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"facebookhelp", http.StatusTemporaryRedirect)
		return
	}*/

	if v.conf.Verbose {
		fmt.Println("facebook called")
	}

	return v.template.RenderTemplate(c.Response().Writer, nil, templates.FacebookHelpTemplate)
}