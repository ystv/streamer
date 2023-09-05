package main

import (
	//"crypto/x509"
	"database/sql"
	"encoding/json"
	//"encoding/pem"
	"encoding/xml"
	//"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ystv/streamer/server/templates"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/ssh"
)

type (
	Web struct {
		mux *mux.Router
		t   *templates.Templater
		cfg Config
	}

	Config struct {
		Forwarder         string `envconfig:"FORWARDER"`
		Recorder          string `envconfig:"RECORDER"`
		ForwarderUsername string `envconfig:"FORWARDER_USERNAME"`
		RecorderUsername  string `envconfig:"RECORDER_USERNAME"`
		ForwarderPassword string `envconfig:"FORWARDER_PASSWORD"`
		RecorderPassword  string `envconfig:"RECORDER_PASSWORD"`
		StreamChecker     string `envconfig:"STREAM_CHECKER"`
		TransmissionLight string `envconfig:"TRANSMISSION_LIGHT"`
		KeyChecker        string `envconfig:"KEY_CHECKER"`
		ServerPort        int    `envconfig:"SERVER_PORT"`
	}

	RTMP struct {
		XMLName xml.Name `xml:"rtmp"`
		Server  Server   `xml:"server"`
	}

	Server struct {
		XMLName      xml.Name      `xml:"server"`
		Applications []Application `xml:"application"`
	}

	Application struct {
		XMLName xml.Name `xml:"application"`
		Name    string   `xml:"name"`
		Live    Live     `xml:"live"`
	}

	Live struct {
		XMLName xml.Name `xml:"live"`
		Streams []Stream `xml:"stream"`
	}

	Stream struct {
		XMLName xml.Name `xml:"stream"`
		Name    string   `xml:"name"`
	}

	/*Claims struct {
		Id    int          `json:"id"`
		Perms []Permission `json:"perms"`
		Exp   int64        `json:"exp"`
		jwt.StandardClaims
	}*/

	/*Permission struct {
		Permission string `json:"perms"`
		jwt.StandardClaims
	}*/

	/*Views struct {
		cookie *sessions.CookieStore
	}*/
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	verbose    bool
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// main function is the start and the root for the website
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

	err := godotenv.Load()
	if err != nil {
		log.Printf("error loading .env file: %s", err)
	}

	var cfg Config
	err = envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("failed to process env vars: %s", err)
	}

	web := Web{
		mux: mux.NewRouter(),
		cfg: cfg,
	}
	//web.mux.HandleFunc("/authenticate1", web.authenticate1)
	web.mux.HandleFunc("/", web.home)                              // Default view
	web.mux.HandleFunc("/endpoints", web.endpoints)                // Call made by home to view endpoints
	web.mux.HandleFunc("/streams", web.streams)                    // Call made by home to view all active streams for the endpoints
	web.mux.HandleFunc("/start", web.start)                        // Call made by home to start forwarding
	web.mux.HandleFunc("/resume", web.resume)                      // To return to the page that controls a stream
	web.mux.HandleFunc("/status", web.status)                      // Call made by home to view status
	web.mux.HandleFunc("/stop", web.stop)                          // Call made by home to stop forwarding
	web.mux.HandleFunc("/list", web.list)                          // List view of current forwards
	web.mux.HandleFunc("/save", web.save)                          // Where you can save a stream for later
	web.mux.HandleFunc("/recall", web.recall)                      // Where you can recall a saved stream to modify it if needed and start it
	web.mux.HandleFunc("/delete", web.delete)                      // Deletes the saved stream if it is no longer needed
	web.mux.HandleFunc("/startUnique", web.startUnique)            // Call made by home to start forwarding from a recalled stream
	web.mux.HandleFunc("/youtubehelp", web.youtubeHelp)            // YouTube help page
	web.mux.HandleFunc("/facebookhelp", web.facebookHelp)          // Facebook help page
	web.mux.HandleFunc("/public/{id:[a-zA-Z0-9_.-]+}", web.public) // This handles all the public pages that the webpage can request, e.g. css, images and jquery

	fmt.Println("Server listening on port", web.cfg.ServerPort, "...")

	err = http.ListenAndServe(net.JoinHostPort("", strconv.Itoa(web.cfg.ServerPort)), web.mux)

	if err != nil {
		fmt.Println(err)
		return
	}
}

