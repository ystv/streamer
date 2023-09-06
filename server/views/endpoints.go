package views

import (
	"encoding/xml"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/ystv/streamer/server/helper"
	"net/http"
	"strings"
)

// EndpointsFunc presents the endpoints to the user
func (v *Views) EndpointsFunc(c echo.Context) error {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication, http.StatusTemporaryRedirect)
		return
	}*/
	if v.conf.Verbose {
		fmt.Println("Endpoints called")
	}
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			fmt.Println("Endpoints POST")
		}
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		streamPageContent, err := helper.GetBody(v.conf.StreamServer + "/stat")
		if err != nil {
			fmt.Println(err)
		}

		var rtmp RTMP

		err = xml.Unmarshal([]byte(streamPageContent), &rtmp)
		if err != nil {
			fmt.Println(err)
		}

		var endpoints []string

		for i := 0; i < len(rtmp.Server.Applications); i++ {
			endpoints = append(endpoints, "endpoint~"+rtmp.Server.Applications[i].Name)
		}

		stringByte := strings.Join(endpoints, "\x20")

		return c.String(http.StatusOK, stringByte)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
