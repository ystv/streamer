package views

import (
	"encoding/xml"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/helper"
	"net/http"
	"strings"
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
			fmt.Println("Streams POST called")
		}

		err := c.Request().ParseForm()
		if err != nil {
			c.Logger().Error(err)
			return err
		}

		streamPageContent, err := helper.GetBody("http://" + v.conf.StreamServer + "stat")
		if err != nil {
			c.Logger().Error(err)
			return err
		}

		var rtmp RTMP

		err = xml.Unmarshal([]byte(streamPageContent), &rtmp)
		if err != nil {
			c.Logger().Error(err)
			return err
		}

		var endpoints []string

		for key := range c.Request().Form {
			endpoint := strings.Split(key, "~")
			for i := 0; i < len(rtmp.Server.Applications); i++ {
				if rtmp.Server.Applications[i].Name == endpoint[1] {
					for j := 0; j < len(rtmp.Server.Applications[i].Live.Streams); j++ {
						endpoints = append(endpoints, endpoint[1]+"/"+rtmp.Server.Applications[i].Live.Streams[j].Name)
					}
				}
			}
		}

		if len(endpoints) != 0 {
			stringByte := strings.Join(endpoints, "\x20")
			return c.String(http.StatusOK, stringByte)
		} else {
			return c.String(http.StatusOK, "No active streams with the current selection")
		}
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
