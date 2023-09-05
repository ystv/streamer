package main

import (
	"fmt"
	"github.com/ystv/streamer/server/templates"
	"net/http"
	"time"
)

// save allows for the functionality of saving a stream's details for later in order to make things easier for massive operations where you have multiple streams at once
func (web *Web) save(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if verbose {
			fmt.Println("Save GET called")
		}
		web.t = templates.NewSave()

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
	} else if r.Method == "POST" {
		if verbose {
			fmt.Println("Save POST called")
		}
	}
}
