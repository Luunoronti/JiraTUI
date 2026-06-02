package themes

import (
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ColorTier int

const (
	TierTrueColor ColorTier = iota
	Tier256
	Tier16
)

var (
	currentTier  ColorTier
	currentTheme *Theme
	allThemes    []*Theme
)

func init() {
	allThemes = []*Theme{
		Dark(),
		Light(),
		TurboPascal(),
		GreenPhosphor(),
		AmberPhosphor(),
		SolarizedDark(),
		SolarizedLight(),
		HighContrast(),
	}
}

func Detect() {
	colorterm := strings.ToLower(os.Getenv("COLORTERM"))
	if colorterm == "truecolor" || colorterm == "24bit" {
		currentTier = TierTrueColor
		return
	}
	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "truecolor") {
		currentTier = TierTrueColor
		return
	}
	if strings.Contains(term, "256color") {
		currentTier = Tier256
		return
	}
	currentTier = TierTrueColor // Windows Terminal default; overridden by DetectFromScreen
}

// DetectFromScreen refines the tier using tcell's own capability detection,
// which checks COLORTERM, terminfo Tc/RGB caps, and TERM_PROGRAM.
// Returns true if the tier changed (caller should re-apply the current theme).
func DetectFromScreen(colors int) bool {
	var t ColorTier
	switch {
	case colors >= 1<<24:
		t = TierTrueColor
	case colors >= 256:
		t = Tier256
	default:
		t = Tier16
	}
	if t == currentTier {
		return false
	}
	currentTier = t
	return true
}

func pickColor(tc ThemeColor) tcell.Color {
	switch currentTier {
	case TierTrueColor:
		return tc.TrueColor
	case Tier256:
		return tc.Color256
	default:
		return tc.Color16
	}
}

func Current() *Theme {
	if currentTheme == nil {
		return allThemes[0]
	}
	return currentTheme
}

func Apply(t *Theme) {
	currentTheme = t
	tview.Styles.PrimitiveBackgroundColor = pickColor(t.Background)
	tview.Styles.ContrastBackgroundColor = pickColor(t.ListSelectedBg)
	tview.Styles.MoreContrastBackgroundColor = pickColor(t.DialogBg)
	tview.Styles.BorderColor = pickColor(t.Border)
	tview.Styles.TitleColor = pickColor(t.TextEmphasis)
	tview.Styles.GraphicsColor = pickColor(t.Border)
	tview.Styles.PrimaryTextColor = pickColor(t.TextNormal)
	tview.Styles.SecondaryTextColor = pickColor(t.TextMuted)
	tview.Styles.TertiaryTextColor = pickColor(t.TextEmphasis)
	tview.Styles.InverseTextColor = pickColor(t.ListSelectedFg)
	tview.Styles.ContrastSecondaryTextColor = pickColor(t.StatusKeyFg)
}

func Switch(name string) bool {
	for _, t := range allThemes {
		if strings.EqualFold(t.Name, name) {
			Apply(t)
			return true
		}
	}
	return false
}

func AvailableThemes() []string {
	names := make([]string, len(allThemes))
	for i, t := range allThemes {
		names[i] = t.Name
	}
	return names
}

func C(tc ThemeColor) tcell.Color {
	return pickColor(tc)
}
