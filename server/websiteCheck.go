package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// websiteCheck checks if the website stream key is valid using software called COBRA
func (web *Web) websiteCheck(endpoint string) bool {
	if verbose {
		fmt.Println("Website Check called")
	}

	data := url.Values{}
	data.Set("call", "publish")
	var splitting []string
	data.Set("app", "live")
	splitting = strings.Split(endpoint, "?pwd=")
	data.Set("name", splitting[0])
	data.Set("pwd", splitting[1])

	client := &http.Client{}
	r, err := http.NewRequest("POST", web.cfg.KeyChecker, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Println(err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(r)
	if err != nil {
		fmt.Println(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(res.Body)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	if string(body) == "Accepted" {
		return true
	} else {
		return false
	}
}
