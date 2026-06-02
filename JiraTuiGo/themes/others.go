package themes

import "github.com/gdamore/tcell/v2"

func Light() *Theme {
	return &Theme{
		Name: "Light",

		Background:    tc256(0xef, 0xf1, 0xf5, 255, tcell.ColorWhite),
		Border:        tc256(0xbc, 0xc0, 0xcc, 250, tcell.ColorDarkGray),
		BorderFocused: tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),

		TextNormal:   tc256(0x4c, 0x4f, 0x69, 240, tcell.ColorBlack),
		TextMuted:    tc256(0x9c, 0xa0, 0xb0, 246, tcell.ColorDarkGray),
		TextEmphasis: tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),

		ListBg:         tc256(0xef, 0xf1, 0xf5, 255, tcell.ColorWhite),
		ListFg:         tc256(0x4c, 0x4f, 0x69, 240, tcell.ColorBlack),
		ListSelectedBg: tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),
		ListSelectedFg: tc256(0xff, 0xff, 0xff, 255, tcell.ColorWhite),
		ListHeaderBg:   tc256(0xe6, 0xe9, 0xef, 253, tcell.ColorWhite),
		ListHeaderFg:   tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),

		NavBg:         tc256(0xe6, 0xe9, 0xef, 253, tcell.ColorWhite),
		NavFg:         tc256(0x4c, 0x4f, 0x69, 240, tcell.ColorBlack),
		NavSelectedBg: tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),
		NavSelectedFg: tc256(0xff, 0xff, 0xff, 255, tcell.ColorWhite),
		NavSectionFg:  tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),

		DetailBg:      tc256(0xef, 0xf1, 0xf5, 255, tcell.ColorWhite),
		DetailFg:      tc256(0x4c, 0x4f, 0x69, 240, tcell.ColorBlack),
		DetailLabelFg: tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),
		DetailValueFg: tc256(0x4c, 0x4f, 0x69, 240, tcell.ColorBlack),

		JqlBg:     tc256(0xe6, 0xe9, 0xef, 253, tcell.ColorWhite),
		JqlFg:     tc256(0x17, 0x99, 0x5b, 29, tcell.ColorGreen),
		JqlHintFg: tc256(0x9c, 0xa0, 0xb0, 246, tcell.ColorDarkGray),

		StatusBg:       tc256(0xe6, 0xe9, 0xef, 253, tcell.ColorWhite),
		StatusFg:       tc256(0x9c, 0xa0, 0xb0, 246, tcell.ColorDarkGray),
		StatusKeyFg:    tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),
		StatusUpdateFg: tc256(0xdf, 0x81, 0x06, 172, tcell.ColorYellow),

		DialogBg:       tc256(0xef, 0xf1, 0xf5, 255, tcell.ColorWhite),
		DialogFg:       tc256(0x4c, 0x4f, 0x69, 240, tcell.ColorBlack),
		DialogBorderFg: tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),
		ButtonBg:       tc256(0xbc, 0xc0, 0xcc, 250, tcell.ColorDarkGray),
		ButtonFg:       tc256(0x4c, 0x4f, 0x69, 240, tcell.ColorBlack),
		ButtonFocusBg:  tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),
		ButtonFocusFg:  tc256(0xff, 0xff, 0xff, 255, tcell.ColorWhite),
		InputBg:        tc256(0xff, 0xff, 0xff, 255, tcell.ColorWhite),
		InputFg:        tc256(0x4c, 0x4f, 0x69, 240, tcell.ColorBlack),
		InputFocusBg:   tc256(0xe8, 0xf0, 0xfe, 254, tcell.ColorWhite),

		TypeBug:     tc256(0xd2, 0x00, 0x00, 160, tcell.ColorRed),
		TypeTask:    tc256(0x00, 0x7a, 0xa2, 31, tcell.ColorAqua),
		TypeStory:   tc256(0x17, 0x99, 0x5b, 29, tcell.ColorGreen),
		TypeEpic:    tc256(0x81, 0x2d, 0xd5, 92, tcell.ColorFuchsia),
		TypeSubtask: tc256(0x9c, 0xa0, 0xb0, 246, tcell.ColorDarkGray),
		TypeOther:   tc256(0x9c, 0xa0, 0xb0, 246, tcell.ColorDarkGray),

		PriHighest: tc256(0xd2, 0x00, 0x00, 160, tcell.ColorRed),
		PriHigh:    tc256(0xe6, 0x4d, 0x00, 166, tcell.ColorRed),
		PriMedium:  tc256(0x4c, 0x4f, 0x69, 240, tcell.ColorBlack),
		PriLow:     tc256(0x00, 0x7a, 0xa2, 31, tcell.ColorAqua),
		PriLowest:  tc256(0x9c, 0xa0, 0xb0, 246, tcell.ColorDarkGray),

		StatusTodo:       tc256(0x9c, 0xa0, 0xb0, 246, tcell.ColorDarkGray),
		StatusInProgress: tc256(0x17, 0x76, 0xd3, 26, tcell.ColorBlue),
		StatusInReview:   tc256(0x81, 0x2d, 0xd5, 92, tcell.ColorFuchsia),
		StatusBlocked:    tc256(0xd2, 0x00, 0x00, 160, tcell.ColorRed),
		StatusDone:       tc256(0x17, 0x99, 0x5b, 29, tcell.ColorGreen),
		StatusCancelled:  tc256(0x9c, 0xa0, 0xb0, 246, tcell.ColorDarkGray),
	}
}

