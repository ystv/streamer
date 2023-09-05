package main

import (
	"fmt"
	"github.com/ystv/streamer/server/templates"
	"net/http"
	"time"
)

// facebookHelp is the handler for the Facebook help page
func (web *Web) facebookHelp(w http.ResponseWriter, _ *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"facebookhelp", http.StatusTemporaryRedirect)
		return
	}*/

	if verbose {
		fmt.Println("Facebook called")
	}

	web.t = templates.NewFacebookHelp()

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
