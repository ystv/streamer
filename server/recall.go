package main

import (
	"fmt"
	"github.com/ystv/streamer/server/templates"
	"net/http"
	"time"
)

// recall can pull back up stream details from the save function and allows you to start a stored stream
func (web *Web) recall(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if verbose {
			fmt.Println("Recall GET called")
		}
		web.t = templates.NewRecall()

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
			fmt.Println("Recall POST called")
		}
	}
}
