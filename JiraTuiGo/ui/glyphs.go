package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"jiratui/themes"
)

// Issue type glyphs
const (
	GlyphBug            = "⊘"
	GlyphTask           = "✓"
	GlyphStory          = "★"
	GlyphEpic           = "⬢"
	GlyphSubtask        = "↳"
	GlyphImprovement    = "⚒"
	GlyphNewFeature     = "✦"
	GlyphTest           = "⌬"
	GlyphIncident       = "‼"
	GlyphServiceRequest = "✉"
	GlyphUnknownType    = "?"
)

// Priority glyphs
const (
	GlyphPriHighest = "⇈"
	GlyphPriHigh    = "▲"
	GlyphPriMedium  = "─"
	GlyphPriLow     = "▼"
	GlyphPriLowest  = "⇊"
)

// Status glyphs
const (
	GlyphStatusTodo       = "○"
	GlyphStatusInProgress = "◐"
	GlyphStatusInReview   = "◑"
	GlyphStatusBlocked    = "✕"
	GlyphStatusDone       = "✓"
	GlyphStatusCancelled  = "⊘"
	GlyphStatusUnknown    = "?"
)

func IssueTypeGlyph(typeName string) (string, tcell.Color) {
	t := themes.Current()
	lower := strings.ToLower(typeName)
	switch {
	case lower == "bug":
		return GlyphBug, themes.C(t.TypeBug)
	case lower == "task":
		return GlyphTask, themes.C(t.TypeTask)
	case lower == "story":
		return GlyphStory, themes.C(t.TypeStory)
	case lower == "epic":
		return GlyphEpic, themes.C(t.TypeEpic)
	case lower == "sub-task" || lower == "subtask":
		return GlyphSubtask, themes.C(t.TypeSubtask)
	case lower == "improvement":
		return GlyphImprovement, themes.C(t.TypeTask)
	case lower == "new feature":
		return GlyphNewFeature, themes.C(t.TypeStory)
	case lower == "test":
		return GlyphTest, themes.C(t.TypeOther)
	case lower == "incident":
		return GlyphIncident, themes.C(t.TypeBug)
	case lower == "service request":
		return GlyphServiceRequest, themes.C(t.TypeTask)
	default:
		return GlyphUnknownType, themes.C(t.TypeOther)
	}
}

func PriorityGlyph(priorityName string) (string, tcell.Color) {
	t := themes.Current()
	lower := strings.ToLower(priorityName)
	switch {
	case lower == "highest" || lower == "critical" || lower == "blocker":
		return GlyphPriHighest, themes.C(t.PriHighest)
	case lower == "high" || lower == "major":
		return GlyphPriHigh, themes.C(t.PriHigh)
	case lower == "medium":
		return GlyphPriMedium, themes.C(t.PriMedium)
	case lower == "low" || lower == "minor":
		return GlyphPriLow, themes.C(t.PriLow)
	case lower == "lowest" || lower == "trivial":
		return GlyphPriLowest, themes.C(t.PriLowest)
	default:
		return GlyphPriMedium, themes.C(t.PriMedium)
	}
}

// StatusGlyph maps a Jira status name to a display glyph.
// Matching is case-insensitive and uses Contains so non-standard status names
// (e.g. "In Testing", "Code Review", "Won't Fix") are handled correctly.
// Color is always tcell.ColorDefault — status glyphs inherit the row
// foreground so they don't clash with theme-specific row colours.
func StatusGlyph(statusName string) (string, tcell.Color) {
	u := strings.ToUpper(strings.TrimSpace(statusName))

	switch {
	case u == "DONE" || u == "CLOSED" || u == "RESOLVED" ||
		u == "COMPLETE" || u == "COMPLETED" || u == "FIXED" ||
		strings.Contains(u, "DEPLOYED"):
		return GlyphStatusDone, tcell.ColorDefault

	case strings.Contains(u, "CANCEL") || strings.Contains(u, "WON'T") ||
		strings.Contains(u, "WONT") || u == "REJECTED" || u == "DUPLICATE":
		return GlyphStatusCancelled, tcell.ColorDefault

	case strings.Contains(u, "BLOCK") || strings.Contains(u, "HOLD") ||
		u == "WAITING" || strings.Contains(u, "STALLED"):
		return GlyphStatusBlocked, tcell.ColorDefault

	case strings.Contains(u, "REVIEW") || u == "QA" ||
		strings.Contains(u, "TEST") || strings.Contains(u, "VERIFY"):
		return GlyphStatusInReview, tcell.ColorDefault

	case strings.Contains(u, "PROGRESS") || u == "DOING" ||
		u == "DEVELOPING" || strings.Contains(u, "WIP"):
		return GlyphStatusInProgress, tcell.ColorDefault

	case u == "TO DO" || u == "TODO" || u == "OPEN" || u == "NEW" ||
		u == "BACKLOG" || u == "READY" || u == "SELECTED FOR DEVELOPMENT":
		return GlyphStatusTodo, tcell.ColorDefault

	default:
		return GlyphStatusUnknown, tcell.ColorDefault
	}
}
