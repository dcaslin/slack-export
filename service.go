package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

var requiredScopes = []string{
	"channels:history",
	"channels:read",
	"groups:history",
	"groups:read",
	"im:history",
	"im:read",
	"mpim:history",
	"mpim:read",
	"files:read",
	"users:read",
}

var sanitizeRe = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

func sanitizeFilename(name string) string {
	return sanitizeRe.ReplaceAllString(name, "_")
}

type APIService struct {
	client     *slack.Client
	httpClient *http.Client
	token      string
	state      *State
	exportRoot string
	startDate  *time.Time
	endDate    *time.Time
	exportJSON bool
	exportFiles bool
	userMap    map[string]slack.User
	users      []slack.User
}

func NewAPIService(state *State, token, exportRoot string, startDateISO, endDateISO string, exportJSON, exportFiles bool) (*APIService, error) {
	s := &APIService{
		client:     slack.New(token),
		httpClient: &http.Client{},
		token:      token,
		state:      state,
		exportRoot: exportRoot,
		exportJSON: exportJSON,
		exportFiles: exportFiles,
		userMap:    make(map[string]slack.User),
	}
	if startDateISO != "" {
		t, err := time.Parse(time.RFC3339, startDateISO)
		if err != nil {
			t, err = time.Parse("2006-01-02", startDateISO)
			if err != nil {
				return nil, fmt.Errorf("invalid MIN_DATE_ISO: %w", err)
			}
		}
		s.startDate = &t
	}
	if endDateISO != "" {
		t, err := time.Parse(time.RFC3339, endDateISO)
		if err != nil {
			t, err = time.Parse("2006-01-02", endDateISO)
			if err != nil {
				return nil, fmt.Errorf("invalid MAX_DATE_ISO: %w", err)
			}
		}
		s.endDate = &t
	}
	return s, nil
}

func (s *APIService) Export() error {
	if _, err := os.Stat(s.exportRoot); err != nil {
		return fmt.Errorf("EXPORT_ROOT %s does not exist or is inaccessible: %w", s.exportRoot, err)
	}

	resp, err := s.client.AuthTest()
	if err != nil {
		return fmt.Errorf("auth test failed: %w", err)
	}
	fmt.Printf("Operating on team %q and user %q\n", resp.Team, resp.User)
	fmt.Printf("Writing output to %s\n", s.exportRoot)

	// Verify scopes by making the auth test - slack-go doesn't expose scopes directly,
	// so we rely on API calls failing if scopes are missing.

	if err := s.loadUsers(); err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	conversations, err := s.loadConversations()
	if err != nil {
		return fmt.Errorf("failed to load conversations: %w", err)
	}

	for _, conv := range conversations {
		if err := s.writeNewMsgs(conv); err != nil {
			return fmt.Errorf("failed to export conversation %s: %w", conv.name, err)
		}
	}
	return nil
}

type conversation struct {
	id        string
	name      string
	convType  string
	isPrivate bool
}

func (s *APIService) loadUsers() error {
	users, err := s.client.GetUsers()
	if err != nil {
		return err
	}
	s.users = users
	for _, u := range users {
		s.userMap[u.ID] = u
	}
	return nil
}

func (s *APIService) loadConversations() ([]conversation, error) {
	params := &slack.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel", "mpim", "im"},
		ExcludeArchived: false,
		Limit:           1000,
	}

	var conversations []conversation
	for {
		channels, cursor, err := s.client.GetConversations(params)
		if err != nil {
			return nil, err
		}
		for _, ch := range channels {
			conv := conversation{
				id:        ch.ID,
				name:      ch.Name,
				isPrivate: ch.IsPrivate,
			}
			if ch.IsIM {
				if u, ok := s.userMap[ch.User]; ok {
					conv.name = u.Name
				}
				conv.isPrivate = true
				conv.convType = "d-msg"
			} else if ch.IsMpIM {
				conv.convType = "g-msg"
			} else if ch.IsPrivate {
				conv.convType = "pr-ch"
			} else {
				conv.convType = "pu-ch"
			}
			conversations = append(conversations, conv)
		}
		if cursor == "" {
			break
		}
		params.Cursor = cursor
	}
	return conversations, nil
}

func (s *APIService) writeNewMsgs(conv conversation) error {
	var oldestDate *time.Time

	if lastISO, ok := s.state.Get(conv.id); ok {
		t, err := time.Parse(time.RFC3339, lastISO)
		if err == nil {
			oldestDate = &t
			fmt.Printf("   %s Using previously last saved date %s\n", conv.name, t.Format("2006-01-02 15:04:05"))
		}
	} else if s.endDate != nil {
		oldestDate = s.startDate
	}

	finishDate, err := s.writeMessages(conv, oldestDate, s.endDate)
	if err != nil {
		return err
	}
	if finishDate != nil {
		return s.state.Set(conv.id, finishDate.Format(time.RFC3339))
	}
	return nil
}

