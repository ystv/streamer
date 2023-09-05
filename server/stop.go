package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

// stop is used when the stream is finished then you can stop the stream by pressing the stop button and that would kill all the ffmpeg commands
func (web *Web) stop(w http.ResponseWriter, r *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"list", http.StatusTemporaryRedirect)
		return
	}*/
	if r.Method == "POST" {
		if verbose {
			fmt.Println("Stop POST called")
		}
		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {

		}

		rows, err := db.Query("SELECT recording FROM streams WHERE stream = ?", r.FormValue("unique"))
		if err != nil {
			fmt.Println(err)
		}
		var recording bool

		data := false

		for rows.Next() {
			err = rows.Scan(&recording)
			if err != nil {
				fmt.Println(err)
			}
			data = true
		}
		err = rows.Close()
		if err != nil {
			fmt.Println(err)
		}
		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}
		if data {
			var wg sync.WaitGroup
			if recording {
				wg.Add(2)
				go func() {
					defer wg.Done()

					stopCmd := "./recorder_stop " + r.FormValue("unique") + " | bash"
					_, err := RunCommandOnHost(web.cfg.Recorder, web.cfg.RecorderUsername, web.cfg.RecorderPassword, stopCmd)
					if err != nil {
						log.Printf("error running recorder stop: %s", err)
						return
					}
				}()
			} else {
				wg.Add(1)
			}
			go func() {
				defer wg.Done()

				stopCmd := "./forwarder_stop " + r.FormValue("unique") + " | bash"
				_, err := RunCommandOnHost(web.cfg.Forwarder, web.cfg.ForwarderUsername, web.cfg.ForwarderPassword, stopCmd)
				if err != nil {
					log.Printf("error running forwarder stop: %s", err)
					return
				}

				fmt.Println("Forwarder stop success")
			}()
			wg.Wait()
			fmt.Println("STOPPED!")

			db, err = sql.Open("sqlite3", "db/streams.db")
			if err != nil {
				fmt.Println(err)
			} else {
				stmt, err := db.Prepare("DELETE FROM streams WHERE stream = ?")
				if err != nil {
					fmt.Println(err, "DELETE")
				}

				res, err := stmt.Exec(r.FormValue("unique"))
				if err != nil {
					fmt.Println(err)
				}

				affect, err := res.RowsAffected()
				if err != nil {
					fmt.Println(err)
				}

				if affect != 0 {
					_, err = w.Write([]byte("STOPPED!"))
					if err != nil {
						fmt.Println(err)
					}
				}
			}
			err = db.Close()
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(existingStreamCheck())
			if !existingStreamCheck() {
				_, err := http.Get(web.cfg.TransmissionLight + "rehearsal_transmission_off") // Output is ignored as it returns a 204 status and there's a weird bug with no content
				if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
					fmt.Println(err.Error())
				}
			}
		} else {

		}
	}
}
