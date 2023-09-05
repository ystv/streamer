package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ystv/streamer/server/templates"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/mattn/go-sqlite3"
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

func errorFunc(errs string, w http.ResponseWriter) {
	fmt.Println("An error has occurred...\n" + errs)
	_, err := w.Write([]byte("An error has occurred...\n" + errs))
	if err != nil {
		fmt.Println(err)
	}
}
