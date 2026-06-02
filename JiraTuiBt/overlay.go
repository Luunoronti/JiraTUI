package main

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// overlayAt renders `over` on top of `bg` at position (x, y) in the bg string.
// ANSI escape sequences are handled so colors in both bg and overlay are preserved.
// The right portion of bg (after the overlay's right edge) is also preserved.
func overlayAt(bg, over string, x, y int) string {
	bgLines := strings.Split(bg, "\n")
	ovLines := strings.Split(over, "\n")

	for i, ovLine := range ovLines {
		row := y + i
		if row < 0 || row >= len(bgLines) {
			continue
		}
		bgLine := bgLines[row]
		ovW := lipgloss.Width(ovLine)
		bgW := lipgloss.Width(bgLine)

		// Left portion: bg chars 0..x-1
		left := ansiTruncate(bgLine, x)
		actualLeftW := lipgloss.Width(left)
		if actualLeftW < x {
			left += strings.Repeat(" ", x-actualLeftW)
		}

		// Right portion: bg chars (x+ovW)..end
		rightStart := x + ovW
		var right string
		if rightStart < bgW {
			right = ansiSkip(bgLine, rightStart)
		}

		bgLines[row] = left + ovLine + right
	}
	return strings.Join(bgLines, "\n")
}

// ansiTruncate returns the first `width` visual characters of s, preserving ANSI codes.
func ansiTruncate(s string, width int) string {
	var out strings.Builder
	vis := 0
	i := 0
	b := []byte(s)

	for i < len(b) {
		if vis >= width {
			break
		}
		if b[i] == '\x1b' {
			// Escape sequence — copy it without counting visual chars.
			start := i
			i++
			if i < len(b) && b[i] == '[' {
				i++
				for i < len(b) && !isANSIFinal(b[i]) {
					i++
				}
				if i < len(b) {
					i++ // include final byte
				}
			}
			out.Write(b[start:i])
			continue
		}
		r, size := utf8.DecodeRune(b[i:])
		out.WriteRune(r)
		vis += runeWidth(r)
		i += size
	}
	// Emit reset so background colour doesn't bleed into overlay.
	out.WriteString("\x1b[0m")
	return out.String()
}

// ansiSkip returns the part of s starting at visual position `skip`,
// preserving ANSI colour state by re-emitting the last active SGR sequence.
func ansiSkip(s string, skip int) string {
	vis := 0
	i := 0
	b := []byte(s)
	lastSGR := "" // track last colour sequence

	for i < len(b) {
		if b[i] == '\x1b' {
			start := i
			i++
			if i < len(b) && b[i] == '[' {
				i++
				for i < len(b) && !isANSIFinal(b[i]) {
					i++
				}
				if i < len(b) {
					i++
				}
			}
			seq := string(b[start:i])
			if len(seq) > 0 && seq[len(seq)-1] == 'm' {
				lastSGR = seq
			}
			continue
		}
		if vis >= skip {
			// Re-emit last colour context then return remainder.
			return lastSGR + string(b[i:])
		}
		_, size := utf8.DecodeRune(b[i:])
		vis++
		i += size
	}
	return ""
}

func isANSIFinal(c byte) bool {
	return (c >= 0x40 && c <= 0x7e)
}

func runeWidth(r rune) int {
	// Simplified: treat most runes as width 1.
	// Could use go-runewidth for CJK, but fine for our glyphs.
	if r == 0 {
		return 0
	}
	return 1
}