//
/*func authenticate(w http.ResponseWriter, r *http.Request) bool {
	_ = w
	response, err := http.Get("https://auth.dev.ystv.co.uk/api/set_token")
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
	fmt.Println(buf.String())
	reqToken := r.Header.Get("Authorization")
	//splitToken := strings.Split(reqToken, "Bearer ")
	//reqToken = splitToken[1]
	fmt.Println("Token - ", reqToken)
	err = godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %s", err)
	}

	sess := session.Get(r)
	if sess == nil {
		fmt.Println("None")
	} else {
		fmt.Println(sess)
	}

	jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

	http.Redirect(w, r, jwtAuthentication+"authenticate1", http.StatusTemporaryRedirect)
	return false
}

//
func (web *Web) authenticate1(w http.ResponseWriter, r *http.Request) {
	_ = w
	reqToken := r.Header.Get("Authorization")
	//splitToken := strings.Split(reqToken, "Bearer ")
	//reqToken = splitToken[1]
	fmt.Println("Token - ", reqToken)
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("error loading .env file: %s", err)
	}
	jwtKey := os.Getenv("JWT_KEY")

	//fmt.Println(r.Cookies())

	fmt.Println(r)

	view := Views{}

	view.cookie = sessions.NewCookieStore(
		[]byte("444bd23239f14b804af0ae40375c8feec80b699684f4d1a6d86f59658edb3706caaa306fd3361e6353bf54c0df66adb7c1e395cac79a72ee0339dc1892fd478e"),
		[]byte("444bd23239f14b804af0ae40375c8feec80b699684f4d1a6d86f59658edb3706"),
	)

	sess, err := view.cookie.Get(r, "session")

	fmt.Println(sess)

	_ = sess

	//store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

	//session, err := store.Get(r, "session-name")

	//fmt.Println(err)

	//fmt.Println(session.ID)

	//fmt.Println(session)

	//fmt.Println(session.Values["token"])

	response, err := http.Get("https://auth.dev.ystv.co.uk/api/set_token")
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

	tokenPage := buf.String()

	_ = tokenPage

	//fmt.Println(tokenPage)

	fmt.Println(r.Cookie("session"))

	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			fmt.Println(err)

		}
		fmt.Println(err)
		return
	}

	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(c.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	fmt.Println(tkn)
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			fmt.Println("Unauthorised")
			fmt.Println(err)
			return
		}
		fmt.Println(err)
		return
	}
	if !tkn.Valid {
		fmt.Println("Unauthorised")
		return
	}
	if time.Now().Unix() > claims.Exp {
		fmt.Println("Expired")
		return
	}
	for _, perm := range claims.Perms {
		if perm.Permission == "Streamer" {
			fmt.Println("~~~Success~~~")
			return
		}
	}
	fmt.Println("Unauthorised")
	return
}*/

// home is the basic html writer that provides the main page for Streamer
func (web *Web) home(w http.ResponseWriter, r *http.Request) {
	_ = r
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"authenticate1", http.StatusTemporaryRedirect)
		return
	}*/
	if verbose {
		fmt.Println("Home called")
	}
	/*tmpl := template.Must(template.ParseFiles("html/main.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
	}*/
	web.t = templates.NewMain()

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
}

// endpoints presents the endpoints to the user
func (web *Web) endpoints(w http.ResponseWriter, r *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication, http.StatusTemporaryRedirect)
		return
	}*/
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

		streamPageContent, err := getBody(web.cfg.StreamChecker)
		if err != nil {
			fmt.Println(err)
		}

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

// streams collects the data from the rtmp stat page of nginx and produces a list of active streaming endpoints from given endpoints
func (web *Web) streams(w http.ResponseWriter, r *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication, http.StatusTemporaryRedirect)
		return
	}*/
	if r.Method == "POST" {
		if verbose {
			fmt.Println("Streams POST called")
		}
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}

		err = godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		streamPageContent, err := getBody(web.cfg.StreamChecker)
		if err != nil {
			fmt.Println(err)
		}

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
				client, session, err = connectToHostPassword(web.cfg.Recorder, web.cfg.RecorderUsername, web.cfg.RecorderPassword)
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
			client, session, err = connectToHostPassword(web.cfg.Forwarder, web.cfg.ForwarderUsername, web.cfg.ForwarderPassword)
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
			err = handleTXLight(web.cfg.TransmissionLight, "start")
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

