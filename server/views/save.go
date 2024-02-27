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

		endpoint := strings.Split(c.FormValue("endpoints_table"), "~")[1]

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

		unique, err := v.generateUnique()
		if err != nil {
			log.Printf("failed to get unique: %+v", err)
			response.Error = fmt.Sprintf("failed to get unique: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		recording := ""
		website := ""

		if c.FormValue("record_checkbox") == "on" {
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
			Stream:    unique,
			Input:     endpoint + "/" + c.FormValue("stream_input"),
			Recording: recording,
			Website:   website,
			Streams:   streams,
		}

		s, err := v.store.AddStored(stored)
		if err != nil {
			log.Printf("failed to add stored: %+v, unique: %s", err, unique)
			response.Error = fmt.Sprintf("failed to add stored: %+v, unique: %s", err, unique)
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

		log.Printf("saved stream: %s", unique)
		response.Unique = unique
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
