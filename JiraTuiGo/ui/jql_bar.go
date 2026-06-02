package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/config"
	"jiratui/themes"
)

const jqlBarHeight = 6 // top border + 4 content rows + bottom border

// JqlBar is a floating multi-line input overlay anchored above the status bar.
// It uses tview.TextArea so long queries wrap and remain fully visible.
// Enter submits; Up/Down navigate history; Escape clears/closes; Ctrl-J closes.
type JqlBar struct {
	*tview.Box

	input   *tview.TextArea
	history *config.JqlHistory

	histIdx   int    // -1 = not browsing; 0 = newest entry
	savedText string // text saved before browsing started

	OnSubmit func(jql string)
	OnClose  func()
}

func NewJqlBar(history *config.JqlHistory) *JqlBar {
	jb := &JqlBar{
		Box:     tview.NewBox(),
		input:   tview.NewTextArea(),
		history: history,
		histIdx: -1,
	}

	jb.SetBorder(true)

	t := themes.Current()
	jb.Box.SetBackgroundColor(themes.C(t.JqlBg))
	jb.Box.SetBorderColor(themes.C(t.Border))

	jb.input.SetWrap(true)
	jb.input.SetWordWrap(true)
	jb.input.SetTextStyle(tcell.StyleDefault.
		Foreground(themes.C(t.JqlFg)).
		Background(themes.C(t.JqlBg)))

	jb.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			jb.submit()
			return nil
		case tcell.KeyUp:
			jb.historyOlder()
			return nil
		case tcell.KeyDown:
			jb.historyNewer()
			return nil
		case tcell.KeyEscape:
			if jb.input.GetText() != "" {
				jb.input.SetText("", false)
			} else if jb.OnClose != nil {
				jb.OnClose()
			}
			return nil
		case tcell.KeyCtrlJ:
			if jb.OnClose != nil {
				jb.OnClose()
			}
			return nil
		}
		return event
	})

	return jb
}

// ─── public API ──────────────────────────────────────────────────────────────

func (jb *JqlBar) SetText(text string) {
	jb.histIdx = -1
	jb.input.SetText(text, true)
}

func (jb *JqlBar) GetText() string {
	return cleanJql(jb.input.GetText())
}

func (jb *JqlBar) FocusAndSelectAll(app *tview.Application) {
	jb.histIdx = -1
	app.SetFocus(jb)
}

// ─── Focus delegation ─────────────────────────────────────────────────────────

func (jb *JqlBar) Focus(delegate func(p tview.Primitive)) {
	delegate(jb.input)
}

func (jb *JqlBar) HasFocus() bool {
	return jb.input.HasFocus()
}

func (jb *JqlBar) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return jb.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if h := jb.input.InputHandler(); h != nil {
			h(event, setFocus)
		}
	})
}

// ─── Draw ────────────────────────────────────────────────────────────────────

func (jb *JqlBar) Draw(screen tcell.Screen) {
	_, termH := screen.Size()
	termW, _ := screen.Size()

	y := termH - 1 - jqlBarHeight
	if y < 0 {
		y = 0
	}
	jb.Box.SetRect(0, y, termW, jqlBarHeight)
	jb.Box.DrawForSubclass(screen, jb)

	innerX, innerY, innerW, innerH := jb.GetInnerRect()
	if innerW <= 0 || innerH <= 0 {
		return
	}

	jb.input.SetRect(innerX, innerY, innerW, innerH)
	jb.input.Draw(screen)
}

// ─── history navigation ───────────────────────────────────────────────────────

func (jb *JqlBar) historyOlder() {
	if jb.histIdx == -1 {
		jb.savedText = jb.input.GetText()
	}
	next := jb.histIdx + 1
	if e, ok := jb.history.GetByRecentIndex(next); ok {
		jb.histIdx = next
		jb.input.SetText(e.EffectiveJql, true)
	}
}

func (jb *JqlBar) historyNewer() {
	if jb.histIdx <= 0 {
		jb.histIdx = -1
		jb.input.SetText(jb.savedText, true)
		return
	}
	jb.histIdx--
	if e, ok := jb.history.GetByRecentIndex(jb.histIdx); ok {
		jb.input.SetText(e.EffectiveJql, true)
	}
}

// ─── submit ───────────────────────────────────────────────────────────────────

func (jb *JqlBar) submit() {
	jql := cleanJql(jb.input.GetText())
	if jql == "" {
		return
	}
	jb.history.Add(jql, jql, false)
	_ = jb.history.Save()
	jb.histIdx = -1
	if jb.OnSubmit != nil {
		jb.OnSubmit(jql)
	}
}

// cleanJql strips newlines (TextArea may insert them on wrap) and trims space.
func cleanJql(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
}
