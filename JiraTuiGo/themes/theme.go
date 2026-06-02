package themes

import "github.com/gdamore/tcell/v2"

type ThemeColor struct {
	TrueColor tcell.Color
	Color256  tcell.Color
	Color16   tcell.Color
}

type Theme struct {
	Name string

	// Chrome
	Background     ThemeColor
	Border         ThemeColor
	BorderFocused  ThemeColor

	// Text
	TextNormal   ThemeColor
	TextMuted    ThemeColor
	TextEmphasis ThemeColor

	// Issue list
	ListBg         ThemeColor
	ListFg         ThemeColor
	ListSelectedBg ThemeColor
	ListSelectedFg ThemeColor
	ListHeaderBg   ThemeColor
	ListHeaderFg   ThemeColor

	// Navigation
	NavBg         ThemeColor
	NavFg         ThemeColor
	NavSelectedBg ThemeColor
	NavSelectedFg ThemeColor
	NavSectionFg  ThemeColor

	// Detail panel
	DetailBg     ThemeColor
	DetailFg     ThemeColor
	DetailLabelFg ThemeColor
	DetailValueFg ThemeColor

	// JQL bar
	JqlBg     ThemeColor
	JqlFg     ThemeColor
	JqlHintFg ThemeColor

	// Status bar
	StatusBg       ThemeColor
	StatusFg       ThemeColor
	StatusKeyFg    ThemeColor
	StatusUpdateFg ThemeColor

	// Dialogs
	DialogBg      ThemeColor
	DialogFg      ThemeColor
	DialogBorderFg ThemeColor
	ButtonBg      ThemeColor
	ButtonFg      ThemeColor
	ButtonFocusBg ThemeColor
	ButtonFocusFg ThemeColor
	InputBg       ThemeColor
	InputFg       ThemeColor
	InputFocusBg  ThemeColor

	// Issue type glyph colors
	TypeBug     ThemeColor
	TypeTask    ThemeColor
	TypeStory   ThemeColor
	TypeEpic    ThemeColor
	TypeSubtask ThemeColor
	TypeOther   ThemeColor

	// Priority glyph colors
	PriHighest ThemeColor
	PriHigh    ThemeColor
	PriMedium  ThemeColor
	PriLow     ThemeColor
	PriLowest  ThemeColor

	// Status glyph colors
	StatusTodo       ThemeColor
	StatusInProgress ThemeColor
	StatusInReview   ThemeColor
	StatusBlocked    ThemeColor
	StatusDone       ThemeColor
	StatusCancelled  ThemeColor
}

func rgb(r, g, b int32) ThemeColor {
	tc := tcell.NewRGBColor(r, g, b)
	return ThemeColor{TrueColor: tc, Color256: tc, Color16: tc}
}

func tc256(r, g, b int32, c256 int32, c16 tcell.Color) ThemeColor {
	return ThemeColor{
		TrueColor: tcell.NewRGBColor(r, g, b),
		Color256:  tcell.PaletteColor(int(c256)),
		Color16:   c16,
	}
}

func ansi(c tcell.Color) ThemeColor {
	return ThemeColor{TrueColor: c, Color256: c, Color16: c}
}
