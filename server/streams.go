package main

import (
	"encoding/xml"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/ystv/streamer/server/helper"
	"net/http"
	"strings"
)

// streams collects the data from the rtmp stat page of nginx and produces a list of active streaming endpoints from given endpoints
func (web *Web) streams(w http.ResponseWriter, r *http.Request) {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication, http.StatusTemporaryRedirect)
		return
	}*/
	if r.Method == "POST" {
		if verbose {
			fmt.Println("Streams POST called")
		}
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
		}

		err = godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		streamPageContent, err := helper.GetBody(web.cfg.StreamChecker)
		if err != nil {
			fmt.Println(err)
		}

		var rtmp RTMP

		err = xml.Unmarshal([]byte(streamPageContent), &rtmp)
		if err != nil {
			fmt.Println(err)
		}

		var endpoints []string

		for key := range r.Form {
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
			_, err := w.Write([]byte(stringByte))
			if err != nil {
				fmt.Println(err)
			}
		} else {
			_, err := w.Write([]byte("No active streams with the current selection"))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
