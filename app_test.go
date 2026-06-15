package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestParseCSV(t *testing.T) {
	csv := "ID,Priority,DueDate,Project,Labels,Content\n" +
		"abc123,p1,25/06/05(Thu) 09:00,#Bills Payments,@bills-payment,Pay Globe\n" +
		"def456,p4,,#Personal,,Read a book\n"
	rows, err := parseCSV(csv)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "abc123" || rows[0][5] != "Pay Globe" {
		t.Fatalf("bad row0: %#v", rows[0])
	}
	if rows[1][4] != "" || rows[1][5] != "Read a book" {
		t.Fatalf("bad row1: %#v", rows[1])
	}
}

func sampleTasks() []Task {
	return []Task{
		{ID: "1", Priority: "p1", Project: "#Bills", Content: "Pay Globe", DueDate: "today"},
		{ID: "2", Priority: "p4", Project: "#Personal", Content: "Read a book"},
	}
}

func TestTasksLoadedPopulatesList(t *testing.T) {
	m := initialModel()
	m.width, m.height = 100, 40
	m.list.SetSize(100, 36)

	nm, _ := m.Update(tasksLoadedMsg{tasks: sampleTasks(), filter: ""})
	mm := nm.(model)
	if n := len(mm.list.Items()); n != 2 {
		t.Fatalf("want 2 items, got %d", n)
	}
	if mm.loading {
		t.Fatal("should not be loading after load")
	}
	if mm.status != "2 tasks" {
		t.Fatalf("status = %q", mm.status)
	}
}

func TestKeyOpensProjectPicker(t *testing.T) {
	m := initialModel()
	m.width, m.height = 100, 40
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if nm.(model).mode != modeProjectPick {
		t.Fatal("pressing 'a' should open the project picker")
	}
}

func TestProjectPickThenAdd(t *testing.T) {
	m := initialModel()
	m.lastProject = Project{} // deterministic: ignore any persisted last project
	m.width, m.height = 100, 40
	m.projList.SetSize(100, 36)
	// load some projects
	nm, _ := m.Update(projectsLoadedMsg{projects: []Project{
		{ID: "p1", Name: "#Personal"},
		{ID: "p2", Name: "#Bills Payments"},
	}})
	m = nm.(model)
	// open picker
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = nm.(model)
	if m.mode != modeProjectPick {
		t.Fatal("expected project pick mode")
	}
	// select first project
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	if m.mode != modeAdd {
		t.Fatal("selecting a project should move to add mode")
	}
	if m.addProject.ID != "p1" {
		t.Fatalf("addProject = %q, want p1", m.addProject.ID)
	}
	if m.lastProject.ID != "p1" {
		t.Fatalf("lastProject should be remembered, got %q", m.lastProject.ID)
	}
}

func TestAddToLastProjectShortcut(t *testing.T) {
	m := initialModel()
	m.width, m.height = 100, 40
	m.lastProject = Project{ID: "p9", Name: "#Work"}
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")})
	mm := nm.(model)
	if mm.mode != modeAdd {
		t.Fatal("'A' with a last project should jump straight to add mode")
	}
	if mm.addProject.ID != "p9" {
		t.Fatalf("addProject = %q, want p9", mm.addProject.ID)
	}
}

func TestKeyOpensSearchMode(t *testing.T) {
	m := initialModel()
	m.width, m.height = 100, 40
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	if nm.(model).mode != modeSearch {
		t.Fatal("pressing '/' should enter search mode")
	}
}

func TestSearchEnterSetsFilterAndReloads(t *testing.T) {
	m := initialModel()
	m.width, m.height = 100, 40
	// enter search mode
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = nm.(model)
	// type "today"
	for _, r := range "today" {
		nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = nm.(model)
	}
	// submit
	nm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	if m.mode != modeList {
		t.Fatal("enter should return to list mode")
	}
	if m.filter != "today" {
		t.Fatalf("filter = %q, want today", m.filter)
	}
	if cmd == nil {
		t.Fatal("expected a reload command")
	}
}

