package main

import "github.com/charmbracelet/lipgloss"

var (
	// TurboPascal 5 colour scheme
	// Classic blue IDE with cyan highlights and yellow hotkeys.
	colBg         = lipgloss.Color("#0000AA") // deep blue background
	colSurface    = lipgloss.Color("#005555") // dialog / detail surface
	colOverlay    = lipgloss.Color("#008888") // separator / muted
	colText       = lipgloss.Color("#FFFFFF") // white body text
	colMuted      = lipgloss.Color("#00AAAA") // cyan muted
	colEmphasis   = lipgloss.Color("#FFFF55") // yellow — hotkeys / titles
	colGreen      = lipgloss.Color("#55FF55") // bright green
	colRed        = lipgloss.Color("#FF5555") // bright red
	colMenuBg     = lipgloss.Color("#000080") // slightly darker menu bar
	colSelected   = lipgloss.Color("#00AAAA") // cyan selection bar
	colSelText    = lipgloss.Color("#000000") // black text on selection
	colDropBg     = lipgloss.Color("#008888") // dropdown background (teal)
	colDropBorder = lipgloss.Color("#FFFF55") // yellow border
	colStatusBg   = lipgloss.Color("#000080")
	colStatusFg   = lipgloss.Color("#00AAAA")
)

var (
	styleMenuBar = lipgloss.NewStyle().
			Background(colMenuBg).
			Foreground(colText)

	styleMenuTitle = lipgloss.NewStyle().
			Background(colMenuBg).
			Foreground(colText)

	styleMenuActive = lipgloss.NewStyle().
			Background(colSelected).
			Foreground(colSelText).
			Bold(true)

	styleDropdownBox = lipgloss.NewStyle().
				Background(colDropBg).
				Border(lipgloss.NormalBorder()).
				BorderForeground(colDropBorder).
				Padding(0, 0)

	styleDropdownItem = lipgloss.NewStyle().
				Background(colDropBg).
				Foreground(colText)

	styleDropdownSel = lipgloss.NewStyle().
				Background(colSelected).
				Foreground(colSelText).
				Bold(true)

	styleDropdownSep = lipgloss.NewStyle().
				Background(colDropBg).
				Foreground(colOverlay)

	styleListRow = lipgloss.NewStyle().
			Background(colBg).
			Foreground(colText)

	styleListSelected = lipgloss.NewStyle().
				Background(colSelected).
				Foreground(colSelText).
				Bold(true)

	styleDetailBox = lipgloss.NewStyle().
			Background(colSurface).
			Foreground(colText).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colDropBorder).
			Padding(0, 1)

	styleDetailKey = lipgloss.NewStyle().
			Foreground(colEmphasis).
			Bold(true)

	styleDetailLabel = lipgloss.NewStyle().
				Foreground(colEmphasis)

	styleStatusBar = lipgloss.NewStyle().
			Background(colStatusBg).
			Foreground(colStatusFg)
)