func TurboPascal() *Theme {
	return &Theme{
		Name: "TurboPascal",

		Background:    ansi(tcell.ColorNavy),
		Border:        ansi(tcell.ColorAqua),
		BorderFocused: ansi(tcell.ColorYellow),

		TextNormal:   ansi(tcell.ColorAqua),
		TextMuted:    ansi(tcell.ColorBlue),
		TextEmphasis: ansi(tcell.ColorYellow),

		ListBg:         ansi(tcell.ColorNavy),
		ListFg:         ansi(tcell.ColorAqua),
		ListSelectedBg: ansi(tcell.ColorAqua),
		ListSelectedFg: ansi(tcell.ColorNavy),
		ListHeaderBg:   ansi(tcell.ColorNavy),
		ListHeaderFg:   ansi(tcell.ColorYellow),

		NavBg:         ansi(tcell.ColorBlue),
		NavFg:         ansi(tcell.ColorAqua),
		NavSelectedBg: ansi(tcell.ColorAqua),
		NavSelectedFg: ansi(tcell.ColorNavy),
		NavSectionFg:  ansi(tcell.ColorYellow),

		DetailBg:      ansi(tcell.ColorNavy),
		DetailFg:      ansi(tcell.ColorAqua),
		DetailLabelFg: ansi(tcell.ColorYellow),
		DetailValueFg: ansi(tcell.ColorAqua),

		JqlBg:     ansi(tcell.ColorBlue),
		JqlFg:     ansi(tcell.ColorWhite),
		JqlHintFg: ansi(tcell.ColorAqua),

		StatusBg:       ansi(tcell.ColorBlue),
		StatusFg:       ansi(tcell.ColorAqua),
		StatusKeyFg:    ansi(tcell.ColorYellow),
		StatusUpdateFg: ansi(tcell.ColorYellow),

		DialogBg:       ansi(tcell.ColorBlue),
		DialogFg:       ansi(tcell.ColorAqua),
		DialogBorderFg: ansi(tcell.ColorYellow),
		ButtonBg:       ansi(tcell.ColorNavy),
		ButtonFg:       ansi(tcell.ColorAqua),
		ButtonFocusBg:  ansi(tcell.ColorYellow),
		ButtonFocusFg:  ansi(tcell.ColorNavy),
		InputBg:        ansi(tcell.ColorNavy),
		InputFg:        ansi(tcell.ColorWhite),
		InputFocusBg:   ansi(tcell.ColorBlue),

		TypeBug:     ansi(tcell.ColorRed),
		TypeTask:    ansi(tcell.ColorAqua),
		TypeStory:   ansi(tcell.ColorGreen),
		TypeEpic:    ansi(tcell.ColorFuchsia),
		TypeSubtask: ansi(tcell.ColorSilver),
		TypeOther:   ansi(tcell.ColorSilver),

		PriHighest: ansi(tcell.ColorRed),
		PriHigh:    ansi(tcell.ColorFuchsia),
		PriMedium:  ansi(tcell.ColorAqua),
		PriLow:     ansi(tcell.ColorBlue),
		PriLowest:  ansi(tcell.ColorNavy),

		StatusTodo:       ansi(tcell.ColorSilver),
		StatusInProgress: ansi(tcell.ColorAqua),
		StatusInReview:   ansi(tcell.ColorFuchsia),
		StatusBlocked:    ansi(tcell.ColorRed),
		StatusDone:       ansi(tcell.ColorGreen),
		StatusCancelled:  ansi(tcell.ColorSilver),
	}
}

