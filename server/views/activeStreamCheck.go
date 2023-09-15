package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

// ActiveStreamCheck checks if there are any existing streams still registered in the database
func (v *Views) ActiveStreamCheck() bool {
	if v.conf.Verbose {
		fmt.Println("Active Stream Check called")
	}

	streams, err := v.store.GetStreams()
	if err != nil {
		fmt.Println(err)
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

	stored, err := v.store.GetStored()
	if err != nil {
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  fmt.Errorf("failed to get stored streams: %w", err),
			Internal: fmt.Errorf("failed to get stored streams: %w", err),
		}
	}

	data := struct {
		Streams int `json:"streams"`
	}{
		Streams: len(streams) + len(stored),
	}

	return c.JSON(http.StatusOK, data)
}
