package views

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
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
			log.Println("Start POST called")
		}

		var response struct {
			Unique string `json:"unique"`
			Error  string `json:"error"`
		}

		var websiteOut string

		streamSelector := c.FormValue("stream_selector")
		if len(streamSelector) < 3 {
			log.Printf("invalid stream selector value")
			response.Error = fmt.Sprintf("invalid stream selector value")
			return c.JSON(http.StatusOK, response)
		}

		recordCheckboxRaw := c.FormValue("record_checkbox")
		if recordCheckboxRaw != "" && recordCheckboxRaw != "on" {
			log.Printf("invalid record checkbox value: %s", recordCheckboxRaw)
			response.Error = fmt.Sprintf("invalid record checkbox value: %s", recordCheckboxRaw)
			return c.JSON(http.StatusOK, response)
		}

		recordCheckbox := recordCheckboxRaw == "on"

		savePath := c.FormValue("save_path")
		if len(savePath) == 0 && recordCheckbox {
			log.Printf("invalid save path value")
			response.Error = fmt.Sprintf("invalid save path value")
			return c.JSON(http.StatusOK, response)
		}

		if recordCheckbox && !strings.HasSuffix(savePath, ".mkv") {
			log.Printf("the save path must end in \".mkv\"")
			response.Error = fmt.Sprintf("the save path must end in \".mkv\"")
			return c.JSON(http.StatusOK, response)
		}

		websiteCheckboxRaw := c.FormValue("website_stream")
		if websiteCheckboxRaw != "" && websiteCheckboxRaw != "on" {
			log.Printf("invalid website stream checkbox value: %s", recordCheckboxRaw)
			response.Error = fmt.Sprintf("invalid website stream checkbox value: %s", recordCheckboxRaw)
			return c.JSON(http.StatusOK, response)
		}

		websiteCheckbox := websiteCheckboxRaw == "on"

		websiteStreamEndpoint := c.FormValue("website_stream_endpoint")
		if websiteCheckbox && !strings.Contains(websiteStreamEndpoint, "?pwd=") {
			log.Printf("the website stream endpoint must contain \"?pwd=\"")
			response.Error = fmt.Sprintf("the website stream endpoint must contain \"?pwd=\"")
			return c.JSON(http.StatusOK, response)
		}

		if websiteCheckbox {
			if v.websiteCheck(websiteStreamEndpoint) {
				websiteOut = websiteStreamEndpoint
			} else {
				log.Printf("website key check has failed")
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

		streamServerRegex, err := regexp.Compile("^(rtmps?:\\/\\/)?" + // protocol
			"((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.)+[a-z]{2,}|" + // domain name
			"((\\d{1,3}\\.){3}\\d{1,3}))" + // OR ip (v4) address
			"(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*" + // port and path
			"(\\?[;&a-z\\d%_.~+=-]*)?" + // query string
			"(\\#[-a-z\\d_]*)?$") // fragment locator
		if err != nil {
			log.Printf("failed to compile regex: %+v", err)
			response.Error = fmt.Sprintf("failed to compile regex: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		var streams []string
		for _, index := range numbers {
			streamServer := c.FormValue("stream_server_" + strconv.Itoa(index))
			if len(streamServer) == 0 {
				log.Printf("invalid length of stream_server_%d", index)
				response.Error = fmt.Sprintf("invalid length of stream_server_%d", index)
				return c.JSON(http.StatusOK, response)
			}
			if !streamServerRegex.Match([]byte(streamServer)) {
				log.Printf("invalid value of stream_server_%d: %+v", index, err)
				response.Error = fmt.Sprintf("invalid value of stream_server_%d: %+v", index, err)
				return c.JSON(http.StatusOK, response)
			}
			if streamServer[len(streamServer)-1] != '/' {
				streamServer += "/"
			}
			streamKey := c.FormValue("stream_key_" + strconv.Itoa(index))
			if len(streamKey) == 0 {
				log.Printf("invalid length of stream_key_%d", index)
				response.Error = fmt.Sprintf("invalid length of stream_key_%d", index)
				return c.JSON(http.StatusOK, response)
			}
			streamServer += streamKey
			streams = append(streams, streamServer)
		}

		unique, err := v.generateUnique()
		if err != nil {
			log.Printf("failed to get unique: %+v", err)
			response.Error = fmt.Sprintf("failed to get unique: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		transporter := commonTransporter.Transporter{
			Action: action.Start,
		}

		transporter.Unique = unique

		rStart := commonTransporter.RecorderStart{
			StreamIn: streamSelector,
			PathOut:  savePath,
		}

		fStart := commonTransporter.ForwarderStart{
			StreamIn:   streamSelector,
			WebsiteOut: websiteOut,
		}

		fStart.Streams = streams

		var wg sync.WaitGroup
		wg.Add(2)
		var errorMessages []string
		go func() {
			defer wg.Done()
			if recordCheckbox {
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
				Input:     streamSelector,
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
