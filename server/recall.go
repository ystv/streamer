package main

import (
	"database/sql"
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
		fmt.Println(r.ParseForm())
		fmt.Println(r)
		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {

		}

		rows, err := db.Query("SELECT * FROM stored WHERE stream = ?", r.FormValue("unique"))
		if err != nil {
			fmt.Println(err)
		}
		var stream, input, recording, website, streams string

		data := false

		accepted := false

		for rows.Next() {
			err = rows.Scan(&stream, &input, &recording, &website, &streams)
			if err != nil {
				fmt.Println(err)
			}
			data = true
			if stream == r.FormValue("unique") {
				accepted = true
			}
		}

		if !data {
			fmt.Println("No data")
		}

		err = rows.Close()
		if err != nil {
			fmt.Println(err)
		}

		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}

		if accepted {
			fmt.Println("ACCEPTED!")
			_, err := w.Write([]byte("ACCEPTED!~" + input + "~" + recording + "~" + website + "~" + streams))
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("REJECTED!")
			_, err := w.Write([]byte("REJECTED!"))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
