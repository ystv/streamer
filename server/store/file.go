package store

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ystv/streamer/server/storage"

	"google.golang.org/protobuf/proto"
)

// FileBackend Applications: apps, Prefix: prefix
type FileBackend struct {
	path  string
	cache *storage.Streamer
	mutex sync.RWMutex
}

func NewFileBackend(root bool) (Backend, error) {
	var fb *FileBackend

	if root {
		fb = &FileBackend{path: "/db/store.db"}
	} else {
		fb = &FileBackend{path: "./db/store.db"}
	}

	state, err := fb.read(root)
	if err != nil {
		return nil, err
	}
	// persist state
	err = fb.save(state)
	if err != nil {
		return nil, err
	}
	fb.cache = state
	return fb, nil
}

// Read parses the store state from a file
func (fb *FileBackend) read(root bool) (*storage.Streamer, error) {
	var streamer storage.Streamer

	if root {
		_, err := os.Stat("/db")
		if err != nil {
			err = os.Mkdir("/db", 0777)
			if err != nil {
				return nil, fmt.Errorf("failed to make folder /db: %w", err)
			}
		}
	} else {
		_, err := os.Stat("./db")
		if err != nil {
			err = os.Mkdir("./db", 0777)
			if err != nil {
				return nil, fmt.Errorf("failed to make folder ./db: %w", err)
			}
		}
	}

	data, err := os.ReadFile(fb.path)
	// Non-existing streamer is ok
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("no previous file read: %w", err)
	}
	if err == nil {
		if err := proto.Unmarshal(data, &streamer); err != nil {
			return nil, fmt.Errorf("failed to parse stream streamer: %w", err)
		}
	}

	log.Printf("db file from: %s", fb.path)
	return &streamer, nil
}

// Save stores the store state in a file
func (fb *FileBackend) save(streamer *storage.Streamer) error {
	out, err := proto.Marshal(streamer)
	if err != nil {
		return fmt.Errorf("failed to encode streamer: %w", err)
	}
	tmp := fmt.Sprintf(fb.path+".%v", time.Now().Format("2006-01-02T15-04-05"))
	if err := os.WriteFile(tmp, out, 0600); err != nil {
		return fmt.Errorf("failed to write streamer: %w", err)
	}
	err = os.Rename(tmp, fb.path)
	if err != nil {
		return fmt.Errorf("failed to move streamer: %w", err)
	}
	return nil
}

func (fb *FileBackend) Read() (*storage.Streamer, error) {
	fb.mutex.RLock()
	defer fb.mutex.RUnlock()
	return fb.cache, nil
}

func (fb *FileBackend) Write(state *storage.Streamer) error {
	fb.mutex.Lock()
	defer fb.mutex.Unlock()
	fb.cache = state
	return fb.save(state)
}