func GreenPhosphor() *Theme {
	return &Theme{
		Name: "Green Phosphor",

		Background:    ansi(tcell.ColorBlack),
		Border:        ansi(tcell.ColorGreen),
		BorderFocused: ansi(tcell.ColorLime),

		TextNormal:   ansi(tcell.ColorGreen),
		TextMuted:    ansi(tcell.ColorOlive),
		TextEmphasis: ansi(tcell.ColorLime),

		ListBg:         ansi(tcell.ColorBlack),
		ListFg:         ansi(tcell.ColorGreen),
		ListSelectedBg: ansi(tcell.ColorGreen),
		ListSelectedFg: ansi(tcell.ColorBlack),
		ListHeaderBg:   ansi(tcell.ColorBlack),
		ListHeaderFg:   ansi(tcell.ColorLime),

		NavBg:         ansi(tcell.ColorBlack),
		NavFg:         ansi(tcell.ColorGreen),
		NavSelectedBg: ansi(tcell.ColorGreen),
		NavSelectedFg: ansi(tcell.ColorBlack),
		NavSectionFg:  ansi(tcell.ColorLime),

		DetailBg:      ansi(tcell.ColorBlack),
		DetailFg:      ansi(tcell.ColorGreen),
		DetailLabelFg: ansi(tcell.ColorLime),
		DetailValueFg: ansi(tcell.ColorGreen),

		JqlBg:     ansi(tcell.ColorBlack),
		JqlFg:     ansi(tcell.ColorLime),
		JqlHintFg: ansi(tcell.ColorOlive),

		StatusBg:       ansi(tcell.ColorBlack),
		StatusFg:       ansi(tcell.ColorOlive),
		StatusKeyFg:    ansi(tcell.ColorLime),
		StatusUpdateFg: ansi(tcell.ColorLime),

		DialogBg:       ansi(tcell.ColorBlack),
		DialogFg:       ansi(tcell.ColorGreen),
		DialogBorderFg: ansi(tcell.ColorLime),
		ButtonBg:       ansi(tcell.ColorOlive),
		ButtonFg:       ansi(tcell.ColorBlack),
		ButtonFocusBg:  ansi(tcell.ColorLime),
		ButtonFocusFg:  ansi(tcell.ColorBlack),
		InputBg:        ansi(tcell.ColorBlack),
		InputFg:        ansi(tcell.ColorGreen),
		InputFocusBg:   ansi(tcell.ColorOlive),

		TypeBug:     ansi(tcell.ColorLime),
		TypeTask:    ansi(tcell.ColorGreen),
		TypeStory:   ansi(tcell.ColorGreen),
		TypeEpic:    ansi(tcell.ColorLime),
		TypeSubtask: ansi(tcell.ColorOlive),
		TypeOther:   ansi(tcell.ColorOlive),

		PriHighest: ansi(tcell.ColorLime),
		PriHigh:    ansi(tcell.ColorGreen),
		PriMedium:  ansi(tcell.ColorGreen),
		PriLow:     ansi(tcell.ColorOlive),
		PriLowest:  ansi(tcell.ColorOlive),

		StatusTodo:       ansi(tcell.ColorOlive),
		StatusInProgress: ansi(tcell.ColorGreen),
		StatusInReview:   ansi(tcell.ColorLime),
		StatusBlocked:    ansi(tcell.ColorLime),
		StatusDone:       ansi(tcell.ColorGreen),
		StatusCancelled:  ansi(tcell.ColorOlive),
	}
}

