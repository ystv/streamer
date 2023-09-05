package main

import (
	"database/sql"
	"fmt"
	"github.com/ystv/streamer/server/helper"
	"github.com/ystv/streamer/server/helper/tx"
	"github.com/ystv/streamer/server/templates"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
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
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}

		endpoint := strings.Split(r.FormValue("endpointsTable"), "~")[1]

		largest := 0
		var numbers []int
		for s := range r.PostForm {
			if strings.Contains(s, "stream_server_") {
				split := strings.Split(s, "_")
				conv, _ := strconv.ParseInt(split[2], 10, 64)
				largest = int(math.Max(float64(largest), float64(conv)))
				numbers = append(numbers, int(conv))
			}
		}
		sort.Ints(numbers)

		var b []byte

		loop := true

		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}

		for loop {
			b = make([]byte, 10)
			for i := range b {
				b[i] = charset[seededRand.Intn(len(charset))]
			}

			rows, err := db.Query("SELECT stream FROM stored")
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}
			var stream string
			data := false

			for rows.Next() {
				err = rows.Scan(&stream)
				if err != nil {
					fmt.Println(err)
					errorFunc(err.Error(), w)
					return
				} else {
					data = true
					if stream == string(b) {
						loop = true
						break
					} else {
						loop = false
					}
				}
			}

			if !data {
				loop = false
			}

			err = rows.Close()
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}
		}

		recording := "§"
		website := "§"

		if r.FormValue("record") == "on" {
			recording = r.FormValue("save_path")
		}

		if r.FormValue("website_stream") == "on" {
			website = r.FormValue("website_stream_endpoint")
		}

		var streams []string
		for _, index := range numbers {
			server := r.FormValue("stream_server_" + strconv.Itoa(index))
			server += "|"
			streams = append(streams, server+r.FormValue("stream_key_"+strconv.Itoa(index)))
		}

		stmt, err := db.Prepare("INSERT INTO stored(stream, input, recording, website, streams) values(?, ?, ?, ?, ?)")
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}

		res, err := stmt.Exec(string(b), endpoint+"/"+r.FormValue("stream_input"), recording, website, strings.Join(streams, "±"))
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}

		err = helper.HandleTXLight(web.cfg.TransmissionLight, tx.RehearsalOn, verbose)
		if err != nil {
			fmt.Println(err)
		}

		err = db.Close()
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}

		if id == 0 {
			fmt.Println("id is 0!")
			errorFunc("id is 0!", w)
			return
		}

		_, err = w.Write(b)
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}
	}
}
