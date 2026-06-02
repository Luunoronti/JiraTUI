package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/config"
	"jiratui/jira"
	"jiratui/themes"
)

type columnDef struct {
	id    string
	width int
}

type IssueList struct {
	*tview.Box
	issues            []jira.Issue
	selected          int
	columns           config.ColumnVisibilityConfig
	app               *tview.Application
	onSelect          func(issue jira.Issue)
	OnSelectionChange func(issue jira.Issue)
	errorMsg          string // non-empty → shown instead of issue rows
}

func NewIssueList(app *tview.Application, cols config.ColumnVisibilityConfig) *IssueList {
	il := &IssueList{
		Box:     tview.NewBox(),
		app:     app,
		columns: cols,
	}
	return il
}

// SetError shows an error message in place of the issue list.
// Call with "" to clear.
func (il *IssueList) SetError(msg string) {
	il.errorMsg = msg
}

func (il *IssueList) SetIssues(issues []jira.Issue) {
	il.errorMsg = ""
	il.issues = issues
	if il.selected >= len(issues) {
		il.selected = len(issues) - 1
	}
	if il.selected < 0 {
		il.selected = 0
	}
	// Notify detail panel of the new selection (first issue after load).
	if il.OnSelectionChange != nil && len(issues) > 0 {
		il.OnSelectionChange(il.issues[il.selected])
	}
}

// SetColumns updates column visibility; takes effect on next Draw.
func (il *IssueList) SetColumns(cols config.ColumnVisibilityConfig) {
	il.columns = cols
}

func (il *IssueList) SetOnSelect(fn func(jira.Issue)) {
	il.onSelect = fn
}

func (il *IssueList) SelectedIssue() *jira.Issue {
	if len(il.issues) == 0 || il.selected < 0 || il.selected >= len(il.issues) {
		return nil
	}
	cp := il.issues[il.selected]
	return &cp
}

func (il *IssueList) SelectedIndex() int {
	return il.selected
}

func (il *IssueList) computeColumns(availWidth int) []columnDef {
	cols := []columnDef{}
	remaining := availWidth

	// Fixed columns
	if il.columns.Key {
		cols = append(cols, columnDef{"key", 10})
		remaining -= 10
	}
	if il.columns.Type {
		cols = append(cols, columnDef{"type", 3})
		remaining -= 3
	}
	if il.columns.Priority {
		cols = append(cols, columnDef{"priority", 3})
		remaining -= 3
	}

	// Reserve minimum for summary
	const summaryMin = 20
	remaining -= summaryMin

	// Optional columns — hide if not enough room
	statusW := 3 // glyph only, like Type and Priority
	assigneeW := 12

	showStatus := il.columns.Status
	showAssignee := il.columns.Assignee

	if showStatus {
		if remaining < statusW {
			showStatus = false
		} else {
			remaining -= statusW
		}
	}
	if showAssignee {
		if remaining < assigneeW {
			showAssignee = false
		} else {
			remaining -= assigneeW
		}
	}

	if showStatus {
		cols = append(cols, columnDef{"status", statusW})
	}
	if showAssignee {
		cols = append(cols, columnDef{"assignee", assigneeW})
	}

	// Summary gets all remaining + summaryMin
	summaryW := summaryMin + remaining
	if summaryW < summaryMin {
		summaryW = summaryMin
	}
	if il.columns.Summary {
		cols = append(cols, columnDef{"summary", summaryW})
	}

	return cols
}

func truncate(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	runes := []rune(s)
	if utf8.RuneCountInString(s) <= maxW {
		return s
	}
	if maxW <= 1 {
		return string(runes[:maxW])
	}
	return string(runes[:maxW-1]) + "…"
}

func pad(s string, width int) string {
	count := utf8.RuneCountInString(s)
	if count >= width {
		return truncate(s, width)
	}
	return s + strings.Repeat(" ", width-count)
}

