package themes

import "github.com/gdamore/tcell/v2"

func Dark() *Theme {
	return &Theme{
		Name: "Dark",

		Background:    tc256(0x1e, 0x1e, 0x2e, 235, tcell.ColorBlack),
		Border:        tc256(0x45, 0x47, 0x5a, 238, tcell.ColorBlue),
		BorderFocused: tc256(0x89, 0xb4, 0xfa, 111, tcell.ColorAqua),

		TextNormal:   tc256(0xcd, 0xd6, 0xf4, 252, tcell.ColorWhite),
		TextMuted:    tc256(0x6c, 0x70, 0x86, 243, tcell.ColorDarkGray),
		TextEmphasis: tc256(0x89, 0xb4, 0xfa, 111, tcell.ColorAqua),

		ListBg:         tc256(0x18, 0x18, 0x26, 234, tcell.ColorBlack),
		ListFg:         tc256(0xcd, 0xd6, 0xf4, 252, tcell.ColorWhite),
		ListSelectedBg: tc256(0x31, 0x33, 0x44, 237, tcell.ColorNavy),
		ListSelectedFg: tc256(0xcb, 0xa6, 0xf7, 183, tcell.ColorWhite),
		ListHeaderBg:   tc256(0x18, 0x18, 0x26, 234, tcell.ColorBlack),
		ListHeaderFg:   tc256(0x89, 0xb4, 0xfa, 111, tcell.ColorAqua),

		NavBg:         tc256(0x24, 0x27, 0x3a, 236, tcell.ColorBlack),
		NavFg:         tc256(0xcd, 0xd6, 0xf4, 252, tcell.ColorWhite),
		NavSelectedBg: tc256(0x31, 0x33, 0x44, 237, tcell.ColorNavy),
		NavSelectedFg: tc256(0xcb, 0xa6, 0xf7, 183, tcell.ColorWhite),
		NavSectionFg:  tc256(0x89, 0xb4, 0xfa, 111, tcell.ColorAqua),

		DetailBg:      tc256(0x1e, 0x1e, 0x2e, 235, tcell.ColorBlack),
		DetailFg:      tc256(0xcd, 0xd6, 0xf4, 252, tcell.ColorWhite),
		DetailLabelFg: tc256(0x89, 0xb4, 0xfa, 111, tcell.ColorAqua),
		DetailValueFg: tc256(0xcd, 0xd6, 0xf4, 252, tcell.ColorWhite),

		JqlBg:     tc256(0x18, 0x18, 0x26, 234, tcell.ColorBlack),
		JqlFg:     tc256(0xa6, 0xe3, 0xa1, 150, tcell.ColorGreen),
		JqlHintFg: tc256(0x6c, 0x70, 0x86, 243, tcell.ColorDarkGray),

		StatusBg:       tc256(0x18, 0x18, 0x26, 234, tcell.ColorBlack),
		StatusFg:       tc256(0x6c, 0x70, 0x86, 243, tcell.ColorDarkGray),
		StatusKeyFg:    tc256(0x89, 0xb4, 0xfa, 111, tcell.ColorAqua),
		StatusUpdateFg: tc256(0xf9, 0xe2, 0xaf, 221, tcell.ColorYellow),

		DialogBg:       tc256(0x24, 0x27, 0x3a, 236, tcell.ColorBlack),
		DialogFg:       tc256(0xcd, 0xd6, 0xf4, 252, tcell.ColorWhite),
		DialogBorderFg: tc256(0x89, 0xb4, 0xfa, 111, tcell.ColorAqua),
		ButtonBg:       tc256(0x45, 0x47, 0x5a, 238, tcell.ColorDarkGray),
		ButtonFg:       tc256(0xcd, 0xd6, 0xf4, 252, tcell.ColorWhite),
		ButtonFocusBg:  tc256(0x89, 0xb4, 0xfa, 111, tcell.ColorBlue),
		ButtonFocusFg:  tc256(0x18, 0x18, 0x26, 234, tcell.ColorBlack),
		InputBg:        tc256(0x18, 0x18, 0x26, 234, tcell.ColorBlack),
		InputFg:        tc256(0xcd, 0xd6, 0xf4, 252, tcell.ColorWhite),
		InputFocusBg:   tc256(0x31, 0x33, 0x44, 237, tcell.ColorNavy),

		TypeBug:     tc256(0xf3, 0x8b, 0xa8, 211, tcell.ColorRed),
		TypeTask:    tc256(0x89, 0xdc, 0xeb, 117, tcell.ColorAqua),
		TypeStory:   tc256(0xa6, 0xe3, 0xa1, 150, tcell.ColorGreen),
		TypeEpic:    tc256(0xcb, 0xa6, 0xf7, 183, tcell.ColorFuchsia),
		TypeSubtask: tc256(0x7f, 0x84, 0x9c, 244, tcell.ColorDarkGray),
		TypeOther:   tc256(0x7f, 0x84, 0x9c, 244, tcell.ColorDarkGray),

		PriHighest: tc256(0xf3, 0x8b, 0xa8, 211, tcell.ColorRed),
		PriHigh:    tc256(0xfa, 0xb3, 0x87, 215, tcell.ColorRed),
		PriMedium:  tc256(0xcd, 0xd6, 0xf4, 252, tcell.ColorWhite),
		PriLow:     tc256(0x89, 0xdc, 0xeb, 117, tcell.ColorAqua),
		PriLowest:  tc256(0x7f, 0x84, 0x9c, 244, tcell.ColorDarkGray),

		StatusTodo:       tc256(0x6c, 0x70, 0x86, 243, tcell.ColorDarkGray),
		StatusInProgress: tc256(0x89, 0xb4, 0xfa, 111, tcell.ColorBlue),
		StatusInReview:   tc256(0xcb, 0xa6, 0xf7, 183, tcell.ColorFuchsia),
		StatusBlocked:    tc256(0xf3, 0x8b, 0xa8, 211, tcell.ColorRed),
		StatusDone:       tc256(0xa6, 0xe3, 0xa1, 150, tcell.ColorGreen),
		StatusCancelled:  tc256(0x7f, 0x84, 0x9c, 244, tcell.ColorDarkGray),
	}
}