// resume is used if the user decides to return at a later date then they can, by inputting the unique code that they were given then they can go to the resume page and enter the code
func (web *Web) resume(w http.ResponseWriter, r *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication, http.StatusTemporaryRedirect)
		return
	}*/
	if r.Method == "GET" {
		if verbose {
			fmt.Println("Resume GET called")
		}
		web.t = templates.NewResume()

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
			fmt.Println("Resume POST called")
		}
		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {

		}

		rows, err := db.Query("SELECT stream, recording, website, streams FROM streams WHERE stream = ?", r.FormValue("unique"))
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

// status is used to check the status of the streams and does this by tail command of the output logs
func (web *Web) status(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if verbose {
			fmt.Println("Status POST called")
		}
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

			m := make(map[string]string)
			var wg sync.WaitGroup
			if data {
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
						client, session, err = connectToHostPassword(web.cfg.Recorder, web.cfg.RecorderUsername, web.cfg.RecorderPassword)
						//}
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
					var client *ssh.Client
					var session *ssh.Session
					var err error
					//if forwarderAuth == "PEM" {
					//	client, session, err = connectToHostPEM(forwarder, forwarderUsername, forwarderPrivateKey, forwarderPassphrase)
					//} else if forwarderAuth == "PASS" {
					client, session, err = connectToHostPassword(web.cfg.Forwarder, web.cfg.ForwarderUsername, web.cfg.ForwarderPassword)
					//}
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
					client, session, err = connectToHostPassword(web.cfg.Recorder, web.cfg.RecorderUsername, web.cfg.RecorderPassword)
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
				client, session, err = connectToHostPassword(web.cfg.Forwarder, web.cfg.ForwarderUsername, web.cfg.ForwarderPassword)
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

			err = handleTXLight(web.cfg.TransmissionLight, "stop")
			if err != nil {
				fmt.Println(err)
			}
		} else {

		}
	}
}

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
			errorFunc(err.Error(), w)
			return
		}

		rows, err := db.Query("SELECT stream, input FROM streams")
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}
		var stream, input string

		var streams []string

		data := false

		for rows.Next() {
			err = rows.Scan(&stream, &input)
			if err != nil {
				errorFunc(err.Error(), w)
				return
			}
			data = true
			streams = append(streams, "Active", "-", stream, "-", input)
			streams = append(streams, "<br>")
		}

		err = rows.Close()
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		rows, err = db.Query("SELECT stream, input FROM stored")
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		for rows.Next() {
			err = rows.Scan(&stream, &input)
			if err != nil {
				errorFunc(err.Error(), w)
				return
			}
			data = true
			streams = append(streams, "Saved", "-", stream, "-", input)
			streams = append(streams, "<br>")
		}

		err = db.Close()
		if err != nil {
			errorFunc(err.Error(), w)
			return
		}

		if !data {
			_, err = w.Write([]byte("No current streams"))
			if err != nil {
				errorFunc(err.Error(), w)
				return
			}
		} else {
			stringByte := strings.Join(streams, "\x20")
			_, err = w.Write([]byte(stringByte))
			if err != nil {
				errorFunc(err.Error(), w)
				return
			}
		}
	}
}

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

		err = handleTXLight(web.cfg.TransmissionLight, "save")
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

// recall can pull back up stream details from the save function and allows you to start a stored stream
func (web *Web) recall(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if verbose {
			fmt.Println("Recall GET called")
		}
		web.t = templates.NewRecall()

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
			fmt.Println("Recall POST called")
		}
		fmt.Println(r.ParseForm())
		fmt.Println(r)
		db, err := sql.Open("sqlite3", "db/streams.db")
		if err != nil {
			fmt.Println(err)
		} else {

		}

		rows, err := db.Query("SELECT * FROM stored WHERE stream = ?", r.FormValue("unique"))
		if err != nil {
			fmt.Println(err)
		}
		var stream, input, recording, website, streams string

		data := false

		accepted := false

		for rows.Next() {
			err = rows.Scan(&stream, &input, &recording, &website, &streams)
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
			_, err := w.Write([]byte("ACCEPTED!~" + input + "~" + recording + "~" + website + "~" + streams))
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
				client, session, err = connectToHostPassword(web.cfg.Recorder, web.cfg.RecorderUsername, web.cfg.RecorderPassword)
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
			client, session, err = connectToHostPassword(web.cfg.Forwarder, web.cfg.ForwarderUsername, web.cfg.ForwarderPassword)
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
			err = handleTXLight(web.cfg.TransmissionLight, "start")
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

// youtubeHelp is the handler for the YouTube help page
func (web *Web) youtubeHelp(w http.ResponseWriter, _ *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"youtubehelp", http.StatusTemporaryRedirect)
		return
	}*/

	if verbose {
		fmt.Println("YouTube called")
	}

	web.t = templates.NewYouTubeHelp()

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
}

