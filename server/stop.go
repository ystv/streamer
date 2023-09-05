package main

import (
	"database/sql"
	"fmt"
	"github.com/ystv/streamer/server/helper"
	"github.com/ystv/streamer/server/helper/tx"
	"golang.org/x/crypto/ssh"
	"net/http"
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
					var client *ssh.Client
					var session *ssh.Session
					var err error
					//if recorderAuth == "PEM" {
					//	client, session, err = connectToHostPEM(recorder, recorderUsername, recorderPrivateKey, recorderPassphrase)
					//} else if recorderAuth == "PASS" {
					client, session, err = helper.ConnectToHostPassword(web.cfg.Recorder, web.cfg.RecorderUsername, web.cfg.RecorderPassword, verbose)
					//}
					if err != nil {
						fmt.Println("Error connecting to Recorder for stop")
						fmt.Println(err)
					}
					_, err = session.CombinedOutput("./recorder_stop " + r.FormValue("unique") + " | bash")
					if err != nil {
						fmt.Println("Error executing on Recorder for stop")
						fmt.Println(err)
					}
					err = client.Close()
					if err != nil {
						fmt.Println(err)
					}
				}()
			} else {
				wg.Add(1)
			}
			go func() {
				defer wg.Done()
				var client *ssh.Client
				var session *ssh.Session
				var err error
				//if forwarderAuth == "PEM" {
				//	client, session, err = connectToHostPEM(forwarder, forwarderUsername, forwarderPrivateKey, forwarderPassphrase)
				//} else if forwarderAuth == "PASS" {
				client, session, err = helper.ConnectToHostPassword(web.cfg.Forwarder, web.cfg.ForwarderUsername, web.cfg.ForwarderPassword, verbose)
				//}
				if err != nil {
					fmt.Println("Error connecting to Forwarder for stop")
					fmt.Println(err)
				}
				_, err = session.CombinedOutput("./forwarder_stop " + r.FormValue("unique") + " | bash")
				if err != nil {
					fmt.Println("Error executing on Forwarder for stop")
					fmt.Println(err)
				}
				err = client.Close()
				if err != nil {
					fmt.Println(err)
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

			err = helper.HandleTXLight(web.cfg.TransmissionLight, tx.AllOff, verbose)
			if err != nil {
				fmt.Println(err)
			}
		} else {

		}
	}
}
