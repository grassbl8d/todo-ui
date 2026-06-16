package main

import "testing"

func TestSortByDateAdded(t *testing.T) {
	m := newTestModel()
	m.cache = newCache()
	m.cache.Items["a"] = apiItem{ID: "a", AddedAt: "2026-01-01T00:00:00Z"} // oldest
	m.cache.Items["b"] = apiItem{ID: "b", AddedAt: "2026-06-01T00:00:00Z"} // newest
	m.cache.Items["c"] = apiItem{ID: "c", AddedAt: "2026-03-01T00:00:00Z"} // middle

	ts := []Task{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	m.sortMode = sortAdded

	// Default direction: newest first.
	m.sortDesc = false
	m.sortTasks(ts)
	if got := []string{ts[0].ID, ts[1].ID, ts[2].ID}; got[0] != "b" || got[1] != "c" || got[2] != "a" {
		t.Fatalf("newest-first order wrong: %v", got)
	}

	// Reversed: oldest first.
	m.sortDesc = true
	m.sortTasks(ts)
	if got := []string{ts[0].ID, ts[1].ID, ts[2].ID}; got[0] != "a" || got[1] != "c" || got[2] != "b" {
		t.Fatalf("oldest-first order wrong: %v", got)
	}

	if sortAdded.label() != "date added" {
		t.Fatalf("label = %q", sortAdded.label())
	}
}
