package main

import "github.com/charmbracelet/lipgloss"

var (
	// Colors (catppuccin Mocha inspired)
	colBg         = lipgloss.Color("#1e1e2e")
	colSurface    = lipgloss.Color("#313244")
	colOverlay    = lipgloss.Color("#45475a")
	colText       = lipgloss.Color("#cdd6f4")
	colMuted      = lipgloss.Color("#6c7086")
	colEmphasis   = lipgloss.Color("#89b4fa")
	colGreen      = lipgloss.Color("#a6e3a1")
	colRed        = lipgloss.Color("#f38ba8")
	colMenuBg     = lipgloss.Color("#181825")
	colSelected   = lipgloss.Color("#313244")
	colSelText    = lipgloss.Color("#cba6f7")
	colDropBg     = lipgloss.Color("#24273a")
	colDropBorder = lipgloss.Color("#89b4fa")
	colStatusBg   = lipgloss.Color("#181825")
	colStatusFg   = lipgloss.Color("#6c7086")
)

var (
	styleMenuBar = lipgloss.NewStyle().
			Background(colMenuBg).
			Foreground(colText)

	styleMenuTitle = lipgloss.NewStyle().
			Background(colMenuBg).
			Foreground(colText)

	styleMenuActive = lipgloss.NewStyle().
			Background(colEmphasis).
			Foreground(colMenuBg).
			Bold(true)

	styleDropdownBox = lipgloss.NewStyle().
				Background(colDropBg).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colDropBorder).
				Padding(0, 0)

	styleDropdownItem = lipgloss.NewStyle().
				Background(colDropBg).
				Foreground(colText)

	styleDropdownSel = lipgloss.NewStyle().
				Background(colEmphasis).
				Foreground(colMenuBg).
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
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colEmphasis).
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
