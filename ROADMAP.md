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

## How this list works

- Items here are intentions, not promises — order and scope can change.
- When a feature ships, move its summary into `RELEASE_NOTES.md` under the
  version that shipped it and drop it from this list.
