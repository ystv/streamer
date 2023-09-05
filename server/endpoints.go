package main

import (
	"encoding/xml"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"strings"
)

// endpoints presents the endpoints to the user
func (web *Web) endpoints(w http.ResponseWriter, r *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication, http.StatusTemporaryRedirect)
		return
	}*/
	if verbose {
		fmt.Println("Endpoints called")
	}
	if r.Method == "POST" {
		if verbose {
			fmt.Println("Endpoints POST")
		}
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		response, err := http.Get(web.cfg.StreamChecker)
		if err != nil {
			fmt.Println(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(response.Body)

		buf := new(strings.Builder)
		_, err = io.Copy(buf, response.Body)
		if err != nil {
			fmt.Println(err)
		}

		streamPageContent := buf.String()

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
		_, err = w.Write([]byte(stringByte))
		if err != nil {
			fmt.Println(err)
		}
	}
}
