package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const whatsNewPageName = "dialog-whatsnew"

// ShowWhatsNewDialog shows a scrollable "What's New" dialog with the embedded
// changelog. onClose is called when the dialog is dismissed.
func ShowWhatsNewDialog(
	app *tview.Application,
	pages *tview.Pages,
	changelog string,
	onClose func(),
) {
	_, _, termW, termH := pages.GetRect()
	if termW < 40 {
		termW = 80
	}
	if termH < 20 {
		termH = 40
	}
	width := min(80, termW-4)
	height := min(40, termH-4)

	tv := tview.NewTextView()
	tv.SetDynamicColors(false)
	tv.SetScrollable(true)
	tv.SetWrap(true)
	tv.SetWordWrap(true)
	tv.SetBorder(true)
	tv.SetTitle(" What's New — F1 to close ")
	tv.SetText(changelog)
	tv.ScrollToBeginning()

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyEnter, tcell.KeyF1:
			pages.RemovePage(whatsNewPageName)
			if onClose != nil {
				onClose()
			}
			return nil
		}
		return event
	})

	modal := centeredBox(tv, width, height)
	pages.AddPage(whatsNewPageName, modal, true, true)
	app.SetFocus(tv)
}
