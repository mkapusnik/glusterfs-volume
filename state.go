package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type volumeState struct {
	Name      string   `json:"name"`
	Servers   []string `json:"servers"`
	Volume    string   `json:"volume"`
	Subdir    string   `json:"subdir"`
	CreatedAt string   `json:"created_at"`
}

type stateStore struct {
	path string
}

func newStateStore(path string) *stateStore {
	return &stateStore{path: path}
}

func (s *stateStore) load() (map[string]volumeState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]volumeState{}, nil
		}
		return nil, err
	}
	var payload map[string]volumeState
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if payload == nil {
		payload = map[string]volumeState{}
	}
	return payload, nil
}

func (s *stateStore) save(state map[string]volumeState) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
