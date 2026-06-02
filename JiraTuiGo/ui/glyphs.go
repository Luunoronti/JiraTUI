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

func StatusGlyph(statusName string) (string, tcell.Color) {
	t := themes.Current()
	lower := strings.ToLower(statusName)
	switch {
	case lower == "to do" || lower == "open" || lower == "backlog":
		return GlyphStatusTodo, themes.C(t.StatusTodo)
	case lower == "in progress":
		return GlyphStatusInProgress, themes.C(t.StatusInProgress)
	case lower == "in review" || lower == "testing" || lower == "qa":
		return GlyphStatusInReview, themes.C(t.StatusInReview)
	case lower == "blocked" || lower == "on hold":
		return GlyphStatusBlocked, themes.C(t.StatusBlocked)
	case lower == "done" || lower == "closed" || lower == "resolved":
		return GlyphStatusDone, themes.C(t.StatusDone)
	case lower == "cancelled" || lower == "won't do":
		return GlyphStatusCancelled, themes.C(t.StatusCancelled)
	default:
		return GlyphStatusUnknown, themes.C(t.StatusTodo)
	}
}
