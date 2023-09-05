package main

import (
	"fmt"
	"github.com/ystv/streamer/server/templates"
	"net/http"
	"time"
)

// home is the basic html writer that provides the main page for Streamer
func (web *Web) home(w http.ResponseWriter, r *http.Request) {
	_ = r
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"authenticate1", http.StatusTemporaryRedirect)
		return
	}*/
	if verbose {
		fmt.Println("Home called")
	}
	/*tmpl := template.Must(template.ParseFiles("html/main.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
	}*/
	web.t = templates.NewMain()

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
