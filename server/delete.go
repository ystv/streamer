package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

// delete will delete the saved stream before it can start
func (web *Web) delete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if verbose {
			fmt.Println("Delete POST called")
		}
		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {

		}

		rows, err := db.Query("SELECT stream FROM stored WHERE stream = ?", r.FormValue("unique"))
		if err != nil {
			fmt.Println(err)
		}
		var stream string

		data := false

		accepted := false

		for rows.Next() {
			err = rows.Scan(&stream)
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
			fmt.Println("DELETED!")
			_, err := w.Write([]byte("DELETED!"))
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
