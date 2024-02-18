package views

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/helper/tx"
	"github.com/ystv/streamer/server/storage"
	"github.com/ystv/streamer/server/templates"
)

// SaveFunc allows for the functionality of saving a stream's details for later in order to make things easier for massive operations where you have multiple streams at once
func (v *Views) SaveFunc(c echo.Context) error {
	if c.Request().Method == "GET" {
		if v.conf.Verbose {
			log.Println("Save GET called")
		}

		return v.template.RenderTemplate(c.Response().Writer, nil, templates.SaveTemplate)
	} else if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Save POST called")
		}

		var response struct {
			Unique string `json:"unique"`
			Error  string `json:"error"`
		}

		endpoint := strings.Split(c.FormValue("endpointsTable"), "~")[1]

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
				log.Printf("failed to get streams: %+v", err)
				response.Error = fmt.Sprintf("failed to get streams: %+v", err)
				return c.JSON(http.StatusOK, response)
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

			if loop {
				continue
			}

			stored, err := v.store.GetStored()
			if err != nil {
				log.Printf("failed to get stored: %+v", err)
				response.Error = fmt.Sprintf("failed to get stored: %+v", err)
				return c.JSON(http.StatusOK, response)
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

		recording := ""
		website := ""

		if c.FormValue("record") == "on" {
			recording = c.FormValue("save_path")
		}

		if c.FormValue("website_stream") == "on" {
			website = c.FormValue("website_stream_endpoint")
		}

		var streams []string
		for _, index := range numbers {
			server := c.FormValue("stream_server_" + strconv.Itoa(index))
			server += "|"
			streams = append(streams, server+c.FormValue("stream_key_"+strconv.Itoa(index)))
		}

		stored := &storage.Stored{
			Stream:    string(b),
			Input:     endpoint + "/" + c.FormValue("stream_input"),
			Recording: recording,
			Website:   website,
			Streams:   streams,
		}

		s, err := v.store.AddStored(stored)
		if err != nil {
			log.Printf("failed to add stored: %+v, unique: %s", err, string(b))
			response.Error = fmt.Sprintf("failed to add stored: %+v, unique: %s", err, string(b))
			return c.JSON(http.StatusOK, response)
		}

		if s == nil {
			log.Printf("failed to add stored, stored is nil")
			response.Error = "failed to add stored, stored is nil"
			return c.JSON(http.StatusOK, response)
		}

		err = v.HandleTXLight(v.conf.TransmissionLight, tx.RehearsalOn)
		if err != nil {
			log.Printf("failed to turn transmission light on: %+v, ignoring and continuing", err)
		}

		log.Printf("saved stream: %s", string(b))
		response.Unique = string(b)
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
