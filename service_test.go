package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/slack-go/slack"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"normal-name", "normal-name"},
		{"has spaces", "has spaces"},
		{"has/slash", "has_slash"},
		{"has\\backslash", "has_backslash"},
		{"has:colon", "has_colon"},
		{"has\"quotes", "has_quotes"},
		{"has<angle>brackets", "has_angle_brackets"},
		{"has|pipe", "has_pipe"},
		{"has?question", "has_question"},
		{"has*star", "has_star"},
		{"multi<>:\"/\\|?*bad", "multi_________bad"},
	}
	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseSlackTs(t *testing.T) {
	ts := parseSlackTs("1619000000.123456")
	if ts.Unix() != 1619000000 {
		t.Errorf("expected unix 1619000000, got %d", ts.Unix())
	}
	// Microseconds should be preserved as nanoseconds
	if ts.Nanosecond()/1000 != 123456 {
		t.Errorf("expected 123456 microseconds, got %d", ts.Nanosecond()/1000)
	}
}

func TestParseSlackTsEmpty(t *testing.T) {
	ts := parseSlackTs("")
	// Empty string parses as Unix epoch (0 seconds)
	if ts.Unix() != 0 {
		t.Errorf("expected unix 0 for empty string, got %d", ts.Unix())
	}
}

func TestDateToSlackTs(t *testing.T) {
	tm := time.Unix(1619000000, 123456000)
	got := dateToSlackTs(&tm)
	if got != "1619000000.123456" {
		t.Errorf("expected 1619000000.123456, got %s", got)
	}
}

func TestDateToSlackTsNil(t *testing.T) {
	got := dateToSlackTs(nil)
	if got != "" {
		t.Errorf("expected empty string for nil, got %s", got)
	}
}

func TestBuildSimpleMsg(t *testing.T) {
	svc := &APIService{
		userMap: map[string]slack.User{
			"U123": {Name: "alice"},
		},
	}

	msg := slack.Message{}
	msg.User = "U123"
	msg.Text = "hello world"
	msg.Timestamp = "1619000000.000000"

	simple := svc.buildSimpleMsg(msg)
	if simple.Name != "alice" {
		t.Errorf("expected name alice, got %s", simple.Name)
	}
	if simple.Text != "hello world" {
		t.Errorf("expected text 'hello world', got %s", simple.Text)
	}
	if simple.Ts.Unix() != 1619000000 {
		t.Errorf("expected unix 1619000000, got %d", simple.Ts.Unix())
	}
}

func TestBuildSimpleMsgUnknownUser(t *testing.T) {
	svc := &APIService{
		userMap: map[string]slack.User{},
	}

	msg := slack.Message{}
	msg.User = "UNOTFOUND"
	msg.Text = "test"
	msg.Timestamp = "1619000000.000000"

	simple := svc.buildSimpleMsg(msg)
	if simple.Name != "unknown" {
		t.Errorf("expected name 'unknown', got %s", simple.Name)
	}
}

func TestWriteMsgs(t *testing.T) {
	dir := t.TempDir()
	svc := &APIService{exportRoot: dir}

	msgs := []SimpleMsg{
		{Name: "alice", Text: "hello", Ts: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
		{Name: "bob", Text: "world", Ts: time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC)},
	}

	conv := conversation{convType: "pu-ch", name: "general"}
	if err := svc.writeMsgs(conv, msgs); err != nil {
		t.Fatalf("writeMsgs failed: %v", err)
	}

	// Find the written file
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}

	raw, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	content := string(raw)

	if !strings.Contains(content, "2024-01-15 10:30:00 alice: hello") {
		t.Errorf("missing first message in output: %s", content)
	}
	if !strings.Contains(content, "2024-01-15 10:31:00 bob: world") {
		t.Errorf("missing second message in output: %s", content)
	}
	// Verify filename format
	name := entries[0].Name()
	if !strings.HasPrefix(name, "pu-ch-general_") || !strings.HasSuffix(name, ".txt") {
		t.Errorf("unexpected filename: %s", name)
	}
}

func TestNewAPIServiceDateParsing(t *testing.T) {
	state := &State{data: make(map[string]string)}
	dir := t.TempDir()

	// RFC3339 format
	svc, err := NewAPIService(state, "token", dir, "2024-01-15T10:00:00Z", "2024-06-15T10:00:00Z", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.startDate == nil || svc.startDate.Year() != 2024 || svc.startDate.Month() != 1 {
		t.Errorf("unexpected startDate: %v", svc.startDate)
	}
	if svc.endDate == nil || svc.endDate.Month() != 6 {
		t.Errorf("unexpected endDate: %v", svc.endDate)
	}

	// YYYY-MM-DD format
	svc2, err := NewAPIService(state, "token", dir, "2024-03-01", "", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc2.startDate == nil || svc2.startDate.Day() != 1 || svc2.startDate.Month() != 3 {
		t.Errorf("unexpected startDate: %v", svc2.startDate)
	}
	if svc2.endDate != nil {
		t.Errorf("expected nil endDate, got %v", svc2.endDate)
	}
}

func TestNewAPIServiceInvalidDate(t *testing.T) {
	state := &State{data: make(map[string]string)}
	dir := t.TempDir()

	_, err := NewAPIService(state, "token", dir, "not-a-date", "", false, false)
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
	if !strings.Contains(err.Error(), "invalid MIN_DATE_ISO") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNewAPIServiceEmptyDates(t *testing.T) {
	state := &State{data: make(map[string]string)}
	dir := t.TempDir()

	svc, err := NewAPIService(state, "token", dir, "", "", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.startDate != nil || svc.endDate != nil {
		t.Error("expected nil dates for empty strings")
	}
}
