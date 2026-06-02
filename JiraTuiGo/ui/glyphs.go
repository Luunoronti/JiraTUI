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

// Visibility glyphs
const (
	GlyphHidden = "⊙" // eye/target symbol for hidden issues
)

// Status glyphs
const (
	GlyphStatusTodo       = "○"
	GlyphStatusQueued     = "▷" // ready/queued — between backlog and active work
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
// StatusGlyph maps a Jira status name to a display glyph.
// Order matters — earlier cases take priority. "Ready For Review" matches
// REVIEW before READY, so it gets ◑ not ○.
func StatusGlyph(statusName string) (string, tcell.Color) {
	u := strings.ToUpper(strings.TrimSpace(statusName))

	switch {
	// ✓ Done — completed, released, shipped, merged…
	case u == "DONE" || u == "CLOSED" || u == "RESOLVED" ||
		u == "COMPLETE" || u == "COMPLETED" || u == "FIXED" ||
		u == "RELEASED" || u == "SHIPPED" || u == "DELIVERED" ||
		u == "MERGED" || u == "ACCEPTED" || u == "VERIFIED" ||
		strings.Contains(u, "DEPLOYED") || strings.Contains(u, "RELEASED"):
		return GlyphStatusDone, tcell.ColorDefault

	// ⊘ Cancelled — won't do, archived, rejected, invalid…
	case strings.Contains(u, "CANCEL") || strings.Contains(u, "WON'T") ||
		strings.Contains(u, "WONT") || strings.Contains(u, "ARCHIVE") ||
		u == "REJECTED" || u == "DUPLICATE" || u == "DECLINED" ||
		u == "NOT A BUG" || u == "INVALID" || u == "ABANDONED":
		return GlyphStatusCancelled, tcell.ColorDefault

	// ✕ Blocked — on hold, suspended, paused, impediment…
	case strings.Contains(u, "BLOCK") || strings.Contains(u, "HOLD") ||
		strings.Contains(u, "STALL") || strings.Contains(u, "SUSPEND") ||
		strings.Contains(u, "PAUSE") || strings.Contains(u, "IMPEDIMENT") ||
		u == "WAITING":
		return GlyphStatusBlocked, tcell.ColorDefault

	// ◑ In Review / Testing — checked before READY so "Ready For Review" → ◑
	case strings.Contains(u, "REVIEW") || strings.Contains(u, "TEST") ||
		strings.Contains(u, "VERIFY") || strings.Contains(u, "VALIDAT") ||
		strings.Contains(u, "AUDIT") || strings.Contains(u, "INSPECT") ||
		u == "QA" || strings.Contains(u, "QA ") || strings.Contains(u, " QA"):
		return GlyphStatusInReview, tcell.ColorDefault

	// ▷ Queued / Ready — selected and waiting to start active work.
	// Checked BEFORE In Progress so "Ready For Build" → ▷, not ◐.
	case strings.Contains(u, "READY") || strings.Contains(u, "SELECTED") ||
		strings.Contains(u, "APPROVED") || strings.Contains(u, "SCHEDULED") ||
		strings.Contains(u, "GROOMED") || strings.Contains(u, "REFINEMENT") ||
		strings.Contains(u, "PRIORITI") || strings.Contains(u, "PENDING"):
		return GlyphStatusQueued, tcell.ColorDefault

	// ◐ In Progress — active work, building, deploying, analysis…
	case strings.Contains(u, "PROGRESS") || strings.Contains(u, "DOING") ||
		strings.Contains(u, "DEVELOP") || strings.Contains(u, "WIP") ||
		strings.Contains(u, "BUILD") || strings.Contains(u, "DEPLOY") ||
		strings.Contains(u, "IMPLEMENT") || strings.Contains(u, "ANALYS") ||
		strings.Contains(u, "CODING") || strings.Contains(u, "WORKING"):
		return GlyphStatusInProgress, tcell.ColorDefault

	// ○ To Do — pure backlog, nothing selected yet
	case u == "TO DO" || u == "TODO" || u == "OPEN" || u == "NEW" ||
		u == "BACKLOG" || u == "PLANNED":
		return GlyphStatusTodo, tcell.ColorDefault

	default:
		return GlyphStatusUnknown, tcell.ColorDefault
	}
}
