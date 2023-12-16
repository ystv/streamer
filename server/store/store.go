package store

import (
	"fmt"
	_ "fmt"
	"github.com/ystv/streamer/server/storage"
)

type Store struct {
	backend Backend
}

func NewStore(root bool) (*Store, error) {
	backend, err := NewFileBackend(root)
	if err != nil {
		return nil, err
	}
	return &Store{backend: backend}, nil
}

func (store *Store) GetStreams() ([]*storage.Stream, error) {
	streamer, err := store.Get()
	if err != nil {
		return nil, err
	}
	return streamer.Stream, err
}

func (store *Store) FindStream(unique string) (*storage.Stream, error) {
	streamer, err := store.Get()
	if err != nil {
		return nil, err
	}
	for _, c1 := range streamer.Stream {
		if c1.Stream == unique {
			return c1, nil
		}
	}
	return nil, fmt.Errorf("unable to find stream for FindStream: %s", unique)
}

func (store *Store) AddStream(stream *storage.Stream) (*storage.Stream, error) {
	streamer, err := store.Get()
	if err != nil {
		return nil, err
	}

	for _, c := range streamer.Stream {
		if c.Stream == stream.Stream {
			return nil, fmt.Errorf("unable to add stream duplicate id for AddStream")
		}
	}

	streamer.Stream = append(streamer.Stream, stream)

	if err = store.backend.Write(streamer); err != nil {
		return nil, err
	}

	return stream, nil
}

func (store *Store) DeleteStream(unique string) error {
	streamer, err := store.backend.Read()
	if err != nil {
		return err
	}

	s := streamer.Stream
	found := false
	var index int
	var v *storage.Stream
	for index, v = range s {
		if v.Stream == unique {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:])   // Shift a[i+1:] left one index
		s[len(s)-1] = nil              // Erase last element (write zero value)
		streamer.Stream = s[:len(s)-1] // Truncate slice
	} else {
		return fmt.Errorf("stream not found for DeleteStream")
	}

	if err = store.backend.Write(streamer); err != nil {
		return err
	}

	return nil
}

func (store *Store) GetStored() ([]*storage.Stored, error) {
	streamer, err := store.Get()
	if err != nil {
		return nil, err
	}
	return streamer.Stored, err
}

func (store *Store) FindStored(unique string) (*storage.Stored, error) {
	streamer, err := store.Get()
	if err != nil {
		return nil, err
	}
	for _, c1 := range streamer.Stored {
		if c1.Stream == unique {
			return c1, nil
		}
	}
	return nil, fmt.Errorf("unable to find stored for FindStored: %s", unique)
}

func (store *Store) AddStored(stored *storage.Stored) (*storage.Stored, error) {
	streamer, err := store.Get()
	if err != nil {
		return nil, err
	}

	for _, c := range streamer.Stored {
		if c.Stream == stored.Stream {
			return nil, fmt.Errorf("unable to add stored duplicate id for AddStored")
		}
	}

	streamer.Stored = append(streamer.Stored, stored)

	if err = store.backend.Write(streamer); err != nil {
		return nil, err
	}

	return stored, nil
}

func (store *Store) Get() (*storage.Streamer, error) {
	return store.backend.Read()
}
