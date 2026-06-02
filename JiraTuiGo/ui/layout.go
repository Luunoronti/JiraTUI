package ui

import "fmt"

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

type Hint struct {
	Key    string
	Action string
}

func StatusBarHints(width int, hints []Hint) string {
	if width >= 80 {
		var parts []string
		for _, h := range hints {
			parts = append(parts, fmt.Sprintf("%s:%s", h.Key, h.Action))
		}
		result := ""
		for i, p := range parts {
			if i > 0 {
				result += "  "
			}
			result += p
		}
		return result
	}
	if width >= 60 {
		var parts []string
		for _, h := range hints {
			parts = append(parts, h.Key)
		}
		result := ""
		for i, p := range parts {
			if i > 0 {
				result += "  "
			}
			result += p
		}
		return result
	}
	return "Ctrl-Q:Quit"
}

func TooSmallMsg(width, height int) string {
	return fmt.Sprintf("Terminal too small (%dx%d). Please resize.", width, height)
}
