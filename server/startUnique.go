package main

import (
	"database/sql"
	"fmt"
	"github.com/ystv/streamer/server/helper"
	"github.com/ystv/streamer/server/helper/tx"
	"golang.org/x/crypto/ssh"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// startUnique is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder with a specified unique key
func (web *Web) startUnique(w http.ResponseWriter, r *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication, http.StatusTemporaryRedirect)
		return
	}*/
	//errors := false
	if r.Method == "POST" {
		if verbose {
			fmt.Println("StartUnique POST called")
		}
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}
		fmt.Println(r)
		unique := r.FormValue("unique_code")
		if len(unique) != 10 {
			errorFunc("Unique key invalid", w)
			return
		}

		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}

		recording := false
		websiteStream := false
		streams := 0
		var forwarderStart string
		if r.FormValue("website_stream") == "on" {
			websiteStream = true
			if web.websiteCheck(r.FormValue("website_stream_endpoint")) {
				forwarderStart = "./forwarder_start " + r.FormValue("stream_selector") + " " + r.FormValue("website_stream_endpoint") + " "
			} else {
				errorFunc("Website key check has failed", w)
				return
			}
		} else {
			forwarderStart = "./forwarder_start \"" + r.FormValue("stream_selector") + "\" no "
		}
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

		rows, err := db.Query("SELECT stream FROM stored")
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}
		var stream string

		for rows.Next() {
			err = rows.Scan(&stream)
			if err != nil {
				errorFunc(err.Error(), w)
				return
			}
			if stream == unique {
				break
			} else {
				errorFunc("Cannot find data in stored", w)
				return
			}
		}

		err = rows.Close()
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		stmt, err := db.Prepare("INSERT INTO streams(stream, recording, website, streams) values(?, false, false, 0)")
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		res, err := stmt.Exec(unique)
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		stmt, err = db.Prepare("DELETE FROM stored WHERE stream = ?")
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		res, err = stmt.Exec(unique)
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		affect, err := res.RowsAffected()
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		if affect == 0 {
			errorFunc("Failed to remove from stored", w)
			return
		}

		err = db.Close()
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		if id == 0 {
			errorFunc("id is 0!", w)
			return
		}

		forwarderStart += unique + " "
		for _, index := range numbers {
			server := r.FormValue("stream_server_" + strconv.Itoa(index))
			if server[len(server)-1] != '/' {
				server += "/"
			}
			forwarderStart += "\"" + server + "\" \"" + r.FormValue("stream_key_"+strconv.Itoa(index)) + "\" "
			streams++
		}
		forwarderStart += "| bash"

		recorderStart := "./recorder_start \"" + r.FormValue("stream_selector") + "\" \"" + r.FormValue("save_path") + "\" " + unique + " | bash"

		var wg sync.WaitGroup
		wg.Add(2)
		errors := false
		go func() {
			defer wg.Done()
			if r.FormValue("record") == "on" {
				recording = true
				var client *ssh.Client
				var session *ssh.Session
				var err error
				//if recorderAuth == "PEM" {
				//	client, session, err = connectToHostPEM(recorder, recorderUsername, recorderPrivateKey, recorderPassphrase)
				//} else if recorderAuth == "PASS" {
				client, session, err = helper.ConnectToHostPassword(web.cfg.Recorder, web.cfg.RecorderUsername, web.cfg.RecorderPassword, verbose)
				//}
				if err != nil {
					fmt.Println(err, "Error connecting to Recorder for start")
					errorFunc(err.Error()+" - Error connecting to Recorder for start", w)
					return
				}
				_, err = session.CombinedOutput(recorderStart)
				if err != nil {
					fmt.Println(err, "Error executing on Recorder for start")
					errorFunc(err.Error()+" - Error executing on Recorder for start", w)
					return
				}
				err = client.Close()
				if err != nil {
					fmt.Println(err)
					errorFunc(err.Error(), w)
					return
				}
			}
		}()
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
				fmt.Println(err, "Error connecting to Forwarder for start")
				errorFunc(err.Error()+" - Error connecting to Forwarder for start", w)
				errors = true
				return
			}
			_, err = session.CombinedOutput(forwarderStart)
			if err != nil {
				fmt.Println(err, "Error executing on Forwarder for start")
				errorFunc(err.Error()+" - Error executing on Forwarder for start", w)
				errors = true
				return
			}
			err = client.Close()
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				errors = true
				return
			}
		}()
		wg.Wait()

		if errors == false {
			err = helper.HandleTXLight(web.cfg.TransmissionLight, tx.TransmissionOn, verbose)
			if err != nil {
				fmt.Println(err)
			}

			db, err = sql.Open("sqlite3", "db/streams.db")
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}

			stmt, err := db.Prepare("UPDATE streams SET recording = ?, website = ?, streams = ? WHERE stream = ?")
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}

			res, err := stmt.Exec(recording, websiteStream, streams, unique)
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}

			id, err = res.LastInsertId()
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}

			err = db.Close()
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}

			_, err = w.Write([]byte(unique))
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}
		}
	}
}