// facebookHelp is the handler for the Facebook help page
func (web *Web) facebookHelp(w http.ResponseWriter, _ *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"facebookhelp", http.StatusTemporaryRedirect)
		return
	}*/

	if verbose {
		fmt.Println("Facebook called")
	}

	web.t = templates.NewFacebookHelp()

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
}

// public is the handler for any public documents, for example, the style sheet or images
func (web *Web) public(w http.ResponseWriter, r *http.Request) {
	if verbose {
		fmt.Println("Public called")
	}
	vars := mux.Vars(r)
	http.ServeFile(w, r, "public/"+vars["id"])
}

// websiteCheck checks if the website stream key is valid using software called COBRA
func (web *Web) websiteCheck(endpoint string) bool {
	if verbose {
		fmt.Println("Website Check called")
	}

	data := url.Values{}
	data.Set("call", "publish")
	var splitting []string
	data.Set("app", "live")
	splitting = strings.Split(endpoint, "?pwd=")
	data.Set("name", splitting[0])
	data.Set("pwd", splitting[1])

	client := &http.Client{}
	r, err := http.NewRequest("POST", web.cfg.KeyChecker, strings.NewReader(data.Encode()))
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

func getBody(url string) (body string, err error) {
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, response.Body)
	if err != nil {
		return
	}

	body = buf.String()

	return
}

func errorFunc(errs string, w http.ResponseWriter) {
	fmt.Println("An error has occurred...\n" + errs)
	_, err := w.Write([]byte("An error has occurred...\n" + errs))
	if err != nil {
		fmt.Println(err)
	}
}

// existingStreamCheck checks if there are any existing streams still registered in the database
func existingStreamCheck() bool {
	if verbose {
		fmt.Println("Existing Stream Check called")
	}
	db, err := sql.Open("sqlite3", "db/streams.db")
	if err != nil {
		fmt.Println(err)
	} else {
		rows, err := db.Query("SELECT stream FROM (SELECT stream FROM streams UNION ALL SELECT stream FROM stored)")
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

// savedStreamCheck checks if there are any existing streams still registered in the database
func savedStreamCheck() bool {
	if verbose {
		fmt.Println("Saved Stream Check called")
	}
	db, err := sql.Open("sqlite3", "db/streams.db")
	if err != nil {
		fmt.Println(err)
	} else {
		rows, err := db.Query("SELECT stream FROM stored")
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

// activeStreamCheck checks if there are any existing streams still registered in the database
func activeStreamCheck() bool {
	if verbose {
		fmt.Println("Active Stream Check called")
	}
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

// connectToHostPassword is a general function to ssh to a remote server, any code execution is handled outside this function
func connectToHostPassword(host, username, password string) (*ssh.Client, *ssh.Session, error) {
	if verbose {
		fmt.Println("Connect To Host Password called")
	}
	sshConfig := &ssh.ClientConfig{
		User: username,
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

func handleTXLight(url, function string) (err error) {
	switch function {
	case "start":
		_, err = http.Get(url + "transmission_on")
		if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
			return
		}
		break
	case "stop":
		if !existingStreamCheck() {
			_, err = http.Get(url + "rehearsal_transmission_off") // Output is ignored as it returns a 204 status and there's a weird bug with no content
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
				return
			}
		} else if !savedStreamCheck() {
			_, err = http.Get(url + "rehearsal_on")
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
				return
			}
		}
		break
	case "save":
		if !activeStreamCheck() {
			_, err = http.Get(url + "rehearsal_on")
			if err != nil && !strings.Contains(err.Error(), "unexpected EOF") {
				return
			}
		}
		break
	default:
		err = fmt.Errorf("unexpected function string: \"%s\"", function)
	}
	return
}
