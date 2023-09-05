package main

import (
	"fmt"
	"github.com/ystv/streamer/server/templates"
	"net/http"
	"time"
)

// youtubeHelp is the handler for the YouTube help page
func (web *Web) youtubeHelp(w http.ResponseWriter, _ *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"youtubehelp", http.StatusTemporaryRedirect)
		return
	}*/

	if verbose {
		fmt.Println("YouTube called")
	}

	web.t = templates.NewYouTubeHelp()

	params := templates.PageParams{
		Base: templates.BaseParams{
			SystemTime: time.Now(),
		},
	}

	err := web.t.Page(w, params)
	if err != nil {
		err = fmt.Errorf("failed to render dashboard: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
