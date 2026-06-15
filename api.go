package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// The sachaos CLI has no comment support, so comments go through Todoist's
// REST API (the new /api/v1 endpoints), reusing the token the CLI stored.
const apiBase = "https://api.todoist.com/api/v1"

var cachedToken string

// apiToken reads the API token from the CLI's config (~/.config/todoist/config.json).
func apiToken() (string, error) {
	if cachedToken != "" {
		return cachedToken, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(filepath.Join(home, ".config", "todoist", "config.json"))
	if err != nil {
		return "", fmt.Errorf("could not read todoist token (%v)", err)
	}
	var cfg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return "", err
	}
	if cfg.Token == "" {
		return "", fmt.Errorf("no token in ~/.config/todoist/config.json")
	}
	cachedToken = cfg.Token
	return cachedToken, nil
}

// Comment is a single task comment from the API.
type Comment struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	PostedAt string `json:"posted_at"`
}

func apiDo(method, endpoint string, body io.Reader) ([]byte, error) {
	token, err := apiToken()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, apiBase+endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(data))
		if len(msg) > 160 {
			msg = msg[:160]
		}
		return nil, fmt.Errorf("todoist API %d: %s", resp.StatusCode, msg)
	}
	return data, nil
}

// ListComments returns the comments on a task, oldest first.
func ListComments(taskID string) ([]Comment, error) {
	data, err := apiDo("GET", "/comments?task_id="+url.QueryEscape(taskID), nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Results []Comment `json:"results"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out.Results, nil
}

// AddComment posts a comment to a task.
func AddComment(taskID, content string) error {
	body, _ := json.Marshal(map[string]string{"task_id": taskID, "content": content})
	_, err := apiDo("POST", "/comments", bytes.NewReader(body))
	return err
}

// RecentTaskIDs returns the IDs of the most recently added tasks, newest first.
// It pages through /tasks collecting (id, added_at) and sorts by added_at desc.
func RecentTaskIDs(limit int) ([]string, error) {
	type meta struct {
		ID      string `json:"id"`
		AddedAt string `json:"added_at"`
	}
	var all []meta
	cursor := ""
	for {
		ep := "/tasks?limit=200"
		if cursor != "" {
			ep += "&cursor=" + url.QueryEscape(cursor)
		}
		data, err := apiDo("GET", ep, nil)
		if err != nil {
			return nil, err
		}
		var out struct {
			Results    []meta  `json:"results"`
			NextCursor *string `json:"next_cursor"`
		}
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		all = append(all, out.Results...)
		if out.NextCursor == nil || *out.NextCursor == "" {
			break
		}
		cursor = *out.NextCursor
	}
	// ISO-8601 timestamps sort correctly as strings.
	sort.Slice(all, func(i, j int) bool { return all[i].AddedAt > all[j].AddedAt })
	ids := make([]string, 0, limit)
	for _, mt := range all {
		ids = append(ids, mt.ID)
		if len(ids) >= limit {
			break
		}
	}
	return ids, nil
}

// shortTime turns "2026-06-15T08:21:16.7Z" into "2026-06-15 08:21".
func shortTime(s string) string {
	if len(s) >= 16 {
		return strings.Replace(s[:16], "T", " ", 1)
	}
	return s
}
