package views

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/helper"
)

func (v *Views) startSaveValidationHelper(c echo.Context, valType ValidationType) StartSaveValidationResponse {
	var response StartSaveValidationResponse

	var input string

	switch valType {
	case startValidation:
		input = c.FormValue("stream_selector")
		if len(input) < 3 {
			response.Error = fmt.Errorf("invalid stream selector value: %s", input)
			return response
		}
	case startUniqueValidation:
		inputEndpoint := c.FormValue("endpoints_table")
		inputStream := c.FormValue("stream_input")

		streamPageContent, err := helper.GetBody("http://" + v.conf.StreamServer + "stat")
		if err != nil {
			response.Error = fmt.Errorf("failed to get streams from stream server: %w", err)
			return response
		}

		var rtmp RTMP

		err = xml.Unmarshal([]byte(streamPageContent), &rtmp)
		if err != nil {
			response.Error = fmt.Errorf("failed to unmarshal xml: %w", err)
			return response
		}

		found := false
	applicationFor:
		for _, application := range rtmp.Server.Applications {
			if application.Name == inputEndpoint {
				for _, stream := range application.Live.Streams {
					if stream.Name == inputStream {
						found = true
						input = inputEndpoint + "/" + stream.Name
						break applicationFor
					}
				}
			}
		}

		if !found {
			response.Error = errors.New("unable to find current stream input")
			return response
		}
	case saveValidation:
		endpoint := c.FormValue("endpoints_table")
		if len(endpoint) < 2 {
			response.Error = fmt.Errorf("invalid endpoint selected: %s", endpoint)
			return response
		}
		streamInput := c.FormValue("stream_input")
		if len(streamInput) < 2 {
			response.Error = fmt.Errorf("invalid stream input: %s", streamInput)
			return response
		}
		input = fmt.Sprintf("%s/%s", endpoint, streamInput)
	default:
		response.Error = errors.New("invalid validation type")
		return response
	}

	recordCheckboxRaw := c.FormValue("record_checkbox")
	if recordCheckboxRaw != "" && recordCheckboxRaw != "on" {
		response.Error = fmt.Errorf("invalid record checkbox value: %s", recordCheckboxRaw)
		return response
	}

	recordCheckbox := recordCheckboxRaw == "on"

	savePath := c.FormValue("save_path")
	if len(savePath) == 0 && recordCheckbox {
		response.Error = errors.New("invalid save path value")
		return response
	}

	if recordCheckbox && !strings.HasSuffix(savePath, ".mkv") {
		response.Error = errors.New("the save path must end in \".mkv\"")
		return response
	}

	websiteCheckboxRaw := c.FormValue("website_stream")
	if websiteCheckboxRaw != "" && websiteCheckboxRaw != "on" {
		response.Error = errors.New("invalid website stream checkbox value: " + recordCheckboxRaw)
		return response
	}

	websiteCheckbox := websiteCheckboxRaw == "on"

	websiteStreamEndpoint := c.FormValue("website_stream_endpoint")
	if websiteCheckbox && !strings.Contains(websiteStreamEndpoint, "?pwd=") {
		response.Error = errors.New("the website stream endpoint must contain \"?pwd=\"")
		return response
	}

	var websiteOut string

	if websiteCheckbox {
		if v.websiteCheck(websiteStreamEndpoint) {
			websiteOut = websiteStreamEndpoint
		} else {
			response.Error = errors.New("website key check has failed")
			return response
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

	streamServerRegex := regexp.MustCompile("^(rtmps?:\\/\\/)?" + // protocol
		"((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.)+[a-z]{2,}|" + // domain name
		"((\\d{1,3}\\.){3}\\d{1,3}))" + // OR ip (v4) address
		"(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*" + // port and path
		"(\\?[;&a-z\\d%_.~+=-]*)?" + // query string
		"(\\#[-a-z\\d_]*)?$") // fragment locator

	streams := make([]string, 0)
	for _, index := range numbers {
		streamServer := c.FormValue("stream_server_" + strconv.Itoa(index))
		if len(streamServer) == 0 {
			response.Error = errors.New("invalid length of stream_server_" + strconv.Itoa(index))
			return response
		}
		if !streamServerRegex.Match([]byte(streamServer)) {
			response.Error = errors.New("invalid value of stream_server_" + strconv.Itoa(index))
			return response
		}
		if streamServer[len(streamServer)-1] != '/' {
			streamServer += "/"
		}
		if valType == saveValidation {
			streamServer += "|"
		}
		streamKey := c.FormValue("stream_key_" + strconv.Itoa(index))
		if len(streamKey) == 0 {
			response.Error = errors.New("invalid length of stream_key_" + strconv.Itoa(index))
			return response
		}
		streamServer += streamKey
		streams = append(streams, streamServer)
	}

	if len(streams) == 0 {
		response.Error = errors.New("invalid length of streams")
		return response
	}

	response.Input = input
	response.RecordCheckbox = recordCheckbox
	response.SavePath = savePath
	response.WebsiteCheckbox = websiteCheckbox
	response.WebsiteOut = websiteOut
	response.Streams = streams

	return response
}
