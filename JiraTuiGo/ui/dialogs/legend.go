package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const legendPageName = "dialog-legend"

// ShowLegendDialog shows a read-only dialog with glyph meanings.
// onClose is called when the dialog is dismissed.
func ShowLegendDialog(app *tview.Application, pages *tview.Pages, onClose func()) {
	const width = 62
	const height = 52

	content := `[::b]Glyphs — Issue Types[::-]

  ⊘  Bug              ⚒  Improvement
  ✓  Task             ✦  New Feature
  ★  Story            ⌬  Test
  ⬢  Epic             ‼  Incident
  ↳  Sub-task         ✉  Service Request

[::b]Glyphs — Priorities[::-]

  ⇈  Highest / Critical / Blocker
  ▲  High / Major
  ─  Medium
  ▼  Low / Minor
  ⇊  Lowest / Trivial

[::b]Glyphs — Statuses[::-]

  ○  To Do / Open / Backlog / Ready
  ◐  In Progress / Doing / WIP
  ◑  In Review / Testing / QA
  ✕  Blocked / On Hold / Waiting
  ✓  Done / Closed / Resolved / Fixed
  ⊘  Cancelled / Won't Do / Rejected
  ?  Other / unrecognised

[::b]Navigation[::-]

  Ctrl-B   Toggle navigation panel
  Ctrl-D   Toggle detail panel (side or fullscreen)
  Ctrl-J   Toggle JQL bar
  Enter    Open fullscreen detail
  Escape   Close panel / dialog

[::b]Issue actions[::-]  (issue must be selected)

  Ctrl-P   Change priority
  Ctrl-T   Change status / transition
  Ctrl-A   Change assignee
  Ctrl-E   Edit description
  Ctrl-K   Add comment
  Ctrl-O   Open in browser
  Ctrl-F   Save JQL as filter

[::b]Other[::-]

  Ctrl-G   AI JQL generation
  Ctrl-L   This legend
  Ctrl-R   Refresh
  Ctrl-\   Column visibility
  Ctrl-Q   Quit

  ↑ ↓      Navigate history (in JQL bar)

Press [::b]Enter[::-] or [::b]Esc[::-] to close.`

	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetText(content)
	tv.SetBorder(true).SetTitle(" Glyph Legend ")

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyEnter:
			pages.RemovePage(legendPageName)
			if onClose != nil {
				onClose()
			}
			return nil
		}
		return event
	})

	modal := centeredBox(tv, width, height)
	pages.AddPage(legendPageName, modal, true, true)
	app.SetFocus(tv)
}
