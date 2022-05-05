package main

import (
	"bytes"
	"database/sql"
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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Web struct {
	mux *mux.Router
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// The main initial function that is called and is the root for the website
func main() {
	web := Web{mux: mux.NewRouter()}
	web.mux.HandleFunc("/", web.home)
	web.mux.HandleFunc("/streams", web.streams)
	web.mux.HandleFunc("/start", web.start)
	web.mux.HandleFunc("/resume", web.resume)
	web.mux.HandleFunc("/stop", web.stop)
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
	tmpl := template.Must(template.ParseFiles("html/main.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
	}
}

// This collects the data from the rtmp stat page of nginx and produces a list of active streaming endpoints from given endpoints
func (web *Web) streams(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %s", err)
	}
	response, err := http.Get(os.Getenv("STREAM_CHECKER"))
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {

		}
	}(response.Body)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, response.Body)
	if err != nil {
		log.Fatal(err)
	}

	streamPageContent := buf.String()

	re := regexp.MustCompile("<application>(.|\n)*?</application>") // The beginning of the separation of the get request
	applications := re.FindAllString(streamPageContent, -1)

	var m map[string][]string

	m = make(map[string][]string)

	if applications == nil {
		fmt.Println("No matches.")
	} else {
		for _, application := range applications {
			re1 := regexp.MustCompile("<stream>(.|\n)*?</stream>")
			lives := re1.FindAllString(application, -1)
			re2 := regexp.MustCompile("<name>(.|\n)*?</name>")
			names := re2.FindAllString(application, 1)
			_ = names
			names[0] = names[0][6 : len(names[0])-7]
			m[names[0]] = lives
		}
	}

	if r.Method == "POST" {
		err = r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}
		var internalEndpoint, prodEndpoint, liveEndpoint, streamEndpoint bool
		var n map[string][]string
		n = make(map[string][]string)
		if r.FormValue("internal_endpoint") == "on" { // There are four different endpoints that can be streamed to, internal, prod, live and stream. The checkboxes can select which one they want
			internalEndpoint = true
			n["internal"] = m["internal"]
		}
		if r.FormValue("prod_endpoint") == "on" {
			prodEndpoint = true
			n["prod"] = m["prod"]
		}
		if r.FormValue("live_endpoint") == "on" {
			liveEndpoint = true
			n["live"] = m["live"]
		}
		if r.FormValue("stream_endpoint") == "on" {
			streamEndpoint = true
			n["stream"] = m["stream"]
		}
		var endpoints []string
		i := 0
		for key, values := range n {
			for _, value := range values {
				re := regexp.MustCompile("<name>(.|\n)*?</name>")
				names := re.FindAllString(value, 1)
				names[0] = names[0][6 : len(names[0])-7]
				endpoints = append(endpoints, key+"/"+names[0])
				i++
			}
		}
		fmt.Println("Endpoints:", "Internal:", internalEndpoint, "Prod:", prodEndpoint, "Live:", liveEndpoint, "Stream:", streamEndpoint)
		b := new(bytes.Buffer)
		for key, value := range n {
			_, err := fmt.Fprintf(b, "%s=%s\n", key, value)
			if err != nil {
				return
			}
		}
		if len(endpoints) != 0 {
			fmt.Println("Data")
			space := regexp.MustCompile(`\s+`)
			s := space.ReplaceAllString(b.String(), " ")
			_ = s
			stringByte := strings.Join(endpoints, "\x20")
			_, err := w.Write([]byte(stringByte))
			if err != nil {
				return
			}
		} else {
			fmt.Println("No data")
			_, err := w.Write([]byte("No active streams with the current selection"))
			if err != nil {
				return
			}
		}
	}
}