func TestIsFilterExpr(t *testing.T) {
	filters := []string{"today", "today | overdue", "#Personal", "@label", "p1", "7 days", "no date", "overdue & p1"}
	for _, f := range filters {
		if !isFilterExpr(f) {
			t.Errorf("%q should be detected as a filter expression", f)
		}
	}
	texts := []string{"anvaya", "pay globe", "groceries", "call mom", "anvaya golf"}
	for _, s := range texts {
		if isFilterExpr(s) {
			t.Errorf("%q should be treated as plain text search", s)
		}
	}
}

func TestLocalTextSearchFiltersTasks(t *testing.T) {
	m := initialModel()
	m.width, m.height = 100, 40
	m.list.SetSize(100, 36)
	tasks := []Task{
		{ID: "1", Priority: "p4", Project: "#Bills", Content: "Pay anvaya golf dues"},
		{ID: "2", Priority: "p4", Project: "#Bills", Content: "Pay Globe Anvaya"},
		{ID: "3", Priority: "p4", Project: "#Personal", Content: "Read a book"},
	}
	nm, _ := m.Update(tasksLoadedMsg{tasks: tasks})
	m = nm.(model)
	if len(m.list.Items()) != 3 {
		t.Fatalf("want 3 items initially, got %d", len(m.list.Items()))
	}
	// open search and type a plain word
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = nm.(model)
	for _, r := range "anvaya" {
		nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = nm.(model)
	}
	// live preview should already narrow to the 2 anvaya tasks
	if got := len(m.list.Items()); got != 2 {
		t.Fatalf("live text search: want 2 items, got %d", got)
	}
	// submit (enter) — stays local, no server reload
	nm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	if m.mode != modeList {
		t.Fatal("enter should return to list mode")
	}
	if m.textQuery != "anvaya" {
		t.Fatalf("textQuery = %q", m.textQuery)
	}
	if m.filter != "" {
		t.Fatalf("plain text search must not set a server filter, got %q", m.filter)
	}
	if cmd != nil {
		t.Fatal("plain text search should not trigger a server reload")
	}
	if len(m.list.Items()) != 2 {
		t.Fatalf("want 2 matched items, got %d", len(m.list.Items()))
	}
}

func TestViewByProject(t *testing.T) {
	m := initialModel()
	m.lastProject = Project{} // deterministic
	m.width, m.height = 100, 40
	m.list.SetSize(100, 36)
	m.projList.SetSize(100, 36)
	// projects + tasks
	nm, _ := m.Update(projectsLoadedMsg{projects: []Project{
		{ID: "b", Name: "#Bills"},
		{ID: "p", Name: "#Personal"},
	}})
	m = nm.(model)
	nm, _ = m.Update(tasksLoadedMsg{tasks: []Task{
		{ID: "1", Project: "#Bills", Content: "Pay Globe"},
		{ID: "2", Project: "#Bills", Content: "Pay Cignal"},
		{ID: "3", Project: "#Personal", Content: "Read a book"},
	}})
	m = nm.(model)
	// open view-by-project
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = nm.(model)
	if m.mode != modeProjectPick || m.pickIntent != pickView {
		t.Fatal("'p' should open the picker in view intent")
	}
	// index 0 is the "All Projects" entry; move down to #Bills
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = nm.(model)
	nm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	if m.mode != modeList {
		t.Fatal("selecting should return to list")
	}
	if m.projectView != "#Bills" {
		t.Fatalf("projectView = %q, want #Bills", m.projectView)
	}
	if cmd != nil {
		t.Fatal("view-by-project is local; should not reload from server")
	}
	if got := len(m.list.Items()); got != 2 {
		t.Fatalf("want 2 #Bills tasks, got %d", got)
	}

	// Reopen the picker and choose "All Projects" (index 0) to go back.
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = nm.(model)
	m.projList.Select(0) // ensure cursor on the All Projects entry
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	if m.projectView != "" {
		t.Fatalf("selecting All Projects should clear projectView, got %q", m.projectView)
	}
	if got := len(m.list.Items()); got != 3 {
		t.Fatalf("All Projects should show all 3 tasks, got %d", got)
	}

	// Re-apply a project view, then verify Esc also clears it.
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = nm.(model)
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = nm.(model)
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	if m.projectView != "#Bills" {
		t.Fatalf("expected #Bills view, got %q", m.projectView)
	}
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = nm.(model)
	if m.projectView != "" {
		t.Fatal("esc should clear projectView")
	}
	if got := len(m.list.Items()); got != 3 {
		t.Fatalf("after clear want 3 tasks, got %d", got)
	}
}

