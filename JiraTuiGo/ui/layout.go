package ui

import (
	"fmt"
	"strings"
)

type Tier int

const (
	Compact Tier = iota
	Normal
	Wide
)

const (
	MinWidth  = 60
	MinHeight = 12
)

func SizeTier(width int) Tier {
	if width < 100 {
		return Compact
	}
	if width < 160 {
		return Normal
	}
	return Wide
}

func TooSmall(width, height int) bool {
	return width < MinWidth || height < MinHeight
}

func NavMaxWidth(tier Tier) int {
	switch tier {
	case Compact:
		return 0 // calculated as % in nav_view
	case Wide:
		return 48
	default:
		return 40
	}
}

func DetailWidth(tier Tier, termWidth int) int {
	switch tier {
	case Normal:
		return termWidth * 40 / 100
	case Wide:
		return termWidth * 38 / 100
	default:
		return 0
	}
}

// Hint holds a status bar entry in one of two formats:
//   "(X)rest"   — Ctrl shortcut; X is the key letter, rest is the label
//   "KEY:Word"  — e.g. "F2:Settings"; KEY is highlighted, Word is the label
type Hint struct {
	Action string
}

// StatusBarHints builds the formatted hints string for the given terminal
// width. keyOpen/keyClose are tview dynamic-colour tags that bracket the
// shortcut key character, e.g. "[aqua]" and "[-]".
func StatusBarHints(width int, hints []Hint, keyOpen, keyClose string) string {
	format := func(h Hint, full bool) string {
		a := h.Action
		// "(X)rest" format
		if len(a) >= 3 && a[0] == '(' {
			end := strings.Index(a[1:], ")") + 1
			if end > 0 {
				key := a[1:end]
				rest := a[end+1:]
				if full {
					return keyOpen + key + keyClose + rest
				}
				return keyOpen + key + keyClose
			}
		}
		// "KEY:Word" format
		if colon := strings.Index(a, ":"); colon > 0 {
			k, r := a[:colon], a[colon+1:]
			if full {
				return keyOpen + k + keyClose + ":" + r
			}
			return keyOpen + k + keyClose
		}
		return a
	}

	sep := "  "
	if width >= 80 {
		var parts []string
		for _, h := range hints {
			parts = append(parts, format(h, true))
		}
		return strings.Join(parts, sep)
	}
	if width >= 60 {
		var parts []string
		for _, h := range hints {
			parts = append(parts, format(h, false))
		}
		return strings.Join(parts, sep)
	}
	return keyOpen + "Q" + keyClose + "uit"
}

func TooSmallMsg(width, height int) string {
	return fmt.Sprintf("Terminal too small (%dx%d). Please resize.", width, height)
}
