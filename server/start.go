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

// start is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder
func (web *Web) start(w http.ResponseWriter, r *http.Request) {
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
			fmt.Println("Start POST called")
		}
		err := r.ParseForm()
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
				fmt.Println("Website key check has failed")
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

			rows, err := db.Query("SELECT stream FROM streams")
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

		stmt, err := db.Prepare("INSERT INTO streams(stream, recording, website, streams) values(?, false, false, 0)")
		if err != nil {
			fmt.Println(err)
			errorFunc(err.Error(), w)
			return
		}

		res, err := stmt.Exec(string(b))
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

		forwarderStart += string(b) + " "
		for _, index := range numbers {
			server := r.FormValue("stream_server_" + strconv.Itoa(index))
			if server[len(server)-1] != '/' {
				server += "/"
			}
			forwarderStart += "\"" + server + "\" \"" + r.FormValue("stream_key_"+strconv.Itoa(index)) + "\" "
			streams++
		}
		forwarderStart += "| bash"

		recorderStart := "./recorder_start \"" + r.FormValue("stream_selector") + "\" \"" + r.FormValue("save_path") + "\" " + string(b) + " | bash"

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
			stmt, err := db.Prepare("UPDATE streams SET input = ?, recording = ?, website = ?, streams = ? WHERE stream = ?")
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}

			res, err := stmt.Exec(r.FormValue("stream_selector"), recording, websiteStream, streams, string(b))
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

			_, err = w.Write(b)
			if err != nil {
				fmt.Println(err)
				errorFunc(err.Error(), w)
				return
			}
		}
	}
}
