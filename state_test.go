package main

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestState(t *testing.T) *State {
	t.Helper()
	dir := t.TempDir()
	s := &State{
		path: filepath.Join(dir, "state.json"),
		data: make(map[string]string),
	}
	return s
}

func TestStateGetSet(t *testing.T) {
	s := newTestState(t)

	// Get on empty state returns false
	_, ok := s.Get("key1")
	if ok {
		t.Fatal("expected key to not exist")
	}

	// Set and Get
	if err := s.Set("key1", "value1"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	v, ok := s.Get("key1")
	if !ok || v != "value1" {
		t.Fatalf("expected value1, got %q (ok=%v)", v, ok)
	}
}

func TestStatePersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	// Create state and set a value
	s1 := &State{path: path, data: make(map[string]string)}
	if err := s1.Set("chan1", "2024-01-01T00:00:00Z"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Create new state pointing to same file and load
	s2 := &State{path: path, data: make(map[string]string)}
	if err := s2.load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	v, ok := s2.Get("chan1")
	if !ok || v != "2024-01-01T00:00:00Z" {
		t.Fatalf("expected persisted value, got %q (ok=%v)", v, ok)
	}
}

func TestStateLoadMissingFile(t *testing.T) {
	s := &State{
		path: filepath.Join(t.TempDir(), "nonexistent.json"),
		data: make(map[string]string),
	}
	if err := s.load(); err != nil {
		t.Fatalf("load of missing file should succeed, got: %v", err)
	}
	if len(s.data) != 0 {
		t.Fatal("expected empty data")
	}
}

func TestStateReset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	s := &State{path: path, data: make(map[string]string)}
	_ = s.Set("key1", "val1")

	// Simulate reset: create fresh state and save empty
	s2 := &State{path: path, data: make(map[string]string)}
	if err := s2.save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Verify file exists but state is empty
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(raw) != "{}" {
		t.Fatalf("expected empty JSON object, got %s", raw)
	}
}
