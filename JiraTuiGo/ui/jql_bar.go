package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/config"
	"jiratui/themes"
)

const jqlBarHeight = 3 // top border + input row + bottom border

// JqlBar is a floating single-line input overlay anchored at the bottom of the
// terminal, just above the status bar. It manages its own rect in Draw() so it
// reflows on every resize without cached coordinates.
//
// Focus is delegated to the embedded InputField so tview's text-editing
// machinery works normally; special keys (Up/Down for history, Escape, Ctrl-J)
// are intercepted via SetInputCapture before they reach the field.
type JqlBar struct {
	*tview.Box

	input   *tview.InputField
	history *config.JqlHistory

	// history navigation state
	histIdx   int    // -1 = not browsing; 0 = newest entry
	savedText string // text saved before browsing started

	OnSubmit func(jql string) // called with the query text when user presses Enter
	OnClose  func()           // called when the bar should be hidden
}

func NewJqlBar(history *config.JqlHistory) *JqlBar {
	jb := &JqlBar{
		Box:     tview.NewBox(),
		input:   tview.NewInputField(),
		history: history,
		histIdx: -1,
	}

	jb.SetBorder(true)

	t := themes.Current()
	jb.Box.SetBackgroundColor(themes.C(t.JqlBg))
	jb.Box.SetBorderColor(themes.C(t.Border))

	jb.input.SetLabel(" JQL: ")
	jb.input.SetLabelColor(themes.C(t.StatusKeyFg))
	jb.input.SetFieldBackgroundColor(themes.C(t.JqlBg))
	jb.input.SetFieldTextColor(themes.C(t.JqlFg))

	// Intercept special keys before the InputField's default handler sees them.
	jb.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			jb.historyOlder()
			return nil
		case tcell.KeyDown:
			jb.historyNewer()
			return nil
		case tcell.KeyEscape:
			if jb.input.GetText() != "" {
				jb.input.SetText("")
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
		// Any regular typing resets history browsing.
		if event.Rune() != 0 {
			jb.histIdx = -1
		}
		return event
	})

	jb.input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			jb.submit()
		}
	})

	return jb
}

// ─── public API ──────────────────────────────────────────────────────────────

// SetText replaces the current text and resets history browsing.
func (jb *JqlBar) SetText(text string) {
	jb.histIdx = -1
	jb.input.SetText(text)
}

// GetText returns the current input text.
func (jb *JqlBar) GetText() string {
	return jb.input.GetText()
}

// FocusAndSelectAll gives focus to the input field and selects all text
// (so the user can start typing to replace without clearing manually).
func (jb *JqlBar) FocusAndSelectAll(app *tview.Application) {
	jb.histIdx = -1
	app.SetFocus(jb)
}

// ─── Focus delegation ─────────────────────────────────────────────────────────

// Focus delegates to the embedded InputField so tview sends key events there.
func (jb *JqlBar) Focus(delegate func(p tview.Primitive)) {
	delegate(jb.input)
}

// HasFocus reports whether the embedded input field is focused.
func (jb *JqlBar) HasFocus() bool {
	return jb.input.HasFocus()
}

// InputHandler delegates key events to the embedded InputField.
// Without this, tview.Pages routes events to JqlBar but tview.Box's default
// handler does nothing — the InputField never receives keystrokes.
func (jb *JqlBar) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return jb.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if h := jb.input.InputHandler(); h != nil {
			h(event, setFocus)
		}
	})
}

// ─── Draw ────────────────────────────────────────────────────────────────────

func (jb *JqlBar) Draw(screen tcell.Screen) {
	t := themes.Current()
	termW, termH := screen.Size()

	// Anchor above the status bar.
	y := termH - 1 - jqlBarHeight
	if y < 1 {
		y = 1
	}
	jb.Box.SetRect(0, y, termW, jqlBarHeight)
	jb.Box.DrawForSubclass(screen, jb)

	innerX, innerY, innerW, _ := jb.GetInnerRect()
	if innerW <= 0 {
		return
	}

	// Right-aligned hint — only shown when there is enough room.
	const hint = "  Enter:run  ↑↓:history  Ctrl-J:close  "
	hintRunes := []rune(hint)
	inputW := innerW
	if innerW >= 60 {
		hintW := len(hintRunes)
		if hintW < innerW-10 {
			inputW = innerW - hintW

			hintStyle := tcell.StyleDefault.
				Foreground(themes.C(t.JqlHintFg)).
				Background(themes.C(t.JqlBg))
			hx := innerX + inputW
			for j, r := range hintRunes {
				if hx+j >= innerX+innerW {
					break
				}
				screen.SetContent(hx+j, innerY, r, nil, hintStyle)
			}
		}
	}

	// Position and draw the input field.
	jb.input.SetRect(innerX, innerY, inputW, 1)
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
		jb.input.SetText(e.EffectiveJql)
	}
}

func (jb *JqlBar) historyNewer() {
	if jb.histIdx <= 0 {
		// Past the newest entry — restore saved text and stop browsing.
		jb.histIdx = -1
		jb.input.SetText(jb.savedText)
		return
	}
	jb.histIdx--
	if e, ok := jb.history.GetByRecentIndex(jb.histIdx); ok {
		jb.input.SetText(e.EffectiveJql)
	}
}

// ─── submit ───────────────────────────────────────────────────────────────────

func (jb *JqlBar) submit() {
	jql := strings.TrimSpace(jb.input.GetText())
	if jql == "" {
		return
	}
	jb.history.Add(jql, jql, false)
	_ = jb.history.Save() // best-effort; ignore errors
	jb.histIdx = -1
	if jb.OnSubmit != nil {
		jb.OnSubmit(jql)
	}
}
