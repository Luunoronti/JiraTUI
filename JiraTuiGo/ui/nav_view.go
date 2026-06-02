package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/jira"
	"jiratui/themes"
)

type navKind int

const (
	navHeader  navKind = iota
	navLeaf
	navProject
)

type navEntry struct {
	kind       navKind
	label      string // display text without extra indent
	jql        string
	indent     int    // spaces prepended when drawing
	projectKey string // for navProject: key used to toggle expansion
}

// NavPanel is a floating overlay anchored top-left, drawn on top of the issue
// list. It manages its own rect inside Draw so it reflows on every resize
// without any cached coordinates.
type NavPanel struct {
	*tview.Box

	entries          []navEntry
	selected         int
	scrollOffset     int
	expandedProjects map[string]bool
	projects         []jira.Project
	filters          []jira.SavedFilter

	OnSelect func(jql string) // called when user selects a leaf; should close panel
	OnClose  func()           // called on Escape / Ctrl-B
}

func NewNavPanel() *NavPanel {
	np := &NavPanel{
		Box:              tview.NewBox(),
		expandedProjects: make(map[string]bool),
		selected:         -1,
	}
	np.SetBorder(true)
	np.SetTitle(" Navigation ")
	np.refreshBoxColors()
	return np
}

func (np *NavPanel) refreshBoxColors() {
	t := themes.Current()
	np.Box.SetBackgroundColor(themes.C(t.NavBg))
	np.Box.SetBorderColor(themes.C(t.BorderFocused))
	np.Box.SetTitleColor(themes.C(t.TextEmphasis))
}

func (np *NavPanel) Populate(projects []jira.Project, filters []jira.SavedFilter) {
	np.projects = projects
	np.filters = filters
	np.rebuildEntries()
}

func (np *NavPanel) rebuildEntries() {
	var entries []navEntry

	// Quick Views
	entries = append(entries, navEntry{kind: navHeader, label: "Quick Views"})
	entries = append(entries, navEntry{kind: navLeaf, indent: 2, label: "My Issues",
		jql: "assignee = currentUser() AND statusCategory != Done ORDER BY updated DESC"})
	entries = append(entries, navEntry{kind: navLeaf, indent: 2, label: "Reported by me",
		jql: "reporter = currentUser() AND statusCategory != Done ORDER BY updated DESC"})
	entries = append(entries, navEntry{kind: navLeaf, indent: 2, label: "Recently updated",
		jql: "updated >= -7d ORDER BY updated DESC"})
	entries = append(entries, navEntry{kind: navLeaf, indent: 2, label: "All issues",
		jql: "ORDER BY updated DESC"})

	// Projects
	if len(np.projects) > 0 {
		entries = append(entries, navEntry{kind: navHeader, label: "Projects"})
		for _, p := range np.projects {
			arrow := "▶"
			if np.expandedProjects[p.Key] {
				arrow = "▼"
			}
			entries = append(entries, navEntry{
				kind:       navProject,
				label:      arrow + " " + p.Key,
				jql:        `project = "` + p.Key + `" ORDER BY updated DESC`,
				projectKey: p.Key,
			})
			if np.expandedProjects[p.Key] {
				entries = append(entries, navEntry{kind: navLeaf, indent: 4, label: "Backlog",
					jql: `project = "` + p.Key + `" AND statusCategory = "To Do" ORDER BY updated DESC`})
				entries = append(entries, navEntry{kind: navLeaf, indent: 4, label: "In Progress",
					jql: `project = "` + p.Key + `" AND statusCategory = "In Progress" ORDER BY updated DESC`})
				entries = append(entries, navEntry{kind: navLeaf, indent: 4, label: "Done",
					jql: `project = "` + p.Key + `" AND statusCategory = Done ORDER BY updated DESC`})
			}
		}
	}

	// Saved Filters
	if len(np.filters) > 0 {
		entries = append(entries, navEntry{kind: navHeader, label: "Filters"})
		for _, f := range np.filters {
			entries = append(entries, navEntry{kind: navLeaf, indent: 2, label: f.Name, jql: f.JQL})
		}
	}

	np.entries = entries
	np.scrollOffset = 0

	// Restore or reset selection to first selectable item.
	if np.selected < 0 || np.selected >= len(entries) || entries[np.selected].kind == navHeader {
		np.selected = np.firstSelectableFrom(0)
	}
}

// ─── selection helpers ───────────────────────────────────────────────────────

func (np *NavPanel) firstSelectableFrom(from int) int {
	for i := from; i < len(np.entries); i++ {
		if np.entries[i].kind != navHeader {
			return i
		}
	}
	return -1
}

func (np *NavPanel) prevSelectable() int {
	for i := np.selected - 1; i >= 0; i-- {
		if np.entries[i].kind != navHeader {
			return i
		}
	}
	return np.selected
}

func (np *NavPanel) nextSelectable() int {
	for i := np.selected + 1; i < len(np.entries); i++ {
		if np.entries[i].kind != navHeader {
			return i
		}
	}
	return np.selected
}

// ─── geometry ────────────────────────────────────────────────────────────────

