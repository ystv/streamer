package views

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/action"
	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/common/wsMessages"
	"github.com/ystv/streamer/server/helper/tx"
	"github.com/ystv/streamer/server/storage"
)

// StartFunc is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder
func (v *Views) StartFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			fmt.Println("Start POST called")
		}

		transporter := commonTransporter.Transporter{
			Action: action.Start,
		}

		fStart := commonTransporter.ForwarderStart{
			StreamIn: c.FormValue("stream_selector"),
		}

		rStart := commonTransporter.RecorderStart{
			StreamIn: c.FormValue("stream_selector"),
			PathOut:  c.FormValue("save_path"),
		}

		recording := false
		websiteStream := false
		if c.FormValue("website_stream") == "on" {
			websiteStream = true
			if v.websiteCheck(c.FormValue("website_stream_endpoint")) {
				fStart.WebsiteOut = c.FormValue("website_stream_endpoint")
			} else {
				return fmt.Errorf("website key check has failed")
			}
		}

		// This section finds the number of the stream from the form
		// You can miss values out and some rearranging will have to be done
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
				return fmt.Errorf("failed to get streams: %w", err)
			}

			if len(streams1) == 0 {
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
				return fmt.Errorf("failed to get stored: %w", err)
			}

			if len(stored) == 0 {
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

		transporter.Unique = string(b)

		var streams []string
		for _, index := range numbers {
			streamServer := c.FormValue("stream_server_" + strconv.Itoa(index))
			if streamServer[len(streamServer)-1] != '/' {
				streamServer += "/"
			}
			streamServer += c.FormValue("stream_key_" + strconv.Itoa(index))
			streams = append(streams, streamServer)
		}

		fStart.Streams = streams

		var wg sync.WaitGroup
		wg.Add(2)
		errors := false
		go func() {
			defer wg.Done()
			if c.FormValue("record") == "on" {
				recording = true
				recorderTransporter := transporter
				recorderTransporter.Payload = rStart
				response, err := v.wsHelper(server.Recorder, recorderTransporter)
				if err != nil {
					log.Println(err, "Error sending to Recorder for start")
					errors = true
					return
				}
				if response.Status == wsMessages.Error {
					log.Printf("Error sending to Recorder for start: %s", response.Payload)
					errors = true
					return
				}
				if response.Status != wsMessages.Okay {
					log.Printf("invalid response from Recorder for start: %s", response.Status)
					errors = true
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			forwarderTransporter := transporter
			forwarderTransporter.Payload = fStart
			response, err := v.wsHelper(server.Forwarder, forwarderTransporter)
			if err != nil {
				log.Println(err, "Error sending to Forwarder for start")
				errors = true
				return
			}
			if response.Status == wsMessages.Error {
				log.Printf("Error sending to Forwarder for start: %s", response)
				errors = true
				return
			}
			if response.Status != wsMessages.Okay {
				log.Printf("invalid response from Forwarder for start: %s", response)
				errors = true
				return
			}
		}()
		wg.Wait()

		if !errors {
			err := v.HandleTXLight(v.conf.TransmissionLight, tx.TransmissionOn)
			if err != nil {
				log.Printf("failed to turn transmission light on: %+v, ignoring and continuing", err)
			}

			s, err := v.store.AddStream(&storage.Stream{
				Stream:    string(b),
				Input:     c.FormValue("stream_selector"),
				Recording: recording,
				Website:   websiteStream,
				Streams:   uint64(len(streams)),
			})
			if err != nil {
				return fmt.Errorf("failed to add stream for start: %w", err)
			}

			if s == nil {
				return fmt.Errorf("failed to add stream, data is empty")
			}

			return c.String(http.StatusOK, string(b))
		}
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
