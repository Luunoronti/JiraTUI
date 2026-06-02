package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/jira"
	"jiratui/themes"
)

// ─── word wrap ───────────────────────────────────────────────────────────────

// wordWrap wraps text at word boundaries to fit within width columns.
// Lines already shorter than width are left as-is; existing newlines are
// honoured and each resulting line is then wrapped independently.
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		words := strings.Fields(line)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		cur := ""
		for _, w := range words {
			// If the single word is wider than width, just emit it as-is.
			if cur == "" {
				cur = w
			} else if len([]rune(cur))+1+len([]rune(w)) <= width {
				cur += " " + w
			} else {
				out = append(out, cur)
				cur = w
			}
		}
		if cur != "" {
			out = append(out, cur)
		}
	}
	return strings.Join(out, "\n")
}

// ─── renderIssue ─────────────────────────────────────────────────────────────

func renderIssue(issue jira.Issue, width int) string {
	if width < 4 {
		width = 4
	}
	sep := strings.Repeat("─", width)

	var sb strings.Builder

	sb.WriteString(issue.Key + "\n")
	sb.WriteString(sep + "\n")

	// Type · Priority · Status
	typePriStatus := issue.IssueType.Name + " · " + issue.Priority.Name + " · " + issue.Status.Name
	sb.WriteString(typePriStatus + "\n")

	// Assignee
	assignee := "(unassigned)"
	if issue.Assignee != nil && issue.Assignee.DisplayName != "" {
		assignee = issue.Assignee.DisplayName
	}
	sb.WriteString(fmt.Sprintf("Assignee:  %s\n", assignee))

	// Reporter
	reporter := "(none)"
	if issue.Reporter != nil && issue.Reporter.DisplayName != "" {
		reporter = issue.Reporter.DisplayName
	}
	sb.WriteString(fmt.Sprintf("Reporter:  %s\n", reporter))

	// Updated
	if !issue.Updated.IsZero() {
		sb.WriteString(fmt.Sprintf("Updated:   %s\n", issue.Updated.Format("2006-01-02 15:04")))
	}

	// Labels
	if len(issue.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("Labels:    %s\n", strings.Join(issue.Labels, ", ")))
	}

	// Sprint
	if issue.Sprint != "" {
		sb.WriteString(fmt.Sprintf("Sprint:    %s\n", issue.Sprint))
	}

	sb.WriteString("\n")

	// Description
	if issue.Description != "" {
		sb.WriteString("Description:\n")
		sb.WriteString(wordWrap(issue.Description, width) + "\n")
	}

	// Comments
	for _, c := range issue.Comments {
		sb.WriteString("\n")
		// Comment header: ── Author · Date ───
		author := c.Author.DisplayName
		date := c.Created.Format("2006-01-02 15:04")
		header := fmt.Sprintf("── %s · %s ───", author, date)
		sb.WriteString(header + "\n")
		sb.WriteString(wordWrap(c.Body, width) + "\n")
	}

	return sb.String()
}

// ─── DetailPanel (side panel) ────────────────────────────────────────────────

// DetailPanel is a floating overlay anchored top-right. It does NOT steal
// focus; the issue list keeps focus while the panel is visible.
// Like NavPanel it self-positions inside Draw() so it reflows on every resize.
type DetailPanel struct {
	*tview.Box
	textView *tview.TextView
	content  string // current rendered text; updated via SetIssue
}

func NewDetailPanel() *DetailPanel {
	dp := &DetailPanel{
		Box:      tview.NewBox(),
		textView: tview.NewTextView(),
	}
	dp.textView.SetDynamicColors(false)
	dp.textView.SetScrollable(true)
	dp.textView.SetWrap(true)
	dp.textView.SetWordWrap(true)
	dp.refreshColors()
	return dp
}

func (dp *DetailPanel) refreshColors() {
	t := themes.Current()
	dp.Box.SetBackgroundColor(themes.C(t.DetailBg))
	dp.Box.SetBorderColor(themes.C(t.BorderFocused))
	dp.Box.SetTitleColor(themes.C(t.TextEmphasis))
	dp.textView.SetBackgroundColor(themes.C(t.DetailBg))
	dp.textView.SetTextColor(themes.C(t.DetailFg))
}