func dateToSlackTs(t *time.Time) string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%d.%06d", t.Unix(), t.Nanosecond()/1000)
}

func (s *APIService) writeMessages(conv conversation, oldestDate, newestDate *time.Time) (*time.Time, error) {
	var msgs []SimpleMsg
	var jsonMsgs []slack.Message

	params := &slack.GetConversationHistoryParameters{
		ChannelID: conv.id,
		Limit:     1000,
	}
	if oldestDate != nil {
		params.Oldest = dateToSlackTs(oldestDate)
	}
	if newestDate != nil {
		params.Latest = dateToSlackTs(newestDate)
	}

	extraPages := false
	for {
		if extraPages {
			fmt.Print("-")
		}

		history, err := s.client.GetConversationHistory(params)
		if err != nil {
			return nil, err
		}

		for _, msg := range history.Messages {
			simpleMsg := s.buildSimpleMsg(msg)
			msgs = append(msgs, simpleMsg)

			if s.exportJSON {
				jsonMsgs = append(jsonMsgs, msg)
			}

			if s.exportFiles && len(msg.Files) > 0 {
				for _, f := range msg.Files {
					if err := s.downloadFile(conv, f); err != nil {
						fmt.Printf("           Error downloading file: %v\n", err)
					}
				}
			}
		}

		if !history.HasMore {
			break
		}
		if !extraPages {
			fmt.Print("    ")
			extraPages = true
		}
		params.Cursor = history.ResponseMetaData.NextCursor
	}

	if extraPages {
		fmt.Println()
	}

	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Ts.Before(msgs[j].Ts)
	})

	var lastDate *time.Time
	if oldestDate != nil {
		lastDate = oldestDate
	}

	if len(msgs) > 0 {
		if err := s.writeMsgs(conv, msgs); err != nil {
			return nil, err
		}
		t := msgs[len(msgs)-1].Ts.Add(time.Second)
		lastDate = &t

		if s.exportJSON {
			target := filepath.Join(s.exportRoot, fmt.Sprintf("%s-%s_%s.json",
				conv.convType, sanitizeFilename(conv.name), time.Now().Format("2006-01-02-150405")))
			raw, err := json.Marshal(jsonMsgs)
			if err != nil {
				return nil, err
			}
			if err := os.WriteFile(target, raw, 0644); err != nil {
				return nil, err
			}
		}
	}

	fmt.Printf("    Conv %s: %d\n", conv.name, len(msgs))
	return lastDate, nil
}

func (s *APIService) buildSimpleMsg(msg slack.Message) SimpleMsg {
	ts := parseSlackTs(msg.Timestamp)
	name := "unknown"
	if u, ok := s.userMap[msg.User]; ok {
		name = u.Name
	}
	return SimpleMsg{
		Name: name,
		Text: msg.Text,
		Ts:   ts,
	}
}

func parseSlackTs(ts string) time.Time {
	var sec, usec int64
	fmt.Sscanf(ts, "%d.%d", &sec, &usec)
	return time.Unix(sec, usec*1000)
}

func (s *APIService) writeMsgs(conv conversation, msgs []SimpleMsg) error {
	target := filepath.Join(s.exportRoot, fmt.Sprintf("%s-%s_%s.txt",
		conv.convType, sanitizeFilename(conv.name), time.Now().Format("2006-01-02-150405")))

	var lines []string
	for _, m := range msgs {
		line := fmt.Sprintf("%s %s: %s", m.Ts.Format("2006-01-02 15:04:05"), m.Name, m.Text)
		lines = append(lines, line)
	}
	return os.WriteFile(target, []byte(strings.Join(lines, "\r\n")), 0644)
}

func (s *APIService) downloadFile(conv conversation, file slack.File) error {
	convFolder := filepath.Join(s.exportRoot, sanitizeFilename(conv.name))
	if err := os.MkdirAll(convFolder, 0755); err != nil {
		return err
	}

	if file.Name == "" {
		fmt.Printf("           Skipping tombstone file %s\n", file.ID)
		return nil
	}

	target := filepath.Join(convFolder, fmt.Sprintf("%s-%s", file.ID, file.Name))
	fmt.Printf("           Downloading %s\n", target)

	req, err := http.NewRequest("GET", file.URLPrivateDownload, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := os.WriteFile(target, data, 0644); err != nil {
		return err
	}
	fmt.Printf("           Downloaded %s\n", target)
	return nil
}
