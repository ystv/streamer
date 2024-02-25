package views

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/helper"
)

// StreamsFunc collects the data from the rtmp stat page of nginx and produces a list of active streaming endpoints from given endpoints
func (v *Views) StreamsFunc(c echo.Context) error {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication, http.StatusTemporaryRedirect)
		return
	}*/
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Streams POST called")
		}

		var response struct {
			Streams []string `json:"streams"`
			Error   string   `json:"error"`
		}

		response.Streams = []string{}

		err := c.Request().ParseForm()
		if err != nil {
			log.Printf("failed to parse form: %+v", err)
			response.Error = fmt.Sprintf("failed to parse form: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

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

		for key := range c.Request().Form {
			endpoint := strings.Split(key, "~")[1]
			for _, application := range rtmp.Server.Applications {
				if application.Name == endpoint {
					for _, applicationStream := range application.Live.Streams {
						response.Streams = append(response.Streams, fmt.Sprintf("%s/%s", endpoint, applicationStream.Name))
					}
				}
			}
		}

		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
