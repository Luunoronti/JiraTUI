package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── menu definitions ─────────────────────────────────────────────────────────

type menuEntry struct {
	label    string
	shortcut string
	sep      bool
	action   func(*model) (model, tea.Cmd)
}

var menuTitles = []string{"File", "View", "Issue", "Help"}

var menuItems = [][]menuEntry{
	// File
	{
		{label: "Refresh", shortcut: "Ctrl-R", action: func(m *model) (model, tea.Cmd) {
			m.status = "Refreshed"
			return *m, nil
		}},
		{sep: true},
		{label: "Quit", shortcut: "Ctrl-Q", action: func(m *model) (model, tea.Cmd) {
			return *m, tea.Quit
		}},
	},
	// View
	{
		{label: "Toggle Detail", shortcut: "Ctrl-D", action: func(m *model) (model, tea.Cmd) {
			m.showDetail = !m.showDetail
			return *m, nil
		}},
		{sep: true},
		{label: "Legend", shortcut: "Ctrl-L", action: func(m *model) (model, tea.Cmd) {
			m.status = "Legend: ⊘Bug ✓Task ★Story ⬢Epic  |  ⇈Highest ▲High ─Medium ▼Low ⇊Lowest"
			return *m, nil
		}},
	},
	// Issue
	{
		{label: "Open in Browser", shortcut: "Ctrl-O", action: func(m *model) (model, tea.Cmd) {
			if m.selected < len(m.issues) {
				m.status = "Opening " + m.issues[m.selected].Key + " in browser..."
			}
			return *m, nil
		}},
		{sep: true},
		{label: "Set Priority...", shortcut: "Ctrl-P", action: func(m *model) (model, tea.Cmd) {
			m.status = "(Priority dialog - not implemented in prototype)"
			return *m, nil
		}},
		{label: "Set Status...", shortcut: "Ctrl-T", action: func(m *model) (model, tea.Cmd) {
			m.status = "(Status dialog - not implemented in prototype)"
			return *m, nil
		}},
		{label: "Set Assignee...", shortcut: "Ctrl-A", action: func(m *model) (model, tea.Cmd) {
			m.status = "(Assignee dialog - not implemented in prototype)"
			return *m, nil
		}},
	},
	// Help
	{
		{label: "About JiraTUI (bt)", shortcut: "", action: func(m *model) (model, tea.Cmd) {
			m.status = "JiraTUI bubbletea prototype — testing Elm architecture"
			return *m, nil
		}},
	},
}

// ─── model ────────────────────────────────────────────────────────────────────

type model struct {
	width, height int

	issues   []Issue
	selected int
	offset   int // scroll offset

	showDetail bool

	// Menu state
	menuOpen   bool
	activeMenu int // 0-3
	menuSel    int // selected item in dropdown

	status string // status bar message
}

func newModel() model {
	return model{
		issues:     dummyIssues,
		activeMenu: 0,
		status:     "F10:Menu  Ctrl-D:Detail  Ctrl-Q:Quit",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// ─── update ───────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Menu open — route to menu navigation
		if m.menuOpen {
			return m.updateMenuOpen(msg)
		}
		// Normal mode
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+q":
		return m, tea.Quit

	case "f10":
		m.menuOpen = true
		m.activeMenu = 0
		m.menuSel = 0
		m.status = "↑↓:select  Enter:open  ←→:menu  Esc:close"

	// Direct menu access via Alt+letter
	case "alt+f":
		m.menuOpen = true
		m.activeMenu = 0
		m.menuSel = 0
	case "alt+v":
		m.menuOpen = true
		m.activeMenu = 1
		m.menuSel = 0
	case "alt+i":
		m.menuOpen = true
		m.activeMenu = 2
		m.menuSel = 0
	case "alt+h":
		m.menuOpen = true
		m.activeMenu = 3
		m.menuSel = 0

	case "up", "k":
		if m.selected > 0 {
			m.selected--
			if m.selected < m.offset {
				m.offset = m.selected
			}
		}
	case "down", "j":
		if m.selected < len(m.issues)-1 {
			m.selected++
			maxVisible := m.listHeight()
			if m.selected >= m.offset+maxVisible {
				m.offset = m.selected - maxVisible + 1
			}
		}

	case "ctrl+d":
		m.showDetail = !m.showDetail

	case "ctrl+r":
		m.status = "Refreshed"
	}
	return m, nil
}

