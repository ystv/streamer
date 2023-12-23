package helper

import (
	"io"
	"log"
	"net/http"
	"strings"
)

func GetBody(url string) (body string, err error) {
	response, err := http.Get(url)
	if err != nil {
		log.Printf("failed to get http: %+v", err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, response.Body)
	if err != nil {
		log.Printf("failed to copy body: %+v", err)
		return
	}

	body = buf.String()

	return
}
