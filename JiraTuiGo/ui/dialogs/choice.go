package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const choicePageName = "dialog-choice"

// ShowChoiceDialog shows a centered modal with a tview.List for selecting from options.
// onDone is called with the selected index, or -1 if cancelled.
func ShowChoiceDialog(
	app *tview.Application,
	pages *tview.Pages,
	title string,
	options []string,
	initialIdx int,
	onDone func(idx int),
) {
	height := len(options) + 6
	if height > 20 {
		height = 20
	}
	const width = 60

	list := tview.NewList()
	list.SetBorder(true).SetTitle(" " + title + " ")
	list.ShowSecondaryText(false)

	for _, opt := range options {
		list.AddItem(opt, "", 0, nil)
	}
	if initialIdx >= 0 && initialIdx < len(options) {
		list.SetCurrentItem(initialIdx)
	}

	done := func(idx int) {
		pages.RemovePage(choicePageName)
		if onDone != nil {
			onDone(idx)
		}
	}

	list.SetSelectedFunc(func(idx int, _ string, _ string, _ rune) {
		done(idx)
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			done(-1)
			return nil
		}
		return event
	})

	modal := centeredBox(list, width, height)
	pages.AddPage(choicePageName, modal, true, true)
	app.SetFocus(list)
}