func (m model) updateMenuOpen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.menuOpen = false
		m.status = "F10:Menu  Ctrl-D:Detail  Ctrl-Q:Quit"

	case "left":
		m.activeMenu = (m.activeMenu - 1 + len(menuTitles)) % len(menuTitles)
		m.menuSel = 0

	case "right":
		m.activeMenu = (m.activeMenu + 1) % len(menuTitles)
		m.menuSel = 0

	case "up":
		items := menuItems[m.activeMenu]
		m.menuSel = (m.menuSel - 1 + len(items)) % len(items)
		// skip separators
		for menuItems[m.activeMenu][m.menuSel].sep {
			m.menuSel = (m.menuSel - 1 + len(items)) % len(items)
		}

	case "down":
		items := menuItems[m.activeMenu]
		m.menuSel = (m.menuSel + 1) % len(items)
		// skip separators
		for menuItems[m.activeMenu][m.menuSel].sep {
			m.menuSel = (m.menuSel + 1) % len(items)
		}

	case "enter":
		item := menuItems[m.activeMenu][m.menuSel]
		if item.action != nil {
			m.menuOpen = false
			newM, cmd := item.action(&m)
			return newM, cmd
		}
	}
	return m, nil
}

// listHeight returns how many issue rows are available.
// Dropdown is an overlay so it never reduces list height.
func (m model) listHeight() int {
	h := m.height - 2 // menu bar + status bar
	if h < 1 {
		h = 1
	}
	return h
}

// menuTitleX returns the visual x position where menu[idx] title starts.
func (m model) menuTitleX(idx int) int {
	x := 1 // leading space
	for i := 0; i < idx; i++ {
		x += 2 + len(menuTitles[i]) + 2 // " Title " per entry
	}
	return x
}

func (m model) dropdownHeight() int {
	items := menuItems[m.activeMenu]
	return len(items) + 2 // +2 for border
}

// ─── view ─────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Render full background at normal height (dropdown does NOT shift content).
	fullHeight := m.height - 2 // minus menu bar and status bar
	if fullHeight < 1 {
		fullHeight = 1
	}

	bg := lipgloss.JoinVertical(lipgloss.Left,
		m.renderMenuBar(),
		m.renderContent(fullHeight),
		m.renderStatusBar(),
	)

	if !m.menuOpen {
		return bg
	}

	// Overlay dropdown on top of the background at (x=menuTitleX, y=1).
	dropdown := m.renderDropdown()
	x := m.menuTitleX(m.activeMenu)
	return overlayAt(bg, dropdown, x, 1)
}

// ─── rendering helpers ────────────────────────────────────────────────────────

func (m model) renderMenuBar() string {
	var sb strings.Builder
	sb.WriteString(" ")

	for i, title := range menuTitles {
		var s string
		if m.menuOpen && m.activeMenu == i {
			s = styleMenuActive.Render(" " + title + " ")
		} else {
			s = styleMenuTitle.Render(" " + title + " ")
		}
		sb.WriteString(s)
	}

	bar := sb.String()
	// Pad to full width
	w := lipgloss.Width(bar)
	if w < m.width {
		bar += styleMenuBar.Render(strings.Repeat(" ", m.width-w))
	}
	return styleMenuBar.Render(bar)
}

func (m model) renderDropdown() string {
	items := menuItems[m.activeMenu]

	var rows []string
	for i, item := range items {
		if item.sep {
			rows = append(rows, styleDropdownSep.Render(strings.Repeat("─", 28)))
			continue
		}
		var row string
		label := item.label
		shortcut := item.shortcut
		// Pad label and shortcut
		padded := fmt.Sprintf("%-18s %s", label, shortcut)
		if i == m.menuSel {
			row = styleDropdownSel.Render(" " + padded + " ")
		} else {
			row = styleDropdownItem.Render(" " + padded + " ")
		}
		rows = append(rows, row)
	}

	content := strings.Join(rows, "\n")
	return styleDropdownBox.Render(content)
}

