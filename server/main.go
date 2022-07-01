package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

type Web struct {
	mux *mux.Router
}

type RTMP struct {
	XMLName xml.Name `xml:"rtmp"`
	Server  Server   `xml:"server"`
}

type Server struct {
	XMLName      xml.Name      `xml:"server"`
	Applications []Application `xml:"application"`
}

type Application struct {
	XMLName xml.Name `xml:"application"`
	Name    string   `xml:"name"`
	Live    Live     `xml:"live"`
}

type Live struct {
	XMLName xml.Name `xml:"live"`
	Streams []Stream `xml:"stream"`
}

type Stream struct {
	XMLName xml.Name `xml:"stream"`
	Name    string   `xml:"name"`
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var verbose bool

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// The main initial function that is called and is the root for the website
func main() {
	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "./streamer") {
		if len(os.Args) > 2 {
			fmt.Println(string(rune(len(os.Args))))
			fmt.Println(os.Args)
			log.Fatalf("Arguments error")
		}
		for i := 0; i < len(os.Args)-1; i++ {
			os.Args[i] = os.Args[i+1]
		}
	} else {
		if len(os.Args) > 1 {
			fmt.Println(string(rune(len(os.Args))))
			fmt.Println(os.Args)
			log.Fatalf("Arguments error")
		}
	}
	if os.Args[0] == "-v" {
		verbose = true
	} else {
		verbose = false
	}
	web := Web{mux: mux.NewRouter()}
	web.mux.HandleFunc("/", web.home)
	web.mux.HandleFunc("/endpoints", web.endpoints)
	web.mux.HandleFunc("/streams", web.streams)
	web.mux.HandleFunc("/start", web.start)
	web.mux.HandleFunc("/resume", web.resume)
	web.mux.HandleFunc("/status", web.status)
	web.mux.HandleFunc("/stop", web.stop)
	web.mux.HandleFunc("/list", web.list)
	web.mux.HandleFunc("/youtubehelp", web.youtubeHelp)
	web.mux.HandleFunc("/facebookhelp", web.facebookHelp)
	web.mux.HandleFunc("/public/{id:[a-zA-Z0-9_.-]+}", web.public) // This handles all the public pages that the webpage can request, e.g. css, images and jquery

	fmt.Println("Server listening...")
	err := http.ListenAndServe(":8080", web.mux)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// This is a basic html writer that provides the main page for Streamer
func (web *Web) home(w http.ResponseWriter, _ *http.Request) {
	if verbose {
		fmt.Println("Home called")
	}
	tmpl := template.Must(template.ParseFiles("html/main.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func (web *Web) endpoints(w http.ResponseWriter, r *http.Request) {
	if verbose {
		fmt.Println("Endpoints called")
	}
	if r.Method == "POST" {
		if verbose {
			fmt.Println("Endpoints POST")
		}
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		response, err := http.Get(os.Getenv("STREAM_CHECKER"))
		if err != nil {
			fmt.Println(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(response.Body)

		buf := new(strings.Builder)
		_, err = io.Copy(buf, response.Body)
		if err != nil {
			fmt.Println(err)
		}

		streamPageContent := buf.String()

		var rtmp RTMP

		err = xml.Unmarshal([]byte(streamPageContent), &rtmp)
		if err != nil {
			fmt.Println(err)
		}

		var endpoints []string

		for i := 0; i < len(rtmp.Server.Applications); i++ {
			endpoints = append(endpoints, "endpoint~"+rtmp.Server.Applications[i].Name)
		}

		stringByte := strings.Join(endpoints, "\x20")
		_, err = w.Write([]byte(stringByte))
		if err != nil {
			fmt.Println(err)
		}
	}
}

// This collects the data from the rtmp stat page of nginx and produces a list of active streaming endpoints from given endpoints
func (web *Web) streams(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}

		err = godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		response, err := http.Get(os.Getenv("STREAM_CHECKER"))
		if err != nil {
			fmt.Println(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(response.Body)

		buf := new(strings.Builder)
		_, err = io.Copy(buf, response.Body)
		if err != nil {
			fmt.Println(err)
		}

		streamPageContent := buf.String()

		var rtmp RTMP

		err = xml.Unmarshal([]byte(streamPageContent), &rtmp)
		if err != nil {
			fmt.Println(err)
		}

		var endpoints []string

		for key := range r.Form {
			endpoint := strings.Split(key, "~")
			for i := 0; i < len(rtmp.Server.Applications); i++ {
				if rtmp.Server.Applications[i].Name == endpoint[1] {
					for j := 0; j < len(rtmp.Server.Applications[i].Live.Streams); j++ {
						endpoints = append(endpoints, endpoint[1]+"/"+rtmp.Server.Applications[i].Live.Streams[j].Name)
					}
				}
			}
		}

		if len(endpoints) != 0 {
			stringByte := strings.Join(endpoints, "\x20")
			_, err := w.Write([]byte(stringByte))
			if err != nil {
				fmt.Println(err)
			}
		} else {
			_, err := w.Write([]byte("No active streams with the current selection"))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

// This section is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder
func (web *Web) start(w http.ResponseWriter, r *http.Request) {
	errors := false
	var errorMessage string
	if r.Method == "POST" {
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
				if websiteCheck(r.FormValue("website_stream_endpoint")) {
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

					err = godotenv.Load()
					if err != nil {
						fmt.Printf("error loading .env file: %s", err)
						errors = true
						errorMessage = err.Error()
					} else {
						forwarder := os.Getenv("FORWARDER")
						recorder := os.Getenv("RECORDER")
						username := os.Getenv("USERNAME")
						password := os.Getenv("PASSWORD")
						transmissionLight := os.Getenv("TRANSMISSION_LIGHT")
						var wg sync.WaitGroup
						wg.Add(2)
						go func() {
							defer wg.Done()
							if r.FormValue("record") == "on" {
								recording = true
								client, session, err := connectToHost(username, password, recorder)
								if err != nil {
									fmt.Println("Error connecting to Recorder for start")
									fmt.Println(err)
									errors = true
									errorMessage = err.Error()
								} else {
									_, err = session.CombinedOutput(recorderStart)
									if err != nil {
										fmt.Println("Error executing on Recorder for start")
										fmt.Println(err)
										errors = true
										errorMessage = err.Error()
									} else {
										err := client.Close()
										if err != nil {
											fmt.Println(err)
											errors = true
											errorMessage = err.Error()
										}
									}
								}
							}
						}()
						go func() {
							defer wg.Done()
							client, session, err := connectToHost(username, password, forwarder)
							if err != nil {
								fmt.Println("Error connecting to Forwarder for start")
								fmt.Println(err)
								errors = true
								errorMessage = err.Error()
							} else {
								_, err = session.CombinedOutput(forwarderStart)
								if err != nil {
									fmt.Println("Error executing on Forwarder for start")
									fmt.Println(err)
									errors = true
									errorMessage = err.Error()
								} else {
									err = client.Close()
									if err != nil {
										fmt.Println(err)
										errors = true
										errorMessage = err.Error()
									}
								}
							}
						}()
						wg.Wait()

						if errors == false {
							_, _ = http.Get(transmissionLight + "transmission_on")

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

// If the user decides to return at a later date then they can, by inputting the unique code that they were given then they can go to the resume page and enter the code
func (web *Web) resume(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("html/resume.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			fmt.Println(err)
		}
	} else if r.Method == "POST" {
		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {

		}

		rows, err := db.Query("SELECT * FROM streams WHERE stream = ?", r.FormValue("unique"))
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

//
func (web *Web) status(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {
			rows, err := db.Query("SELECT recording, website, streams FROM streams WHERE stream = ?", r.FormValue("unique"))
			if err != nil {
				fmt.Println(err)
			}

			var recording, website bool
			var streams int

			data := false

			for rows.Next() {
				err = rows.Scan(&recording, &website, &streams)
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

			err = godotenv.Load()
			if err != nil {
				fmt.Printf("error loading .env file: %s", err)
			}
			forwarder := os.Getenv("FORWARDER")
			recorder := os.Getenv("RECORDER")
			username := os.Getenv("USERNAME")
			password := os.Getenv("PASSWORD")
			m := make(map[string]string)
			var wg sync.WaitGroup
			if data {
				if recording {
					wg.Add(2)
					go func() {
						defer wg.Done()
						client, session, err := connectToHost(username, password, recorder)
						if err != nil {
							fmt.Println("Error connecting to Recorder for status")
							fmt.Println(err)
						}
						dataOut, err := session.CombinedOutput("./recorder_status.sh " + r.FormValue("unique"))
						if err != nil {
							fmt.Println("Error executing on Recorder for status")
							fmt.Println(err)
						}
						err = client.Close()
						if err != nil {
							fmt.Println(err)
						}

						dataOut1 := string(dataOut)[:len(dataOut)-2]

						if len(dataOut1) > 0 {
							if strings.Contains(dataOut1, "frame=") {
								first := strings.Index(dataOut1, "frame=") - 1
								last := strings.LastIndex(dataOut1, "\r")
								dataOut1 = dataOut1[:last]
								last = strings.LastIndex(dataOut1, "\r") + 1
								m["recording"] = dataOut1[:first] + "\n" + dataOut1[last:]
							} else {
								m["recording"] = dataOut1
							}
						}

						fmt.Println("Recorder status success")
					}()
				} else {
					wg.Add(1)
				}
				go func() {
					defer wg.Done()
					client, session, err := connectToHost(username, password, forwarder)
					if err != nil {
						fmt.Println("Error connecting to Forwarder for status")
						fmt.Println(err)
					}
					dataOut, err := session.CombinedOutput("./forwarder_status " + strconv.FormatBool(website) + " " + strconv.Itoa(streams) + " " + r.FormValue("unique"))
					if err != nil {
						fmt.Println("Error executing on Forwarder for status")
						fmt.Println(err)
					}
					err = client.Close()
					if err != nil {
						fmt.Println(err)
					}

					dataOut1 := string(dataOut)[4 : len(dataOut)-2]

					dataOut2 := strings.Split(dataOut1, "\u0000")

					for _, dataOut3 := range dataOut2 {
						if len(dataOut3) > 0 {
							if strings.Contains(dataOut3, "frame=") {
								dataOut4 := strings.Split(dataOut3, "~:~")
								first := strings.Index(dataOut4[1], "frame=") - 1
								last := strings.LastIndex(dataOut4[1], "\r")
								dataOut4[1] = dataOut4[1][:last]
								last = strings.LastIndex(dataOut4[1], "\r") + 1
								m[strings.Trim(dataOut4[0], " ")] = dataOut4[1][:first] + "\n" + dataOut4[1][last:]
							} else {
								dataOut4 := strings.Split(dataOut3, "~:~")
								m[strings.Trim(dataOut4[0], " ")] = dataOut4[1]
							}
						}
					}

					fmt.Println("Forwarder status success")
				}()
				wg.Wait()
				jsonStr, err := json.Marshal(m)
				output := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(string(jsonStr[1:len(jsonStr)-1]), "\\n", "<br>"), "\"", ""), " , ", "<br><br><br>"), " ,", "<br><br><br>"), "<br>,", "<br><br>")
				_, err = w.Write([]byte(output))
				if err != nil {
					fmt.Println(err.Error())
				}
			} else {
				fmt.Println("ERROR DATA STATUS")
			}
		}
	}
}

// When the stream is finished then you can stop the stream by pressing the stop button and that would kill all the ffmpeg commands
func (web *Web) stop(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
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
			err = godotenv.Load()
			if err != nil {
				fmt.Printf("error loading .env file: %s", err)
			}
			forwarder := os.Getenv("FORWARDER")
			recorder := os.Getenv("RECORDER")
			username := os.Getenv("USERNAME")
			password := os.Getenv("PASSWORD")
			transmissionLight := os.Getenv("TRANSMISSION_LIGHT")
			var wg sync.WaitGroup
			if recording {
				wg.Add(2)
				go func() {
					defer wg.Done()
					client, session, err := connectToHost(username, password, recorder)
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
				client, session, err := connectToHost(username, password, forwarder)
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
			fmt.Println(existingStreamCheck())
			if !existingStreamCheck() {
				_, err := http.Get(transmissionLight + "rehearsal_transmission_off") // Output is ignored as it returns a 204 status and there's a weird bug with no content
				if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
					fmt.Println(err.Error())
				}
			}
		} else {

		}
	}
}

// This lists all current streams that are registered in the database
func (web *Web) list(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("html/list.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			fmt.Println(err)
		}
	} else if r.Method == "POST" {
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

// This is the handler for the YouTube help page
func (web *Web) youtubeHelp(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.ParseFiles("html/youtubeHelp.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
	}
}

// This is the handler for the Facebook help page
func (web *Web) facebookHelp(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.ParseFiles("html/facebookHelp.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
	}
}

// This is the handler for any public documents, for example, the style sheet or images
func (web *Web) public(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.ServeFile(w, r, "public/"+vars["id"])
}

// This checks if the website stream key is valid using software called COBRA
func websiteCheck(endpoint string) bool {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %s", err)
	}
	keyChecker := os.Getenv("KEY_CHECKER")
	data := url.Values{}
	data.Set("call", "publish")
	var splitting []string
	data.Set("app", "live")
	splitting = strings.Split(endpoint, "?pwd=")
	data.Set("name", splitting[0])
	data.Set("pwd", splitting[1])

	client := &http.Client{}
	r, err := http.NewRequest("POST", keyChecker, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println(err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(r)
	if err != nil {
		fmt.Println(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(res.Body)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	if string(body) == "Accepted" {
		return true
	} else {
		return false
	}
}

// This checks if there are any existing streams still registered in the database
func existingStreamCheck() bool {
	db, err := sql.Open("sqlite3", "db/streams.db")
	if err != nil {
		fmt.Println(err)
	} else {
		rows, err := db.Query("SELECT stream FROM streams")
		if err != nil {
			fmt.Println(err)
		}

		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}

		var stream string

		for rows.Next() {
			err = rows.Scan(&stream)
			if err != nil {
				fmt.Println(err)
			}
			err = rows.Close()
			if err != nil {
				fmt.Println(err)
			}
			err = db.Close()
			if err != nil {
				fmt.Println(err)
			}
			return true
		}
		err = rows.Close()
		if err != nil {
			fmt.Println(err)
		}
		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}
		return false
	}
	err = db.Close()
	if err != nil {
		fmt.Println(err)
	}
	return false
}

// This is a general function to ssh to a remote server, any code execution is handled outside this function
func connectToHost(user, password, host string) (*ssh.Client, *ssh.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		err := client.Close()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}

	return client, session, nil
}