func AmberPhosphor() *Theme {
	return &Theme{
		Name: "Amber Phosphor",

		Background:    ansi(tcell.ColorBlack),
		Border:        ansi(tcell.ColorOlive),
		BorderFocused: ansi(tcell.ColorYellow),

		TextNormal:   ansi(tcell.ColorOlive),
		TextMuted:    tc256(0x80, 0x60, 0x00, 130, tcell.ColorOlive),
		TextEmphasis: ansi(tcell.ColorYellow),

		ListBg:         ansi(tcell.ColorBlack),
		ListFg:         ansi(tcell.ColorOlive),
		ListSelectedBg: ansi(tcell.ColorOlive),
		ListSelectedFg: ansi(tcell.ColorBlack),
		ListHeaderBg:   ansi(tcell.ColorBlack),
		ListHeaderFg:   ansi(tcell.ColorYellow),

		NavBg:         ansi(tcell.ColorBlack),
		NavFg:         ansi(tcell.ColorOlive),
		NavSelectedBg: ansi(tcell.ColorOlive),
		NavSelectedFg: ansi(tcell.ColorBlack),
		NavSectionFg:  ansi(tcell.ColorYellow),

		DetailBg:      ansi(tcell.ColorBlack),
		DetailFg:      ansi(tcell.ColorOlive),
		DetailLabelFg: ansi(tcell.ColorYellow),
		DetailValueFg: ansi(tcell.ColorOlive),

		JqlBg:     ansi(tcell.ColorBlack),
		JqlFg:     ansi(tcell.ColorYellow),
		JqlHintFg: ansi(tcell.ColorOlive),

		StatusBg:       ansi(tcell.ColorBlack),
		StatusFg:       ansi(tcell.ColorOlive),
		StatusKeyFg:    ansi(tcell.ColorYellow),
		StatusUpdateFg: ansi(tcell.ColorYellow),

		DialogBg:       ansi(tcell.ColorBlack),
		DialogFg:       ansi(tcell.ColorOlive),
		DialogBorderFg: ansi(tcell.ColorYellow),
		ButtonBg:       ansi(tcell.ColorOlive),
		ButtonFg:       ansi(tcell.ColorBlack),
		ButtonFocusBg:  ansi(tcell.ColorYellow),
		ButtonFocusFg:  ansi(tcell.ColorBlack),
		InputBg:        ansi(tcell.ColorBlack),
		InputFg:        ansi(tcell.ColorOlive),
		InputFocusBg:   tc256(0x40, 0x30, 0x00, 58, tcell.ColorOlive),

		TypeBug:     ansi(tcell.ColorYellow),
		TypeTask:    ansi(tcell.ColorOlive),
		TypeStory:   ansi(tcell.ColorOlive),
		TypeEpic:    ansi(tcell.ColorYellow),
		TypeSubtask: tc256(0x80, 0x60, 0x00, 130, tcell.ColorOlive),
		TypeOther:   tc256(0x80, 0x60, 0x00, 130, tcell.ColorOlive),

		PriHighest: ansi(tcell.ColorYellow),
		PriHigh:    ansi(tcell.ColorOlive),
		PriMedium:  ansi(tcell.ColorOlive),
		PriLow:     tc256(0x80, 0x60, 0x00, 130, tcell.ColorOlive),
		PriLowest:  tc256(0x80, 0x60, 0x00, 130, tcell.ColorOlive),

		StatusTodo:       tc256(0x80, 0x60, 0x00, 130, tcell.ColorOlive),
		StatusInProgress: ansi(tcell.ColorOlive),
		StatusInReview:   ansi(tcell.ColorYellow),
		StatusBlocked:    ansi(tcell.ColorYellow),
		StatusDone:       ansi(tcell.ColorOlive),
		StatusCancelled:  tc256(0x80, 0x60, 0x00, 130, tcell.ColorOlive),
	}
}

