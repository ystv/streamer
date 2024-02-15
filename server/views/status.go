package views

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/action"
	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/common/wsMessages"
)

type (
	StatusResponse struct {
		Status []StatusResponseIndividual `json:"status"`
		Error  string                     `json:"error"`
	}

	StatusResponseIndividual struct {
		Name     string `json:"name"`
		Response string `json:"response"`
		Error    string `json:"error"`
	}
)

// StatusFunc is used to check the status of the streams and does this by tail command of the output logs
func (v *Views) StatusFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Status POST called")
		}

		errResponse := struct {
			Error string `json:"error"`
		}{}

		unique := c.FormValue("unique_code")
		if len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			errResponse.Error = fmt.Sprintf("unique key invalid: %s", unique)
			return c.JSON(http.StatusOK, errResponse)
		}

		stream, err := v.store.FindStream(unique)
		if err != nil {
			log.Printf("unable to find stream for status: %s, %+v", unique, err)
			errResponse.Error = fmt.Sprintf("unable to find stream for status: %s, %+v", unique, err)
			return c.JSON(http.StatusOK, errResponse)
		}

		if stream == nil {
			log.Println("failed to get stream as data is empty")
			errResponse.Error = "failed to get stream as data is empty"
			return c.JSON(http.StatusOK, errResponse)
		}

		transporter := commonTransporter.Transporter{
			Action: action.Status,
			Unique: unique,
		}

		//nolint:staticcheck
		fStatus := commonTransporter.ForwarderStatus{
			Website: stream.Website,
			Streams: int(stream.Streams),
		}

		var statusResponse StatusResponse
		//m := make(map[string]string)
		var wg sync.WaitGroup
		if stream.Recording {
			wg.Add(2)
			go func() {
				defer wg.Done()
				recorderTransporter := transporter

				individualResponse := StatusResponseIndividual{
					Name: "recording",
				}

				var response commonTransporter.ResponseTransporter
				response, err = v.wsHelper(server.Recorder, recorderTransporter)
				if err != nil {
					log.Printf("failed to send or receive message from recorder for status: %+v", err)
					individualResponse.Error = fmt.Sprintf("failed to send or receive message from recorder for status: %+v", err)
					statusResponse.Status = append(statusResponse.Status, individualResponse)
					//m["recording"] = fmt.Sprintf("failed to send or receive message from recorder for status: %+v", err)
					return
				}
				if response.Status == wsMessages.Error {
					log.Printf("failed to get correct response from recorder for status: %s", response.Payload)
					//m["recording"] = fmt.Sprintf("failed to get correct response from recorder for status: %s", response.Payload)
					individualResponse.Error = fmt.Sprintf("failed to get correct response from recorder for status: %s", response.Payload)
					statusResponse.Status = append(statusResponse.Status, individualResponse)
					return
				}
				if response.Status != wsMessages.Okay {
					log.Printf("invalid response from recorder for status: %s", response)
					individualResponse.Error = fmt.Sprintf("invalid response from recorder for status: %s", response)
					statusResponse.Status = append(statusResponse.Status, individualResponse)
					//m["recording"] = fmt.Sprintf("invalid response from recorder for status: %s", response)
					return
				}
				//m["recording"] = response.Payload.(string)

				individualResponse.Response = response.Payload.(string)
				statusResponse.Status = append(statusResponse.Status, individualResponse)

				log.Println("Recorder status success")
			}()
		} else {
			wg.Add(1)
		}
		go func() {
			defer wg.Done()
			forwarderTransporter := transporter

			forwarderTransporter.Payload = fStatus

			individualErrResponse := StatusResponseIndividual{
				Name: "1",
			}

			var response commonTransporter.ResponseTransporter
			response, err = v.wsHelper(server.Forwarder, forwarderTransporter)
			if err != nil {
				log.Printf("failed to send or receive message from forwarder for status: %+v", err)
				individualErrResponse.Error = fmt.Sprintf("failed to send or receive message from forwarder for status: %+v", err)
				statusResponse.Status = append(statusResponse.Status, individualErrResponse)
				//m["0"] = fmt.Sprintf("failed to send or receive message from forwarder for status: %+v", err)
				return
			}
			if response.Status == wsMessages.Error {
				log.Printf("failed to get correct response from forwarder for status: %s", response.Payload)
				individualErrResponse.Error = fmt.Sprintf("failed to get correct response from forwarder for status: %s", response.Payload)
				statusResponse.Status = append(statusResponse.Status, individualErrResponse)
				//m["0"] = fmt.Sprintf("failed to get correct response from forwarder for status: %s", response.Payload)
				return
			}
			if response.Status != wsMessages.Okay {
				log.Printf("invalid response from recorder for status: %s", response)
				individualErrResponse.Error = fmt.Sprintf("invalid response from recorder for status: %s", response)
				statusResponse.Status = append(statusResponse.Status, individualErrResponse)
				//m["0"] = fmt.Sprintf("invalid response from recorder for status: %s", response)
				return
			}

			var forwarderStatus commonTransporter.ForwarderStatusResponse

			err = mapstructure.Decode(response.Payload, &forwarderStatus)
			if err != nil {
				log.Printf("failed to decode message from forwarder for status: %+v", err)
				individualErrResponse.Error = fmt.Sprintf("failed to decode message from forwarder for status: %+v", err)
				statusResponse.Status = append(statusResponse.Status, individualErrResponse)
				//m["0"] = fmt.Sprintf("failed to decode message from forwarder for status: %+v", err)
				return
			}

			if len(forwarderStatus.Website) > 0 {
				tempResp := response.Payload.(string)
				tempRespArr := strings.Split(tempResp, "\r")
				var individualResponse StatusResponseIndividual
				if len(tempRespArr) == 0 {
					individualResponse = StatusResponseIndividual{
						Name:  "website",
						Error: "failed to get message response from forwarder for website",
					}
				} else {
					responseBuild := strings.ReplaceAll(tempRespArr[0], "\n", "<br>")
					responseBuild += "<br>"
					responseBuild += tempRespArr[len(tempRespArr)-1]
					individualResponse = StatusResponseIndividual{
						Name:     "website",
						Response: responseBuild,
					}
				}
				statusResponse.Status = append(statusResponse.Status, individualResponse)
				//m["website"] = forwarderStatus.Website
			}

			for index, streamOut := range forwarderStatus.Streams {
				tempRespArr := strings.Split(streamOut, "\r")
				var individualResponse StatusResponseIndividual
				if len(tempRespArr) == 0 {
					individualResponse = StatusResponseIndividual{
						Name:  index,
						Error: "failed to get message response from forwarder for stream " + index,
					}
				} else {
					responseBuild := strings.ReplaceAll(tempRespArr[0], "\n", "<br>")
					responseBuild += "<br>"
					responseBuild += tempRespArr[len(tempRespArr)-1]
					individualResponse = StatusResponseIndividual{
						Name:     index,
						Response: responseBuild,
					}
				}
				statusResponse.Status = append(statusResponse.Status, individualResponse)
				//m[strconv.Itoa(int(index))] = streamOut
			}

			log.Println("Forwarder status success")
		}()
		wg.Wait()

		return c.JSON(http.StatusOK, statusResponse)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
