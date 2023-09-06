package store

import (
	"github.com/ystv/streamer/server/storage"
)

type Backend interface {
	Read() (*storage.Streamer, error)
	Write(state *storage.Streamer) error
}
