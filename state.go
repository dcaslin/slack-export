package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type State struct {
	mu   sync.Mutex
	path string
	data map[string]string
}

func NewState(resetConf bool) (*State, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("unable to determine config directory: %w", err)
	}
	dir := filepath.Join(configDir, "slack-export")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create config directory: %w", err)
	}
	s := &State{
		path: filepath.Join(dir, "state.json"),
		data: make(map[string]string),
	}
	if resetConf {
		fmt.Println("Clearing config state, all messages will be exported")
		return s, s.save()
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *State) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *State) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return s.save()
}

func (s *State) load() error {
	raw, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to read state file: %w", err)
	}
	return json.Unmarshal(raw, &s.data)
}

func (s *State) save() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0644)
}
