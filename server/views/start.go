package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/helper"
	"github.com/ystv/streamer/server/helper/tx"
	"github.com/ystv/streamer/server/storage"
	"golang.org/x/crypto/ssh"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// StartFunc is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder
func (v *Views) StartFunc(c echo.Context) error {
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
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			fmt.Println("Start POST called")
		}

		recording := false
		websiteStream := false
		var streams uint64
		var forwarderStart string
		if c.FormValue("website_stream") == "on" {
			websiteStream = true
			if v.websiteCheck(c.FormValue("website_stream_endpoint")) {
				forwarderStart = "./forwarder_start " + c.FormValue("stream_selector") + " " + c.FormValue("website_stream_endpoint") + " "
			} else {
				return fmt.Errorf("website key check has failed")
			}
		} else {
			forwarderStart = "./forwarder_start \"" + c.FormValue("stream_selector") + "\" no "
		}
		largest := 0
		var numbers []int
		for s := range c.Request().PostForm {
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

		for loop {
			b = make([]byte, 10)
			for i := range b {
				b[i] = charset[seededRand.Intn(len(charset))]
			}

			streams1, err := v.store.GetStreams()
			if err != nil {
				return err
			}

			if len(streams1) == 0 {
				loop = false
				break
			}

			for _, s := range streams1 {
				if s.Stream == string(b) {
					loop = true
					break
				}
				loop = false
			}

			stored, err := v.store.GetStored()
			if err != nil {
				return err
			}

			if len(stored) == 0 {
				loop = false
				break
			}

			for _, s := range stored {
				if s.Stream == string(b) {
					loop = true
					break
				}
				loop = false
			}
		}

		forwarderStart += string(b) + " "
		for _, index := range numbers {
			server := c.FormValue("stream_server_" + strconv.Itoa(index))
			if server[len(server)-1] != '/' {
				server += "/"
			}
			forwarderStart += "\"" + server + "\" \"" + c.FormValue("stream_key_"+strconv.Itoa(index)) + "\" "
			streams++
		}
		forwarderStart += "| bash"

		recorderStart := "./recorder_start \"" + c.FormValue("stream_selector") + "\" \"" + c.FormValue("save_path") + "\" " + string(b) + " | bash"

		var wg sync.WaitGroup
		wg.Add(2)
		errors := false
		go func() {
			defer wg.Done()
			if c.FormValue("record") == "on" {
				recording = true
				var client *ssh.Client
				var session *ssh.Session
				var err error
				//if recorderAuth == "PEM" {
				//	client, session, err = connectToHostPEM(recorder, recorderUsername, recorderPrivateKey, recorderPassphrase)
				//} else if recorderAuth == "PASS" {
				client, session, err = helper.ConnectToHostPassword(v.conf.Recorder, v.conf.RecorderUsername, v.conf.RecorderPassword, v.conf.Verbose)
				//}
				if err != nil {
					fmt.Println(err, "Error connecting to Recorder for start")
					return
				}
				_, err = session.CombinedOutput(recorderStart)
				if err != nil {
					fmt.Println(err, "Error executing on Recorder for start")
					return
				}
				err = client.Close()
				if err != nil {
					fmt.Println(err)
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
			client, session, err = helper.ConnectToHostPassword(v.conf.Forwarder, v.conf.ForwarderUsername, v.conf.ForwarderPassword, v.conf.Verbose)
			//}
			if err != nil {
				fmt.Println(err, "Error connecting to Forwarder for start")
				errors = true
				return
			}
			_, err = session.CombinedOutput(forwarderStart)
			if err != nil {
				fmt.Println(err, "Error executing on Forwarder for start")
				errors = true
				return
			}
			err = client.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()
		wg.Wait()

		if errors == false {
			err := v.HandleTXLight(v.conf.TransmissionLight, tx.TransmissionOn)
			if err != nil {
				fmt.Println(err)
			}

			s, err := v.store.AddStream(&storage.Stream{
				Stream:    string(b),
				Input:     c.FormValue("stream_selector"),
				Recording: recording,
				Website:   websiteStream,
				Streams:   streams,
			})
			if err != nil {
				return err
			}

			if s == nil {
				return fmt.Errorf("failed to add stream, data is empty")
			}

			return c.String(http.StatusOK, string(b))
		}
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