// This section is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder
func (web *Web) start(w http.ResponseWriter, r *http.Request) {
	errors := false
	var errorMessage string
	if r.Method == "POST" {
		fmt.Println(r.Body)
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			errorMessage = err.Error()
			errors = true
		} else {
			websiteValid := true
			var forwarderStart string
			if r.FormValue("website_stream") == "on" {
				if websiteCheck(r.FormValue("website_stream_endpoint")) {
					fmt.Println("Success")
					websiteValid = true
					forwarderStart = "./forwarder_start " + r.FormValue("stream_selector") + " " + r.FormValue("website_stream_endpoint") + " "
				} else {
					websiteValid = false
					fmt.Println("Failed")
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

					rows, err := db.Query("SELECT * FROM streams")
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

				stmt, err := db.Prepare("INSERT INTO streams(stream) values(?)")
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
						fmt.Println(r.FormValue("record"))
						if r.FormValue("record") == "on" {
							client, session, err := connectToHost(username, password, recorder)
							if err != nil {
								fmt.Println("Error connecting to Recorder")
								fmt.Println(err)
								errors = true
								errorMessage = err.Error()
							} else {
								_, err = session.CombinedOutput(recorderStart)
								if err != nil {
									fmt.Println("Error executing on Recorder")
									fmt.Println(err)
									errors = true
									errorMessage = err.Error()
								} else {
									err := client.Close()
									if err != nil {
										fmt.Println(err)
										errors = true
										errorMessage = err.Error()
									} else {
										fmt.Println("Recorder success")
									}
								}
							}
						}
						if !errors {
							client1, session1, err := connectToHost(username, password, forwarder)
							if err != nil {
								fmt.Println("Error connecting to Forwarder")
								fmt.Println(err)
								errors = true
								errorMessage = err.Error()
							} else {
								fmt.Println(forwarderStart)
								_, err = session1.CombinedOutput(forwarderStart)
								if err != nil {
									fmt.Println("Error executing on Forwarder")
									fmt.Println(err)
									errors = true
									errorMessage = err.Error()
								} else {
									err = client1.Close()
									if err != nil {
										fmt.Println(err)
										errors = true
										errorMessage = err.Error()
									} else {
										fmt.Println("Forwarder success")

										_, err := http.Get(transmissionLight + "transmission_on") // Output is ignored as it returns a 204 status and there's a weird bug with no content
										if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
											fmt.Println(err)
											errors = true
											errorMessage = err.Error()
										}

										fmt.Println("STARTED!")

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
			return
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

		rows, err := db.Query("SELECT * FROM streams")
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
			return
		}

		err = db.Close()
		if err != nil {
			return
		}

		if accepted {
			fmt.Println("ACCEPTED!")
			_, err := w.Write([]byte("ACCEPTED!"))
			if err != nil {
				return
			}
		} else {
			fmt.Println("REJECTED!")
			_, err := w.Write([]byte("REJECTED!"))
			if err != nil {
				return
			}
		}
	}
}

// When the stream is finished then you can stop the stream by pressing the stop button and that would kill all the ffmpeg commands
func (web *Web) stop(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}
		forwarder := os.Getenv("FORWARDER")
		recorder := os.Getenv("RECORDER")
		username := os.Getenv("USERNAME")
		password := os.Getenv("PASSWORD")
		transmissionLight := os.Getenv("TRANSMISSION_LIGHT")
		client2, session2, err := connectToHost(username, password, recorder)
		if err != nil {
			fmt.Println("Error connecting to Forwarder")
			panic(err)
		}
		_, err = session2.CombinedOutput("./recorder_stop " + r.FormValue("unique") + " | bash")
		if err != nil {
			fmt.Println("Error executing on Recorder")
			panic(err)
		}
		err = client2.Close()
		if err != nil {
			return
		}

		fmt.Println("Recorder success")

		client3, session3, err := connectToHost(username, password, forwarder)
		if err != nil {
			fmt.Println("Error connecting to Forwarder")
			panic(err)
		}
		_, err = session3.CombinedOutput("./forwarder_stop " + r.FormValue("unique") + " | bash")
		if err != nil {
			fmt.Println("Error executing on Forwarder")
			panic(err)
		}
		err = client3.Close()
		if err != nil {
			return
		}

		fmt.Println("Forwarder success")

		fmt.Println("STOPPED!")

		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {
			stmt, err := db.Prepare("delete from streams where stream=?")
			if err != nil {
				fmt.Println(err)
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
					return
				}
			}
		}

		fmt.Println("CLOSING STOP")
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

			/*if response.StatusCode != 204 {
			        fmt.Println("Transmission light error")
			}*/
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
		log.Fatal(err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(res.Body)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
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
		rows, err := db.Query("SELECT * FROM streams")
		if err != nil {
			fmt.Println(err)
		}

		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}

		var stream string

		fmt.Println(rows)

		for rows.Next() {
			err = rows.Scan(&stream)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(len(stream), " - ", stream)
			fmt.Println("CLOSING FOR")
			err = rows.Close()
			if err != nil {
				fmt.Println(err.Error())
			}
			err = db.Close()
			if err != nil {
				fmt.Println(err)
			}
			return true
		}
		fmt.Println("CLOSING AFTER FOR")
		err = rows.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}
		return false
	}
	fmt.Println("CLOSING ELSE")
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
