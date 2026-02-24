package internal

import "fmt"

// toolCategory maps a Claude CLI tool name to an internal category key.
var toolCategory = map[string]string{
	"WebSearch": "web",
	"WebFetch":  "web",
	"Read":      "read",
	"Edit":      "write",
	"Write":     "write",
	"Bash":      "bash",
	"Grep":      "search",
	"Glob":      "search",
	"Task":      "subagent",
}

type categoryInfo struct {
	emoji      string
	activeText string
	doneText   string
}

var categories = map[string]categoryInfo{
	"web":      {"🔍", "Searching the web", "Searched the web"},
	"read":     {"📄", "Reading files", "Read files"},
	"write":    {"✏️", "Editing files", "Edited files"},
	"bash":     {"⚡", "Running command", "Ran commands"},
	"search":   {"🔎", "Searching codebase", "Searched codebase"},
	"subagent": {"🤖", "Running sub-agent", "Ran sub-agent"},
	"unknown":  {"⚙️", "Working", "Worked"},
}

type statusTracker struct {
	order  []string         // categories in first-seen order
	counts map[string]int   // count per category
}

func newStatusTracker() *statusTracker {
	return &statusTracker{
		counts: make(map[string]int),
	}
}

// Add maps a tool name to its category, updates counts, and returns true if this is the first event overall.
func (s *statusTracker) Add(toolName string) bool {
	cat, ok := toolCategory[toolName]
	if !ok {
		cat = "unknown"
	}

	first := len(s.order) == 0

	if s.counts[cat] == 0 {
		s.order = append(s.order, cat)
	}
	s.counts[cat]++

	return first
}

// Render returns the status text while the agent is still running.
// All lines except the last are in past tense; the last line is active with "...".
func (s *statusTracker) Render() string {
	if len(s.order) == 0 {
		return ""
	}

	var text string
	for i, cat := range s.order {
		info := categories[cat]
		count := s.counts[cat]
		if i < len(s.order)-1 {
			text += info.emoji + " " + countLabel(info.doneText, count) + "\n"
		} else {
			text += info.emoji + " " + countLabel(info.activeText, count) + "..."
		}
	}
	return text
}

// RenderDone returns all lines in past tense — no suffix. Used as a base for final/cancel/error states.
func (s *statusTracker) RenderDone() string {
	if len(s.order) == 0 {
		return ""
	}

	var text string
	for _, cat := range s.order {
		info := categories[cat]
		text += info.emoji + " " + countLabel(info.doneText, s.counts[cat]) + "\n"
	}
	return text
}

// RenderFinal returns all lines in past tense with a "✅ Done" suffix.
func (s *statusTracker) RenderFinal() string {
	return s.RenderDone() + "✅ Done"
}

func countLabel(base string, count int) string {
	if count > 1 {
		return fmt.Sprintf("%s (%d)", base, count)
	}
	return base
}
