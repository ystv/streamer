package views

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/templates"
)

// YoutubeHelpFunc is the handler for the YouTube help page
func (v *Views) YoutubeHelpFunc(c echo.Context) error {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"youtubehelp", http.StatusTemporaryRedirect)
		return
	}*/

	if v.conf.Verbose {
		log.Println("YouTube called")
	}

	data := struct {
		ActivePage string
	}{}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.YouTubeHelpTemplate)
}
