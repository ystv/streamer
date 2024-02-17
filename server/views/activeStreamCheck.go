package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ActiveStreamCheck checks if there are any existing streams still registered in the database
func (v *Views) ActiveStreamCheck() bool {
	if v.conf.Verbose {
		log.Println("Active Stream Check called")
	}

	streams, err := v.store.GetStreams()
	if err != nil {
		log.Printf("failed to get streams for activeStreamCheck: %+v", err)
		return false
	}

	return len(streams) > 0
}

func (v *Views) ActiveStreamsFunc(c echo.Context) error {
	streams, err := v.store.GetStreams()
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  fmt.Errorf("failed to get active streams: %w", err),
			Internal: fmt.Errorf("failed to get active streams: %w", err),
		}
	}

	//stored, err := v.store.GetStored()
	//if err != nil {
	//	return &echo.HTTPError{
	//		Code:     http.StatusInternalServerError,
	//		Message:  fmt.Errorf("failed to get stored streams: %w", err),
	//		Internal: fmt.Errorf("failed to get stored streams: %w", err),
	//	}
	//}

	data := struct {
		Streams int `json:"streams"`
	}{
		Streams: len(streams), /* + len(stored)*/
	}

	return c.JSON(http.StatusOK, data)
}
