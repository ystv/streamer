package helper

import (
	"io"
	"net/http"
	"strings"
)

func GetBody(url string) (body string, err error) {
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, response.Body)
	if err != nil {
		return
	}

	body = buf.String()

	return
}
