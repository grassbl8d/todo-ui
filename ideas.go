package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Idea is a quick thought captured locally (not synced to Todoist, for now).
type Idea struct {
	Text string `json:"text"`
	At   string `json:"at"` // RFC3339 capture time
}

func ideasPath() string {
	d := stateDir()
	if d == "" {
		return ""
	}
	return filepath.Join(d, "ideas.json")
}

// LoadIdeas reads captured ideas, newest first.
func LoadIdeas() []Idea {
	b, err := os.ReadFile(ideasPath())
	if err != nil {
		return nil
	}
	var ideas []Idea
	if json.Unmarshal(b, &ideas) != nil {
		return nil
	}
	return ideas
}

// SaveIdeas persists the ideas list.
func SaveIdeas(ideas []Idea) {
	if p := ideasPath(); p != "" {
		if b, err := json.Marshal(ideas); err == nil {
			_ = os.WriteFile(p, b, 0o600)
		}
	}
}

// addIdea returns the list with a new idea prepended (newest first).
func addIdea(ideas []Idea, text string) []Idea {
	return append([]Idea{{Text: text, At: nowStamp()}}, ideas...)
}
