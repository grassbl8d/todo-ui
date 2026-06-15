package main

import (
	"os"
	"path/filepath"
	"strings"
)

// stateDir returns ~/.config/todoui, creating it if needed.
func stateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(home, ".config", "todoui")
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

func stateFile() string {
	d := stateDir()
	if d == "" {
		return ""
	}
	return filepath.Join(d, "recent_projects.txt")
}

// maxRecentProjects is how many recently-chosen projects we remember/show.
const maxRecentProjects = 3

// LoadRecentProjects reads the recently-chosen projects, most recent first.
func LoadRecentProjects() []Project {
	f := stateFile()
	if f == "" {
		return nil
	}
	b, err := os.ReadFile(f)
	if err != nil {
		return nil
	}
	var out []Project
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "\t", 2)
		if len(parts) == 2 && parts[0] != "" {
			out = append(out, Project{ID: parts[0], Name: parts[1]})
		}
	}
	if len(out) > maxRecentProjects {
		out = out[:maxRecentProjects]
	}
	return out
}

// SaveRecentProjects persists the recent-projects list (most recent first).
func SaveRecentProjects(ps []Project) {
	f := stateFile()
	if f == "" {
		return
	}
	var b strings.Builder
	for i, p := range ps {
		if i >= maxRecentProjects {
			break
		}
		b.WriteString(p.ID + "\t" + p.Name + "\n")
	}
	_ = os.WriteFile(f, []byte(b.String()), 0o644)
}

// pushRecentProject returns ps with p moved to the front, deduped and capped.
func pushRecentProject(ps []Project, p Project) []Project {
	if p.ID == "" {
		return ps
	}
	out := []Project{p}
	for _, x := range ps {
		if x.ID != p.ID {
			out = append(out, x)
		}
	}
	if len(out) > maxRecentProjects {
		out = out[:maxRecentProjects]
	}
	return out
}