func (m model) renderContent(h int) string {
	if m.showDetail && m.selected < len(m.issues) {
		// Side-by-side: list on left, detail on right
		listW := m.width * 55 / 100
		detailW := m.width - listW - 1

		list := m.renderList(listW, h)
		detail := m.renderDetail(detailW, h)

		return lipgloss.JoinHorizontal(lipgloss.Top, list, " ", detail)
	}

	return m.renderList(m.width, h)
}

func (m model) renderList(w, h int) string {
	// Column widths
	keyW := 10
	typeW := 2
	priW := 2
	statW := 2
	assigneeW := 16
	summaryW := w - keyW - typeW - priW - statW - assigneeW - 5 // separators
	if summaryW < 10 {
		summaryW = 10
	}

	var rows []string
	for i := m.offset; i < len(m.issues) && len(rows) < h; i++ {
		issue := m.issues[i]
		isSelected := i == m.selected

		key := truncPad(issue.Key, keyW)
		typ := pad(typeGlyph(issue.Type), typeW)
		pri := pad(priGlyph(issue.Priority), priW)
		stat := pad(statusGlyph(issue.Status), statW)
		assignee := truncPad(issue.Assignee, assigneeW)
		summary := truncPad(issue.Summary, summaryW)

		line := " " + key + " " + typ + " " + pri + " " + stat + " " + assignee + " " + summary

		if isSelected {
			rows = append(rows, styleListSelected.Render(
				padRight(line, w),
			))
		} else {
			rows = append(rows, styleListRow.Render(
				padRight(line, w),
			))
		}
	}

	// Fill remaining rows
	for len(rows) < h {
		rows = append(rows, styleListRow.Render(strings.Repeat(" ", w)))
	}

	return strings.Join(rows, "\n")
}

func (m model) renderDetail(w, h int) string {
	if m.selected >= len(m.issues) {
		return styleDetailBox.Width(w).Height(h).Render("")
	}
	issue := m.issues[m.selected]

	var sb strings.Builder
	sb.WriteString(styleDetailKey.Render(issue.Key) + "\n")
	sb.WriteString(strings.Repeat("─", w-4) + "\n")
	sb.WriteString(typeGlyph(issue.Type) + " " + issue.Type + " · " + priGlyph(issue.Priority) + " " + issue.Priority + " · " + statusGlyph(issue.Status) + " " + issue.Status + "\n")
	sb.WriteString("\n")
	sb.WriteString(styleDetailLabel.Render("Assignee: ") + issue.Assignee + "\n")
	sb.WriteString(styleDetailLabel.Render("Updated:  ") + issue.Updated.Format("2006-01-02 15:04") + "\n")
	sb.WriteString("\n")
	sb.WriteString(styleDetailLabel.Render("Summary:") + "\n")

	// Word-wrap summary
	words := strings.Fields(issue.Summary)
	line := ""
	maxW := w - 4
	for _, w := range words {
		if line == "" {
			line = w
		} else if len(line)+1+len(w) <= maxW {
			line += " " + w
		} else {
			sb.WriteString(line + "\n")
			line = w
		}
	}
	if line != "" {
		sb.WriteString(line + "\n")
	}

	return styleDetailBox.Width(w).Height(h).Render(sb.String())
}

func (m model) renderStatusBar() string {
	left := " " + m.status
	right := fmt.Sprintf("  %d/%d issues  | JiraTUI (bt) ", m.selected+1, len(m.issues))

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	pad := m.width - leftW - rightW
	if pad < 0 {
		pad = 0
	}

	return styleStatusBar.Render(left + strings.Repeat(" ", pad) + right)
}

// ─── string helpers ───────────────────────────────────────────────────────────

func truncPad(s string, w int) string {
	runes := []rune(s)
	if len(runes) > w {
		if w > 1 {
			return string(runes[:w-1]) + "…"
		}
		return string(runes[:w])
	}
	return s + strings.Repeat(" ", w-len(runes))
}

func pad(s string, w int) string {
	runes := []rune(s)
	if len(runes) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(runes))
}

func padRight(s string, w int) string {
	vis := lipgloss.Width(s)
	if vis >= w {
		return s
	}
	return s + strings.Repeat(" ", w-vis)
}

