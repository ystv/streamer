package views

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// websiteCheck checks if the website stream key is valid using software called COBRA
func (v *Views) websiteCheck(endpoint string) bool {
	if v.conf.Verbose {
		log.Println("Website Check called")
	}

	data := url.Values{}
	data.Set("call", "publish")
	var splitting []string
	data.Set("app", "live")
	splitting = strings.Split(endpoint, "?pwd=")
	data.Set("name", splitting[0])
	data.Set("pwd", splitting[1])

	ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, v.conf.KeyChecker, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("failed to create new request for website check: %+v", err)
		return true // sending back true if the checker is down
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	client.Timeout = 10 * time.Second

	res, err := client.Do(r)
	if err != nil {
		log.Printf("failed to send request for website check: %+v", err)
		return true // sending back true if the checker is down
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("failed to read body for website check: %+v", err)
		return true // sending back true if the checker is down
	}

	if (res.StatusCode >= 500 && res.StatusCode < 600) || res.StatusCode == http.StatusNotFound {
		log.Printf("failed to get correct status from key checker: %d", res.StatusCode)
		return true // sending back true if the checker is down
	} else if res.StatusCode == http.StatusUnauthorized {
		log.Println(res)
		url1, err := res.Location()
		log.Println(url1, err)
		log.Println(r)
		return true // sending back true if the checker is down
	}
	return string(body) == "Accepted"
}