func (np *NavPanel) computePanelWidth(termW int) (panelW int, tooNarrow bool) {
	tier := SizeTier(termW)
	maxW := NavMaxWidth(tier)
	if maxW == 0 {
		// Compact tier: cap at 60 % of terminal width
		maxW = termW * 60 / 100
		if maxW < 20 {
			maxW = 20
		}
	}

	// Minimum width to show content without excessive truncation.
	// " Navigation " title is the starting reference.
	minW := len(" Navigation ") + 2
	for _, e := range np.entries {
		var textLen int
		switch e.kind {
		case navProject:
			textLen = len([]rune(e.label)) // label already includes arrow prefix
		default:
			textLen = e.indent + len([]rune(e.label))
		}
		// +2 for inner padding on each side (the box border is separate)
		if textLen+2 > minW {
			minW = textLen + 2
		}
	}

	panelW = minW
	if panelW > maxW {
		panelW = maxW
	}

	// If less than 10 columns remain for the issue list, expand to full width.
	if termW-panelW < 10 {
		return termW, true
	}
	return panelW, false
}

// ─── Draw ────────────────────────────────────────────────────────────────────

func (np *NavPanel) Draw(screen tcell.Screen) {
	t := themes.Current()
	termW, termH := screen.Size()

	panelW, tooNarrow := np.computePanelWidth(termW)
	// y=0 starts at top; h leaves 1 row for status bar at termH-1.
	panelH := termH - 1

	// Self-position: the panel decides its own rect each frame.
	np.Box.SetRect(0, 0, panelW, panelH)

	// DrawForSubclass fills the background and draws the border + title using
	// the colors set in NewNavPanel / RefreshColors (no Set* calls here → no
	// SetNeedsDisplay → no redraw loop).
	np.Box.DrawForSubclass(screen, np)

	innerX, innerY, innerW, innerH := np.GetInnerRect()
	if innerW <= 0 || innerH <= 0 {
		return
	}

	// Reserve last row for the "too narrow" hint when needed.
	contentRows := innerH
	if tooNarrow {
		contentRows = innerH - 1
	}

	// Scroll to keep selected visible.
	np.clampScroll(contentRows)

	navBg := themes.C(t.NavBg)
	navFg := themes.C(t.NavFg)
	secFg := themes.C(t.NavSectionFg)
	selBg := themes.C(t.NavSelectedBg)
	selFg := themes.C(t.NavSelectedFg)

	for i, e := range np.entries {
		row := i - np.scrollOffset
		if row < 0 {
			continue
		}
		if row >= contentRows {
			break
		}
		y := innerY + row

		// Build display text.
		var text string
		switch e.kind {
		case navProject:
			text = e.label // already contains "▶ KEY" or "▼ KEY"
		default:
			text = strings.Repeat(" ", e.indent) + e.label
		}

		// Choose colors.
		var fg, bg tcell.Color
		switch {
		case e.kind == navHeader:
			fg, bg = secFg, navBg
		case i == np.selected:
			fg, bg = selFg, selBg
		default:
			fg, bg = navFg, navBg
		}

		style := tcell.StyleDefault.Foreground(fg).Background(bg)

		// Fill row.
		for cx := innerX; cx < innerX+innerW; cx++ {
			screen.SetContent(cx, y, ' ', nil, style)
		}

		// Draw text (truncated with … if needed).
		runes := []rune(text)
		for j, r := range runes {
			if j >= innerW {
				break
			}
			ch := r
			if j == innerW-1 && len(runes) > innerW {
				ch = '…'
			}
			screen.SetContent(innerX+j, y, ch, nil, style)
		}
	}

	// "Too narrow" hint at the last row of the inner area.
	if tooNarrow {
		hintY := innerY + innerH - 1
		hint := []rune("  terminal too narrow — Esc to close  ")
		hintStyle := tcell.StyleDefault.
			Foreground(themes.C(t.TextMuted)).
			Background(navBg)
		for j, r := range hint {
			if j >= innerW {
				break
			}
			screen.SetContent(innerX+j, hintY, r, nil, hintStyle)
		}
	}
}

func (np *NavPanel) clampScroll(visibleRows int) {
	if np.selected < 0 {
		return
	}
	if np.selected < np.scrollOffset {
		np.scrollOffset = np.selected
	}
	if np.selected >= np.scrollOffset+visibleRows {
		np.scrollOffset = np.selected - visibleRows + 1
	}
	if np.scrollOffset < 0 {
		np.scrollOffset = 0
	}
}

// ─── InputHandler ─────────────────────────────────────────────────────────────

func (np *NavPanel) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return np.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			np.selected = np.prevSelectable()
		case tcell.KeyDown:
			np.selected = np.nextSelectable()
		case tcell.KeyEnter:
			np.activateSelected()
		case tcell.KeyEscape, tcell.KeyCtrlB:
			if np.OnClose != nil {
				np.OnClose()
			}
		}
	})
}

func (np *NavPanel) activateSelected() {
	if np.selected < 0 || np.selected >= len(np.entries) {
		return
	}
	e := np.entries[np.selected]
	switch e.kind {
	case navProject:
		// Toggle expansion and rebuild. Keep selection on same project.
		key := e.projectKey
		np.expandedProjects[key] = !np.expandedProjects[key]
		np.rebuildEntries()
		for i, ne := range np.entries {
			if ne.kind == navProject && ne.projectKey == key {
				np.selected = i
				break
			}
		}
	case navLeaf:
		if e.jql != "" && np.OnSelect != nil {
			np.OnSelect(e.jql)
		}
	}
}