func SolarizedDark() *Theme {
	return &Theme{
		Name: "Solarized Dark",

		Background:    tc256(0x00, 0x2b, 0x36, 235, tcell.ColorBlack),
		Border:        tc256(0x07, 0x36, 0x42, 236, tcell.ColorBlue),
		BorderFocused: tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorAqua),

		TextNormal:   tc256(0x83, 0x94, 0x96, 244, tcell.ColorWhite),
		TextMuted:    tc256(0x65, 0x7b, 0x83, 242, tcell.ColorDarkGray),
		TextEmphasis: tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorAqua),

		ListBg:         tc256(0x00, 0x2b, 0x36, 235, tcell.ColorBlack),
		ListFg:         tc256(0x83, 0x94, 0x96, 244, tcell.ColorWhite),
		ListSelectedBg: tc256(0x07, 0x36, 0x42, 236, tcell.ColorNavy),
		ListSelectedFg: tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite),
		ListHeaderBg:   tc256(0x00, 0x2b, 0x36, 235, tcell.ColorBlack),
		ListHeaderFg:   tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorAqua),

		NavBg:         tc256(0x07, 0x36, 0x42, 236, tcell.ColorBlack),
		NavFg:         tc256(0x83, 0x94, 0x96, 244, tcell.ColorWhite),
		NavSelectedBg: tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorBlue),
		NavSelectedFg: tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite),
		NavSectionFg:  tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorAqua),

		DetailBg:      tc256(0x00, 0x2b, 0x36, 235, tcell.ColorBlack),
		DetailFg:      tc256(0x83, 0x94, 0x96, 244, tcell.ColorWhite),
		DetailLabelFg: tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorAqua),
		DetailValueFg: tc256(0x83, 0x94, 0x96, 244, tcell.ColorWhite),

		JqlBg:     tc256(0x07, 0x36, 0x42, 236, tcell.ColorBlack),
		JqlFg:     tc256(0x85, 0x99, 0x00, 64, tcell.ColorGreen),
		JqlHintFg: tc256(0x65, 0x7b, 0x83, 242, tcell.ColorDarkGray),

		StatusBg:       tc256(0x07, 0x36, 0x42, 236, tcell.ColorBlack),
		StatusFg:       tc256(0x65, 0x7b, 0x83, 242, tcell.ColorDarkGray),
		StatusKeyFg:    tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorAqua),
		StatusUpdateFg: tc256(0xb5, 0x89, 0x00, 136, tcell.ColorYellow),

		DialogBg:       tc256(0x07, 0x36, 0x42, 236, tcell.ColorBlack),
		DialogFg:       tc256(0x83, 0x94, 0x96, 244, tcell.ColorWhite),
		DialogBorderFg: tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorAqua),
		ButtonBg:       tc256(0x07, 0x36, 0x42, 236, tcell.ColorDarkGray),
		ButtonFg:       tc256(0x83, 0x94, 0x96, 244, tcell.ColorWhite),
		ButtonFocusBg:  tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorBlue),
		ButtonFocusFg:  tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite),
		InputBg:        tc256(0x00, 0x2b, 0x36, 235, tcell.ColorBlack),
		InputFg:        tc256(0x83, 0x94, 0x96, 244, tcell.ColorWhite),
		InputFocusBg:   tc256(0x07, 0x36, 0x42, 236, tcell.ColorNavy),

		TypeBug:     tc256(0xdc, 0x32, 0x2f, 160, tcell.ColorRed),
		TypeTask:    tc256(0x2a, 0xa1, 0x98, 37, tcell.ColorAqua),
		TypeStory:   tc256(0x85, 0x99, 0x00, 64, tcell.ColorGreen),
		TypeEpic:    tc256(0x6c, 0x71, 0xc4, 62, tcell.ColorFuchsia),
		TypeSubtask: tc256(0x65, 0x7b, 0x83, 242, tcell.ColorDarkGray),
		TypeOther:   tc256(0x65, 0x7b, 0x83, 242, tcell.ColorDarkGray),

		PriHighest: tc256(0xdc, 0x32, 0x2f, 160, tcell.ColorRed),
		PriHigh:    tc256(0xcb, 0x4b, 0x16, 166, tcell.ColorRed),
		PriMedium:  tc256(0x83, 0x94, 0x96, 244, tcell.ColorWhite),
		PriLow:     tc256(0x2a, 0xa1, 0x98, 37, tcell.ColorAqua),
		PriLowest:  tc256(0x65, 0x7b, 0x83, 242, tcell.ColorDarkGray),

		StatusTodo:       tc256(0x65, 0x7b, 0x83, 242, tcell.ColorDarkGray),
		StatusInProgress: tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorBlue),
		StatusInReview:   tc256(0x6c, 0x71, 0xc4, 62, tcell.ColorFuchsia),
		StatusBlocked:    tc256(0xdc, 0x32, 0x2f, 160, tcell.ColorRed),
		StatusDone:       tc256(0x85, 0x99, 0x00, 64, tcell.ColorGreen),
		StatusCancelled:  tc256(0x65, 0x7b, 0x83, 242, tcell.ColorDarkGray),
	}
}

