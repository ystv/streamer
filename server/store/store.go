package store

import (
	"errors"
	"fmt"

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
	return streamer.GetStream(), err
}

func (store *Store) FindStream(unique string) (*storage.Stream, error) {
	streamer, err := store.Get()
	if err != nil {
		return nil, err
	}
	for _, c1 := range streamer.GetStream() {
		if c1.GetStream() == unique {
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

	for _, c := range streamer.GetStream() {
		if c.GetStream() == stream.GetStream() {
			return nil, errors.New("unable to add stream duplicate id for AddStream")
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

	s := streamer.GetStream()
	found := false
	var index int
	var v *storage.Stream
	for index, v = range s {
		if v.GetStream() == unique {
			found = true
			break
		}
	}

	if !found {
		return errors.New("stream not found for DeleteStream")
	}

	copy(s[index:], s[index+1:])   // Shift a[i+1:] left one index
	s[len(s)-1] = nil              // Erase last element (write zero value)
	streamer.Stream = s[:len(s)-1] // Truncate slice

	if err = store.backend.Write(streamer); err != nil {
		return err
	}

	return nil
}

func (store *Store) GetStored() ([]*storage.Stored, error) {
	streamer, err := store.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get stored from GetStored: %w", err)
	}
	return streamer.GetStored(), nil
}

func (store *Store) FindStored(unique string) (*storage.Stored, error) {
	streamer, err := store.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get stored from FindStored: %w", err)
	}
	for _, c1 := range streamer.GetStored() {
		if c1.GetStream() == unique {
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

	for _, c := range streamer.GetStored() {
		if c.GetStream() == stored.GetStream() {
			return nil, errors.New("unable to add stored duplicate id for AddStored")
		}
	}

	streamer.Stored = append(streamer.Stored, stored)

	if err = store.backend.Write(streamer); err != nil {
		return nil, err
	}

	return stored, nil
}

func (store *Store) DeleteStored(unique string) error {
	streamer, err := store.backend.Read()
	if err != nil {
		return err
	}

	s := streamer.GetStored()
	found := false
	var index int
	var v *storage.Stored
	for index, v = range s {
		if v.GetStream() == unique {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:])   // Shift a[i+1:] left one index
		s[len(s)-1] = nil              // Erase last element (write zero value)
		streamer.Stored = s[:len(s)-1] // Truncate slice
	} else {
		return errors.New("stream not found for DeleteStored")
	}

	if err = store.backend.Write(streamer); err != nil {
		return err
	}

	return nil
}

func (store *Store) Get() (*storage.Streamer, error) {
	return store.backend.Read()
}
