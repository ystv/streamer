package views

import (
	"encoding/xml"
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
	"github.com/ystv/streamer/server/helper"
	"github.com/ystv/streamer/server/helper/tx"
	"github.com/ystv/streamer/server/storage"
)

// StartUniqueFunc is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder with a specified unique key
func (v *Views) StartUniqueFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("StartUnique POST called")
		}

		var response struct {
			Unique string `json:"unique"`
			Error  string `json:"error"`
		}

		unique := c.FormValue("unique_code")
		if len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			response.Error = fmt.Sprintf("unique key invalid: %s", unique)
			return c.JSON(http.StatusOK, response)
		}

		stored, err := v.store.FindStored(unique)
		if err != nil {
			log.Printf("unable to find unique code for startUnique: %s, %+v", unique, err)
			response.Error = fmt.Sprintf("unable to find unique code for startUnique: %s, %+v", unique, err)
			return c.JSON(http.StatusOK, response)
		}

		if stored == nil {
			log.Printf("failed to get stored as data is empty")
			response.Error = "failed to get stored as data is empty"
			return c.JSON(http.StatusOK, response)
		}

		transporter := commonTransporter.Transporter{
			Action: action.Start,
			Unique: unique,
		}

		inputEndpoint := c.FormValue("endpoints_table")
		inputStream := c.FormValue("stream_input")

		streamPageContent, err := helper.GetBody("http://" + v.conf.StreamServer + "stat")
		if err != nil {
			log.Printf("failed to get streams from stream server: %+v", err)
			response.Error = fmt.Sprintf("failed to get streams from stream server: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		var rtmp RTMP

		err = xml.Unmarshal([]byte(streamPageContent), &rtmp)
		if err != nil {
			log.Printf("failed to unmarshal xml: %+v", err)
			response.Error = fmt.Sprintf("failed to unmarshal xml: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		found := false

		var streamIn string
		endpoint := strings.Split(inputEndpoint, "~")
	applicationFor:
		for i := 0; i < len(rtmp.Server.Applications); i++ {
			if rtmp.Server.Applications[i].Name == endpoint[1] {
				for j := 0; j < len(rtmp.Server.Applications[i].Live.Streams); j++ {
					if rtmp.Server.Applications[i].Live.Streams[j].Name == inputStream {
						found = true
						streamIn = endpoint[1] + "/" + rtmp.Server.Applications[i].Live.Streams[j].Name
						break applicationFor
					}
				}
			}
		}

		if !found {
			log.Printf("unable to find current stream input")
			response.Error = "unable to find current stream input"
			return c.JSON(http.StatusOK, response)
		}

		fStart := commonTransporter.ForwarderStart{
			StreamIn: streamIn,
		}

		rStart := commonTransporter.RecorderStart{
			StreamIn: streamIn,
			PathOut:  c.FormValue("save_path"),
		}

		if c.FormValue("website_stream") == "on" {
			if v.websiteCheck(c.FormValue("website_stream_endpoint")) {
				fStart.WebsiteOut = c.FormValue("website_stream_endpoint")
			} else {
				response.Error = "website key check has failed"
				return c.JSON(http.StatusOK, response)
			}
		}

		// This section finds the number of the stream from the form
		// You can miss values out, and some rearranging will have to be done
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
		var errorMessages []string
		go func() {
			defer wg.Done()
			if c.FormValue("record_checkbox") == "on" {
				recorderTransporter := transporter
				recorderTransporter.Payload = rStart
				var wsResponse commonTransporter.ResponseTransporter
				wsResponse, err = v.wsHelper(server.Recorder, recorderTransporter)
				if err != nil {
					log.Printf("failed sending to Recorder for start: %+v", err)
					errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Recorder for start: %+v", err))
					return
				}
				if wsResponse.Status == wsMessages.Error {
					log.Printf("failed sending to Recorder for start: %s", wsResponse.Payload)
					errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Recorder for start: %s", wsResponse.Payload))
					return
				}
				if wsResponse.Status != wsMessages.Okay {
					log.Printf("invalid response from Recorder for start: %s", wsResponse.Status)
					errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Recorder for start: %s", wsResponse.Status))
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			forwarderTransporter := transporter
			forwarderTransporter.Payload = fStart
			var wsResponse commonTransporter.ResponseTransporter
			wsResponse, err = v.wsHelper(server.Forwarder, forwarderTransporter)
			if err != nil {
				log.Printf("failed sending to Forwarder for start: %+v", err)
				errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Forwarder for start: %+v", err))
				return
			}
			if wsResponse.Status == wsMessages.Error {
				log.Printf("failed sending to Forwarder for start: %s", response)
				errorMessages = append(errorMessages, fmt.Sprintf("failed sending to Forwarder for start: %s", response))
				return
			}
			if wsResponse.Status != wsMessages.Okay {
				log.Printf("invalid response from Forwarder for start: %s", response)
				errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Forwarder for start: %s", response))
				return
			}
		}()
		wg.Wait()

		if len(errorMessages) == 0 {
			err = v.HandleTXLight(v.conf.TransmissionLight, tx.TransmissionOn)
			if err != nil {
				log.Printf("failed to turn transmission light on: %+v, ignoring and continuing", err)
			}

			var s *storage.Stream
			s, err = v.store.AddStream(&storage.Stream{
				Stream:    unique,
				Input:     streamIn,
				Recording: rStart.PathOut,
				Website:   fStart.WebsiteOut,
				Streams:   streams,
			})
			if err != nil {
				log.Printf("invalid response from Forwarder for start: %s", response)
				errorMessages = append(errorMessages, fmt.Sprintf("invalid response from Forwarder for start: %s", response))
				response.Error = strings.Join(errorMessages, ",")
				return c.JSON(http.StatusOK, response)
			}

			if s == nil {
				log.Printf("failed to add stream, data is empty")
				errorMessages = append(errorMessages, "failed to add stream, data is empty")
				response.Error = strings.Join(errorMessages, ",")
				return c.JSON(http.StatusOK, response)
			}

			err = v.store.DeleteStored(unique)
			if err != nil {
				log.Printf("failed to delete stored: %+v, unique: %s", err, unique)
				response.Error = fmt.Sprintf("failed to delete stored: %+v, unique: %s", err, unique)
				return c.JSON(http.StatusOK, response)
			}

			log.Printf("started stream: %s", unique)

			response.Unique = unique
			return c.JSON(http.StatusOK, response)
		}

		errMsg := strings.Join(errorMessages, ",")
		log.Printf("failed to start: %+v", errMsg)
		response.Error = fmt.Sprintf("failed to start: %+v", errMsg)
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