func SolarizedLight() *Theme {
	t := SolarizedDark()
	t.Name = "Solarized Light"
	t.Background = tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite)
	t.Border = tc256(0xee, 0xe8, 0xd5, 254, tcell.ColorDarkGray)
	t.ListBg = tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite)
	t.ListFg = tc256(0x65, 0x7b, 0x83, 242, tcell.ColorBlack)
	t.ListSelectedBg = tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorBlue)
	t.ListSelectedFg = tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite)
	t.ListHeaderBg = tc256(0xee, 0xe8, 0xd5, 254, tcell.ColorWhite)
	t.NavBg = tc256(0xee, 0xe8, 0xd5, 254, tcell.ColorWhite)
	t.NavFg = tc256(0x65, 0x7b, 0x83, 242, tcell.ColorBlack)
	t.NavSelectedBg = tc256(0x26, 0x8b, 0xd2, 32, tcell.ColorBlue)
	t.NavSelectedFg = tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite)
	t.DetailBg = tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite)
	t.DetailFg = tc256(0x65, 0x7b, 0x83, 242, tcell.ColorBlack)
	t.JqlBg = tc256(0xee, 0xe8, 0xd5, 254, tcell.ColorWhite)
	t.JqlFg = tc256(0x85, 0x99, 0x00, 64, tcell.ColorGreen)
	t.StatusBg = tc256(0xee, 0xe8, 0xd5, 254, tcell.ColorWhite)
	t.StatusFg = tc256(0x93, 0xa1, 0xa1, 245, tcell.ColorDarkGray)
	t.DialogBg = tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite)
	t.DialogFg = tc256(0x65, 0x7b, 0x83, 242, tcell.ColorBlack)
	t.ButtonBg = tc256(0xee, 0xe8, 0xd5, 254, tcell.ColorDarkGray)
	t.ButtonFg = tc256(0x65, 0x7b, 0x83, 242, tcell.ColorBlack)
	t.InputBg = tc256(0xfd, 0xf6, 0xe3, 230, tcell.ColorWhite)
	t.InputFg = tc256(0x65, 0x7b, 0x83, 242, tcell.ColorBlack)
	t.InputFocusBg = tc256(0xee, 0xe8, 0xd5, 254, tcell.ColorWhite)
	t.TextNormal = tc256(0x65, 0x7b, 0x83, 242, tcell.ColorBlack)
	t.TextMuted = tc256(0x93, 0xa1, 0xa1, 245, tcell.ColorDarkGray)
	return t
}

