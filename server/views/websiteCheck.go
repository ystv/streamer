package views

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	//apiStream "github.com/ystv/web-api/services/stream"
)

// websiteCheck checks if the website stream key is valid using software called COBRA
func (v *Views) websiteCheck(c echo.Context, endpoint string) bool {
	if v.conf.Verbose {
		log.Println("Website Check called")
	}

	var splitting []string
	splitting = strings.Split(endpoint, "?pwd=")

	ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	findEndpoint := struct {
		// Application defines which RTMP application this is valid for
		Application string `json:"application,omitempty"`
		// Name is the unique name given in an application
		Name string `json:"name,omitempty"`
		// Pwd defines an extra layer of security for authentication
		Pwd string `json:"pwd,omitempty"`
	}{
		Application: "live",
		Name:        splitting[0],
		Pwd:         splitting[1],
	}

	bytes, err := json.Marshal(findEndpoint)
	if err != nil {
		log.Printf("error marshalling website check request: %v", err)
		return false
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.conf.APIEndpoint+"/v1/internal/streams/find", strings.NewReader(string(bytes)))
	if err != nil {
		log.Printf("failed to create new request for website check: %+v", err)
		return true // sending back true if the checker is down
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(bytes)))

	session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
	if err != nil {
		log.Printf("failed to get session for authenticated: %+v", err)
		return false
	}

	token, ok := session.Values["token"].(string)
	if ok {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to send request for website check: %+v", err)
		return true // sending back true if the checker is down
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read body for website check: %+v", err)
		return true // sending back true if the checker is down
	}

	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		log.Printf("failed to get correct status from key checker: %d", resp.StatusCode)
		return true // sending back true if the checker is down
	} else if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		log.Printf("website check request failed with status code: %d", resp.StatusCode)
		return false
	} else if resp.StatusCode == http.StatusUnauthorized {
		log.Println(resp)
		url1, err := resp.Location()
		log.Println(url1, err)
		log.Println(req)
		return true // sending back true if the checker is down
	}
	var result struct {
		EndpointID int `json:"endpointId"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("failed to unmarshal response for website check request: %v", err)
		return false
	}
	return true
}
