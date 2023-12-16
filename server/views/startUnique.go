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

// StartUniqueFunc is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder with a specified unique key
func (v *Views) StartUniqueFunc(c echo.Context) error {
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
			fmt.Println("StartUnique POST called")
		}

		unique := c.FormValue("unique_code")
		if len(unique) != 10 {
			return fmt.Errorf("unique key invalid")
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
				return fmt.Errorf("eebsite key check has failed")
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

		stored, err := v.store.FindStored(unique)
		if err != nil {
			return err
		}

		if stored == nil {
			return fmt.Errorf("failed to get stored as data is empty")
		}

		forwarderStart += unique + " "
		for _, index := range numbers {
			server := c.FormValue("stream_server_" + strconv.Itoa(index))
			if server[len(server)-1] != '/' {
				server += "/"
			}
			forwarderStart += "\"" + server + "\" \"" + c.FormValue("stream_key_"+strconv.Itoa(index)) + "\" "
			streams++
		}
		forwarderStart += "| bash"

		recorderStart := "./recorder_start \"" + c.FormValue("stream_selector") + "\" \"" + c.FormValue("save_path") + "\" " + unique + " | bash"

		var wg sync.WaitGroup
		wg.Add(2)
		errors := false
		go func() {
			defer wg.Done()
			if c.FormValue("record") == "on" {
				recording = true
				var client *ssh.Client
				var session *ssh.Session
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
				errors = true
				return
			}
		}()
		wg.Wait()

		if !errors {
			err = v.HandleTXLight(v.conf.TransmissionLight, tx.TransmissionOn)
			if err != nil {
				fmt.Println(err)
			}

			var s *storage.Stream

			s, err = v.store.AddStream(&storage.Stream{
				Stream:    unique,
				Input:     c.FormValue("stream_selector"),
				Recording: recording,
				Website:   websiteStream,
				Streams:   streams,
			})
			if err != nil {
				return err
			}

			if s == nil {
				return fmt.Errorf("failed to add stream as data is empty")
			}

			return c.String(http.StatusOK, unique)
		}
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
