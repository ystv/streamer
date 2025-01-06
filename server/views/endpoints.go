package views

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/helper"
)

// EndpointsFunc presents the endpoints to the user
func (v *Views) EndpointsFunc(c echo.Context) error {
	if v.conf.Verbose {
		log.Println("Endpoints called")
	}
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Endpoints POST")
		}

		var response struct {
			Endpoints []string `json:"endpoints"`
			Error     string   `json:"error"`
		}

		response.Endpoints = []string{}

		streamPageContent, err := helper.GetBody("http://" + v.conf.StreamServer + "stat")
		if err != nil {
			log.Printf("failed to get stats page body: %+v", err)
			response.Error = fmt.Sprintf("failed to get stats page body: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		var rtmp RTMP

		err = xml.Unmarshal([]byte(streamPageContent), &rtmp)
		if err != nil {
			log.Printf("failed to unmarshal xml: %+v", err)
			response.Error = fmt.Sprintf("failed to unmarshal xml: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		for _, application := range rtmp.Server.Applications {
			response.Endpoints = append(response.Endpoints, application.Name)
		}

		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