func (il *IssueList) Draw(screen tcell.Screen) {
	il.Box.DrawForSubclass(screen, il)
	x, y, w, h := il.GetInnerRect()
	if w <= 0 || h <= 0 {
		return
	}
	t := themes.Current()

	if w < MinWidth || h < 3 {
		tview.Print(screen, TooSmallMsg(w, h), x, y, w, tview.AlignLeft, themes.C(t.TextNormal))
		return
	}

	if il.errorMsg != "" {
		tview.Print(screen, " Error: "+il.errorMsg, x, y+1, w, tview.AlignLeft, themes.C(t.PriHighest))
		return
	}

	cols := il.computeColumns(w)

	// Header row
	headerBg := themes.C(t.ListHeaderBg)
	headerFg := themes.C(t.ListHeaderFg)
	drawRow(screen, x, y, w, cols, func(col columnDef) (string, tcell.Color, tcell.Color) {
		var text string
		switch col.id {
		case "key":
			text = "Key"
		case "type":
			text = "T"
		case "priority":
			text = "P"
		case "status":
			text = "S"
		case "assignee":
			text = "Assignee"
		case "summary":
			text = "Summary"
		}
		return pad(text, col.width), headerFg, headerBg
	})

	// "JiraTUI" title right-aligned in the header row.
	{
		title := []rune(" JiraTUI ")
		titleX := x + w - len(title)
		if titleX > x {
			titleStyle := tcell.StyleDefault.
				Foreground(themes.C(t.TextEmphasis)).
				Background(headerBg)
			for j, r := range title {
				if titleX+j >= x+w {
					break
				}
				screen.SetContent(titleX+j, y, r, nil, titleStyle)
			}
		}
	}

	// Separator
	sepY := y + 1
	for i := 0; i < w; i++ {
		screen.SetContent(x+i, sepY, '─', nil, tcell.StyleDefault.
			Foreground(themes.C(t.Border)).Background(themes.C(t.ListBg)))
	}

	// Issue rows
	rowY := y + 2
	maxRows := h - 2
	if len(il.issues) == 0 {
		tview.Print(screen, " No issues found.", x, rowY, w, tview.AlignLeft, themes.C(t.TextMuted))
		return
	}

	// Ensure selected is visible — simple scrolling
	startIdx := 0
	if il.selected >= maxRows {
		startIdx = il.selected - maxRows + 1
	}

	for i := 0; i < maxRows && startIdx+i < len(il.issues); i++ {
		issue := il.issues[startIdx+i]
		isSelected := (startIdx + i) == il.selected

		bgColor := themes.C(t.ListBg)
		fgColor := themes.C(t.ListFg)
		if isSelected {
			bgColor = themes.C(t.ListSelectedBg)
			fgColor = themes.C(t.ListSelectedFg)
		}

		drawRow(screen, x, rowY+i, w, cols, func(col columnDef) (string, tcell.Color, tcell.Color) {
			var text string
			var color tcell.Color = fgColor

			switch col.id {
			case "key":
				text = pad(issue.Key, col.width)
				if !isSelected {
					color = themes.C(t.TextEmphasis)
				}
			case "type":
				glyph, gc := IssueTypeGlyph(issue.IssueType.Name)
				text = pad(glyph, col.width)
				if !isSelected {
					color = gc
				}
			case "priority":
				glyph, gc := PriorityGlyph(issue.Priority.Name)
				text = pad(glyph, col.width)
				if !isSelected {
					color = gc
				}
			case "status":
				// Show glyph (○◐◑✕✓⊘?) — no color override, inherits row fg.
				glyph, _ := StatusGlyph(issue.Status.Name)
				text = pad(glyph, col.width)
			case "assignee":
				name := ""
				if issue.Assignee != nil {
					name = issue.Assignee.DisplayName
				}
				text = pad(truncate(name, col.width), col.width)
			case "summary":
				text = pad(truncate(issue.Summary, col.width), col.width)
			}
			return text, color, bgColor
		})
	}
}

func drawRow(screen tcell.Screen, x, y, w int, cols []columnDef,
	getCellContent func(columnDef) (string, tcell.Color, tcell.Color)) {

	cx := x
	for _, col := range cols {
		text, fg, bg := getCellContent(col)
		style := tcell.StyleDefault.Foreground(fg).Background(bg)

		runes := []rune(text)
		for i, r := range runes {
			if cx+i >= x+w {
				break
			}
			screen.SetContent(cx+i, y, r, nil, style)
		}
		// Fill to col width
		for i := len(runes); i < col.width; i++ {
			if cx+i >= x+w {
				break
			}
			screen.SetContent(cx+i, y, ' ', nil, style)
		}
		cx += col.width
	}
	// Fill remaining
	bg := tcell.ColorDefault
	if len(cols) > 0 {
		_, _, bg = getCellContent(cols[0])
	}
	for cx < x+w {
		screen.SetContent(cx, y, ' ', nil, tcell.StyleDefault.Background(bg))
		cx++
	}
}

func (il *IssueList) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return il.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			if il.selected > 0 {
				il.selected--
				if il.onSelect != nil && len(il.issues) > 0 {
					il.onSelect(il.issues[il.selected])
				}
				if il.OnSelectionChange != nil && len(il.issues) > 0 {
					il.OnSelectionChange(il.issues[il.selected])
				}
			}
		case tcell.KeyDown:
			if il.selected < len(il.issues)-1 {
				il.selected++
				if il.onSelect != nil && len(il.issues) > 0 {
					il.onSelect(il.issues[il.selected])
				}
				if il.OnSelectionChange != nil && len(il.issues) > 0 {
					il.OnSelectionChange(il.issues[il.selected])
				}
			}
		case tcell.KeyEnter:
			// Enter is handled by the app via the global key handler;
			// we fire OnSelectionChange so the fullscreen detail can be
			// populated before the page switch.
			if il.OnSelectionChange != nil && len(il.issues) > 0 {
				il.OnSelectionChange(il.issues[il.selected])
			}
		}
	})
}

// UpdateIssue replaces the in-memory copy of an issue if it exists in the list.
func (il *IssueList) UpdateIssue(updated jira.Issue) {
	for i, iss := range il.issues {
		if iss.Key == updated.Key {
			il.issues[i] = updated
			return
		}
	}
}

// statusLine returns a short summary for the status bar
func (il *IssueList) statusLine() string {
	if len(il.issues) == 0 {
		return "0 issues"
	}
	return fmt.Sprintf("%d/%d issues", il.selected+1, len(il.issues))
}
