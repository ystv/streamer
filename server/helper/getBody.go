package helper

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func GetBody(url string) (body string, err error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("failed to creating request: %v", err)
		return
	}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("failed to get http: %+v", err)
		return
	}
	defer response.Body.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, response.Body)
	if err != nil {
		log.Printf("failed to copy body: %+v", err)
		return
	}

	body = buf.String()

	return
}
