package main

import (
	"database/sql"
	"fmt"
	"github.com/ystv/streamer/server/templates"
	"net/http"
	"strings"
	"time"
)

// list lists all current streams that are registered in the database
func (web *Web) list(w http.ResponseWriter, r *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"list", http.StatusTemporaryRedirect)
		return
	}*/
	if r.Method == "GET" {
		if verbose {
			fmt.Println("Stop GET called")
		}
		web.t = templates.NewList()

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
			fmt.Println("Stop POST called")
		}
		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {

		}

		rows, err := db.Query("SELECT stream FROM streams")
		if err != nil {
			fmt.Println(err)
		}
		var stream string

		var streams []string

		data := false

		for rows.Next() {
			err = rows.Scan(&stream)
			if err != nil {
				fmt.Println(err)
			}
			data = true
			streams = append(streams, stream)
		}

		err = rows.Close()
		if err != nil {
			fmt.Println(err)
		}

		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}

		if !data {
			_, err = w.Write([]byte("No current streams"))
			if err != nil {
				fmt.Println(err)
			}
		} else {
			stringByte := strings.Join(streams, "\x20")
			_, err = w.Write([]byte(stringByte))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
