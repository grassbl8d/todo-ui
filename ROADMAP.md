# Roadmap

Planned features for todo-ui. This is a living list — items aren't committed to a
specific release. Shipped features are recorded in
[`RELEASE_NOTES.md`](RELEASE_NOTES.md).

## Planned

### ⏱ Pomodoro
A built-in Pomodoro timer for focused work on the selected (or pinned) task:
configurable work/break lengths, a visible countdown in the UI, and a chime/flash
when an interval ends. Pairs naturally with pin/focus mode.

### ⏳ Countdown timer
A general-purpose countdown (set a duration, watch it tick down) independent of
the Pomodoro cycle — e.g. timeboxing a quick task or a meeting. Lightweight,
dismissible, shown alongside the task list.

### 🔁 Support for recurring
Fuller support for Todoist recurring tasks: surface the recurrence rule in the
detail view, complete-and-reschedule a recurring task to its next occurrence
(instead of just marking it done), and preserve the schedule on edits. (Recurring
due strings already round-trip through the due picker; this is about first-class
handling.)

### 🖼 Support for image
Display image attachments on a task — render inline previews where the terminal
supports it (e.g. iTerm2/Kitty graphics protocols), with a text fallback (link +
filename) elsewhere. Read-only first; uploading is a later step.

### 🗺 Mind-map node scheduling
Let mind-map nodes carry the same task metadata as real tasks: a **due date**, a
**deadline**, and a **priority** (P1–P4), set right in the map with the same
quick-pickers the task detail uses. When a node is converted to / linked with a
Todoist task (`T`), these round-trip through sync; until then they're stored
locally on the node and shown on its box.

### 🔄 Mind-map online sync
Sync ideas and their mind maps to the cloud so they're not local-only — survive a
cache wipe and follow you across machines. Likely stored as Todoist task
metadata/comments (or a dedicated project) so it rides the existing Sync API,
with the same offline-first queue the tasks use.

### 🗂 Group by mode
An explicit grouping mode for the task list, separate from sorting: group tasks
by project, priority, label, due/added date, etc., with a real section header per
group and a count. Builds on the existing sort subheaders — adds collapsible
groups (fold/unfold a section) and lets you pick the grouping field independently
of the sort order. The grouping field will be chosen with **`Ctrl`+a number**
(mirroring the `1`–`7` sort keys), so e.g. `Ctrl+4` groups by project while the
`4` sort still orders within each group.

## How this list works

- Items here are intentions, not promises — order and scope can change.
- When a feature ships, move its summary into `RELEASE_NOTES.md` under the
  version that shipped it and drop it from this list.