// SetIssue updates the displayed issue. Must be called from outside Draw().
func (dp *DetailPanel) SetIssue(issue *jira.Issue, panelWidth int) {
	if issue == nil {
		dp.content = ""
		dp.textView.SetText("")
		return
	}
	// Inner width = panelWidth - 2 (border) - 2 (padding)
	innerW := panelWidth - 4
	if innerW < 4 {
		innerW = 4
	}
	dp.content = renderIssue(*issue, innerW)
	dp.textView.SetText(dp.content)
	dp.textView.ScrollToBeginning()
}

func (dp *DetailPanel) Draw(screen tcell.Screen) {
	t := themes.Current()
	termW, termH := screen.Size()

	tier := SizeTier(termW)
	panelW := DetailWidth(tier, termW)
	if panelW <= 0 {
		return
	}
	panelH := termH - 2 // y=1 to termH-2

	panelX := termW - panelW
	panelY := 1

	// Self-position: anchored top-right.
	dp.Box.SetRect(panelX, panelY, panelW, panelH)
	dp.Box.SetBorder(true)
	dp.Box.SetTitle(" Detail ")

	dp.Box.DrawForSubclass(screen, dp)

	innerX, innerY, innerW, innerH := dp.GetInnerRect()
	if innerW <= 0 || innerH <= 0 {
		return
	}

	// Position the embedded TextView inside the inner rect and draw it.
	dp.textView.SetRect(innerX, innerY, innerW, innerH)

	// Draw the text view content manually line by line from its backing text.
	// We use the textView to handle scrolling state; actual rendering we do
	// ourselves to avoid tview's color tag processing interfering with plain text.
	bg := themes.C(t.DetailBg)
	fg := themes.C(t.DetailFg)
	style := tcell.StyleDefault.Foreground(fg).Background(bg)

	lines := strings.Split(dp.content, "\n")
	// Simple scroll offset from the textView's row offset.
	rowOff, _ := dp.textView.GetScrollOffset()

	for row := 0; row < innerH; row++ {
		lineIdx := rowOff + row
		y := innerY + row

		// Blank the row first.
		for cx := innerX; cx < innerX+innerW; cx++ {
			screen.SetContent(cx, y, ' ', nil, style)
		}

		if lineIdx >= len(lines) {
			continue
		}
		line := []rune(lines[lineIdx])
		for col, r := range line {
			if col >= innerW {
				break
			}
			screen.SetContent(innerX+col, y, r, nil, style)
		}
	}
}

// InputHandler for the detail panel (handles scrolling when focused, though
// we try not to give it focus for the side panel use-case).
func (dp *DetailPanel) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return dp.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		dp.textView.InputHandler()(event, setFocus)
	})
}

// ─── DetailFullView (fullscreen view) ────────────────────────────────────────

// DetailFullView wraps a tview.TextView for fullscreen issue display.
// It receives focus so the user can scroll with arrow keys.
type DetailFullView struct {
	*tview.TextView
}

func NewDetailFullView() *DetailFullView {
	tv := tview.NewTextView()
	tv.SetDynamicColors(false)
	tv.SetScrollable(true)
	tv.SetWrap(true)
	tv.SetWordWrap(true)
	tv.SetBorder(true)
	tv.SetTitle(" Detail ")
	t := themes.Current()
	tv.SetBackgroundColor(themes.C(t.DetailBg))
	tv.SetTextColor(themes.C(t.DetailFg))
	tv.SetBorderColor(themes.C(t.BorderFocused))
	tv.SetTitleColor(themes.C(t.TextEmphasis))
	return &DetailFullView{TextView: tv}
}

// SetIssue updates the content for the fullscreen view. Must be called outside Draw().
func (dfv *DetailFullView) SetIssue(issue *jira.Issue, termW int) {
	if issue == nil {
		dfv.TextView.SetText("")
		return
	}
	// Full width minus border (2) and small margin (2).
	innerW := termW - 4
	if innerW < 4 {
		innerW = 4
	}
	text := renderIssue(*issue, innerW)
	dfv.TextView.SetText(text)
	dfv.TextView.ScrollToBeginning()
}