func TestBackAndHomeNavigation(t *testing.T) {
	m := initialModel()
	m.lastProject = Project{}
	m.width, m.height = 100, 40
	m.list.SetSize(100, 36)
	m.projList.SetSize(100, 36)
	nm, _ := m.Update(projectsLoadedMsg{projects: []Project{{ID: "b", Name: "#Bills"}}})
	m = nm.(model)
	nm, _ = m.Update(tasksLoadedMsg{tasks: []Task{
		{ID: "1", Project: "#Bills", Content: "Pay Globe anvaya"},
		{ID: "2", Project: "#Bills", Content: "Pay Cignal"},
		{ID: "3", Project: "#Personal", Content: "Read a book"},
	}})
	m = nm.(model)

	// View #Bills
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = nm.(model)
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = nm.(model)
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	if m.projectView != "#Bills" || len(m.list.Items()) != 2 {
		t.Fatalf("expected #Bills view with 2 items, got %q/%d", m.projectView, len(m.list.Items()))
	}

	// Then search "anvaya" within it
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = nm.(model)
	for _, r := range "anvaya" {
		nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = nm.(model)
	}
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	if m.textQuery != "anvaya" || m.projectView != "#Bills" || len(m.list.Items()) != 1 {
		t.Fatalf("search-in-project: got query=%q project=%q items=%d", m.textQuery, m.projectView, len(m.list.Items()))
	}

	// Back → returns to #Bills view (no search)
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m = nm.(model)
	if m.textQuery != "" || m.projectView != "#Bills" || len(m.list.Items()) != 2 {
		t.Fatalf("after back: query=%q project=%q items=%d", m.textQuery, m.projectView, len(m.list.Items()))
	}

	// Back again → all tasks
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	m = nm.(model)
	if m.projectView != "" || len(m.list.Items()) != 3 {
		t.Fatalf("after second back: project=%q items=%d", m.projectView, len(m.list.Items()))
	}

	// Home from a project view jumps straight to all tasks
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	m = nm.(model)
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = nm.(model)
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(model)
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m = nm.(model)
	if m.projectView != "" || len(m.list.Items()) != 3 {
		t.Fatalf("home: project=%q items=%d", m.projectView, len(m.list.Items()))
	}
}

func TestHelpPageToggle(t *testing.T) {
	m := initialModel()
	m.width, m.height = 100, 40
	m.list.SetSize(100, 36)
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("H")})
	m = nm.(model)
	if m.mode != modeHelp {
		t.Fatal("H should open the help page")
	}
	if !strings.Contains(m.View(), "Navigation") {
		t.Fatal("help view should render its sections")
	}
	// any key closes
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if nm.(model).mode != modeList {
		t.Fatal("any key should close help")
	}
}

func TestDeleteConfirmFlow(t *testing.T) {
	m := initialModel()
	m.width, m.height = 100, 40
	m.list.SetSize(100, 36)
	nm, _ := m.Update(tasksLoadedMsg{tasks: sampleTasks()})
	m = nm.(model)
	// press d -> confirm mode
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = nm.(model)
	if m.mode != modeConfirm {
		t.Fatal("d should open confirm")
	}
	// press n -> cancel back to list
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if nm.(model).mode != modeList {
		t.Fatal("n should cancel delete")
	}
}

func TestViewRendersWithoutPanic(t *testing.T) {
	m := initialModel()
	m.width, m.height = 100, 40
	m.list.SetSize(100, 36)
	nm, _ := m.Update(tasksLoadedMsg{tasks: sampleTasks()})
	out := nm.(model).View()
	if out == "" {
		t.Fatal("view should render content")
	}
}