func HighContrast() *Theme {
	return &Theme{
		Name: "High Contrast",

		Background:    ansi(tcell.ColorBlack),
		Border:        ansi(tcell.ColorWhite),
		BorderFocused: ansi(tcell.ColorYellow),

		TextNormal:   ansi(tcell.ColorWhite),
		TextMuted:    ansi(tcell.ColorSilver),
		TextEmphasis: ansi(tcell.ColorYellow),

		ListBg:         ansi(tcell.ColorBlack),
		ListFg:         ansi(tcell.ColorWhite),
		ListSelectedBg: ansi(tcell.ColorWhite),
		ListSelectedFg: ansi(tcell.ColorBlack),
		ListHeaderBg:   ansi(tcell.ColorBlack),
		ListHeaderFg:   ansi(tcell.ColorYellow),

		NavBg:         ansi(tcell.ColorBlack),
		NavFg:         ansi(tcell.ColorWhite),
		NavSelectedBg: ansi(tcell.ColorWhite),
		NavSelectedFg: ansi(tcell.ColorBlack),
		NavSectionFg:  ansi(tcell.ColorYellow),

		DetailBg:      ansi(tcell.ColorBlack),
		DetailFg:      ansi(tcell.ColorWhite),
		DetailLabelFg: ansi(tcell.ColorYellow),
		DetailValueFg: ansi(tcell.ColorWhite),

		JqlBg:     ansi(tcell.ColorBlack),
		JqlFg:     ansi(tcell.ColorGreen),
		JqlHintFg: ansi(tcell.ColorSilver),

		StatusBg:       ansi(tcell.ColorBlack),
		StatusFg:       ansi(tcell.ColorSilver),
		StatusKeyFg:    ansi(tcell.ColorYellow),
		StatusUpdateFg: ansi(tcell.ColorYellow),

		DialogBg:       ansi(tcell.ColorBlack),
		DialogFg:       ansi(tcell.ColorWhite),
		DialogBorderFg: ansi(tcell.ColorYellow),
		ButtonBg:       ansi(tcell.ColorBlack),
		ButtonFg:       ansi(tcell.ColorWhite),
		ButtonFocusBg:  ansi(tcell.ColorYellow),
		ButtonFocusFg:  ansi(tcell.ColorBlack),
		InputBg:        ansi(tcell.ColorBlack),
		InputFg:        ansi(tcell.ColorWhite),
		InputFocusBg:   ansi(tcell.ColorBlue),

		TypeBug:     ansi(tcell.ColorRed),
		TypeTask:    ansi(tcell.ColorAqua),
		TypeStory:   ansi(tcell.ColorGreen),
		TypeEpic:    ansi(tcell.ColorFuchsia),
		TypeSubtask: ansi(tcell.ColorSilver),
		TypeOther:   ansi(tcell.ColorSilver),

		PriHighest: ansi(tcell.ColorRed),
		PriHigh:    ansi(tcell.ColorFuchsia),
		PriMedium:  ansi(tcell.ColorWhite),
		PriLow:     ansi(tcell.ColorAqua),
		PriLowest:  ansi(tcell.ColorSilver),

		StatusTodo:       ansi(tcell.ColorSilver),
		StatusInProgress: ansi(tcell.ColorAqua),
		StatusInReview:   ansi(tcell.ColorFuchsia),
		StatusBlocked:    ansi(tcell.ColorRed),
		StatusDone:       ansi(tcell.ColorGreen),
		StatusCancelled:  ansi(tcell.ColorSilver),
	}
}
