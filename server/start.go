package main

import (
	"database/sql"
	"fmt"
	"log"
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
	errors := false
	var errorMessage string
	if r.Method == "POST" {
		if verbose {
			fmt.Println("Start POST called")
		}
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			errorMessage = err.Error()
			errors = true
		} else {
			recording := false
			websiteStream := false
			streams := 0
			websiteValid := true
			var forwarderStart string
			if r.FormValue("website_stream") == "on" {
				websiteStream = true
				if web.websiteCheck(r.FormValue("website_stream_endpoint")) {
					websiteValid = true
					forwarderStart = "./forwarder_start " + r.FormValue("stream_selector") + " " + r.FormValue("website_stream_endpoint") + " "
				} else {
					websiteValid = false
					errors = true
					errorMessage = "Website key check has failed"
				}
			} else {
				forwarderStart = "./forwarder_start \"" + r.FormValue("stream_selector") + "\" no "
			}
			if websiteValid {
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
					errors = true
					errorMessage = err.Error()
				}

				for loop {
					b = make([]byte, 10)
					for i := range b {
						b[i] = charset[seededRand.Intn(len(charset))]
					}

					rows, err := db.Query("SELECT stream FROM streams")
					if err != nil {
						fmt.Println(err)
						errors = true
						errorMessage = err.Error()
					}
					var stream string
					data := false

					for rows.Next() {
						err = rows.Scan(&stream)
						if err != nil {
							fmt.Println(err)
							errors = true
							errorMessage = err.Error()
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
						errors = true
						errorMessage = err.Error()
					}
				}

				stmt, err := db.Prepare("INSERT INTO streams(stream, recording, website, streams) values(?, false, false, 0)")
				if err != nil {
					fmt.Println(err)
					errors = true
					errorMessage = err.Error()
				}

				res, err := stmt.Exec(string(b))
				if err != nil {
					fmt.Println(err)
					errors = true
					errorMessage = err.Error()
				}

				id, err := res.LastInsertId()
				if err != nil {
					fmt.Println(err)
					errors = true
					errorMessage = err.Error()
				}

				err = db.Close()
				if err != nil {
					fmt.Println(err)
					errors = true
					errorMessage = err.Error()
				} else if id != 0 && !errors {
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
					if r.FormValue("record") == "on" {
						recording = true
						wg.Add(1)
						go func() {
							defer wg.Done()
							_, err := RunCommandOnHost(web.cfg.Recorder, web.cfg.RecorderUsername, web.cfg.RecorderPassword, recorderStart)
							if err != nil {
								log.Printf("Error starting recorder: %v", err)
								errors = true
								errorMessage += err.Error()
							}
						}()
					}
					wg.Add(1)
					go func() {
						defer wg.Done()
						_, err := RunCommandOnHost(web.cfg.Forwarder, web.cfg.ForwarderUsername, web.cfg.ForwarderPassword, forwarderStart)
						if err != nil {
							log.Printf("Error starting forwarder: %v", err)
							errors = true
							errorMessage += err.Error()
						}
					}()
					wg.Wait()

					if errors == false {
						_, _ = http.Get(web.cfg.TransmissionLight + "transmission_on")

						db, err = sql.Open("sqlite3", "db/streams.db")
						if err != nil {
							fmt.Println(err)
							errors = true
							errorMessage = err.Error()
						} else {
							stmt, err := db.Prepare("UPDATE streams SET recording = ?, website = ?, streams = ? WHERE stream = ?")
							if err != nil {
								fmt.Println(err)
								errors = true
								errorMessage = err.Error()
							}

							res, err := stmt.Exec(recording, websiteStream, streams, string(b))
							if err != nil {
								fmt.Println(err)
								errors = true
								errorMessage = err.Error()
							}

							id, err = res.LastInsertId()
							if err != nil {
								fmt.Println(err)
								errors = true
								errorMessage = err.Error()
							}

							err = db.Close()
							if err != nil {
								fmt.Println(err)
								errors = true
								errorMessage = err.Error()
							} else {
								_, err = w.Write(b)
								if err != nil {
									fmt.Println(err)
									errors = true
									errorMessage = err.Error()
								}
							}
						}
					}
				}
			} else {
				fmt.Println("Failed to authenticate website stream")
				errors = true
				errorMessage = "Failed to authenticate website stream"
			}
		}
	}
	if errors {
		fmt.Println("An error has occurred...\n" + errorMessage)
		_, err := w.Write([]byte("An error has occurred...\n" + errorMessage))
		if err != nil {
			fmt.Println(err)
		}
	}
}
