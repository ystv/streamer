package main

import (
	"database/sql"
	"fmt"
	"github.com/ystv/streamer/server/templates"
	"net/http"
	"strconv"
	"time"
)

// resume is used if the user decides to return at a later date then they can, by inputting the unique code that they were given then they can go to the resume page and enter the code
func (web *Web) resume(w http.ResponseWriter, r *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication, http.StatusTemporaryRedirect)
		return
	}*/
	if r.Method == "GET" {
		if verbose {
			fmt.Println("Resume GET called")
		}
		web.t = templates.NewResume()

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
			fmt.Println("Resume POST called")
		}
		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {

		}

		rows, err := db.Query("SELECT stream, recording, website, streams FROM streams WHERE stream = ?", r.FormValue("unique"))
		if err != nil {
			fmt.Println(err)
		}
		var stream string
		var recording, website bool
		var streams int

		data := false

		accepted := false

		for rows.Next() {
			err = rows.Scan(&stream, &recording, &website, &streams)
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
			_, err := w.Write([]byte("ACCEPTED!~" + strconv.FormatBool(recording) + "~" + strconv.FormatBool(website) + "~" + strconv.Itoa(streams)))
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
