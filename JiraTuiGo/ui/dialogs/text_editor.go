package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const textEditorPageName = "dialog-text-editor"

// ShowTextEditorDialog shows a multi-line text editor dialog using tview.TextArea.
// onDone is called with the text content and saved=true on confirm, saved=false on cancel.
func ShowTextEditorDialog(
	app *tview.Application,
	pages *tview.Pages,
	title string,
	initial string,
	buttonLabel string,
	onDone func(text string, saved bool),
) {
	// Get terminal size for sizing the dialog.
	_, _, termW, termH := pages.GetRect()
	if termW < 60 {
		termW = 80
	}
	if termH < 20 {
		termH = 24
	}

	width := max(50, min(120, termW-8))
	height := max(10, min(40, termH-8))

	done := func(text string, saved bool) {
		pages.RemovePage(textEditorPageName)
		if onDone != nil {
			onDone(text, saved)
		}
	}

	// Try to use TextArea (tview v0.29+); it provides multi-line editing.
	textArea := tview.NewTextArea()
	textArea.SetText(initial, true)
	textArea.SetBorder(false)

	saveBtn := tview.NewButton(buttonLabel)
	cancelBtn := tview.NewButton("Cancel")

	btnFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(saveBtn, len(buttonLabel)+4, 0, false).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(cancelBtn, 10, 0, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textArea, 0, 1, true).
		AddItem(btnFlex, 1, 0, false)

	frame := tview.NewFrame(flex).
		SetBorders(1, 1, 1, 1, 2, 2)
	frame.SetBorder(true).SetTitle(" " + title + " ")

	// Focus order: textArea → saveBtn → cancelBtn → textArea
	focusOrder := []tview.Primitive{textArea, saveBtn, cancelBtn}
	focusIdx := 0

	cycleFocus := func(forward bool) {
		if forward {
			focusIdx = (focusIdx + 1) % len(focusOrder)
		} else {
			focusIdx = (focusIdx - 1 + len(focusOrder)) % len(focusOrder)
		}
		app.SetFocus(focusOrder[focusIdx])
	}

	saveBtn.SetSelectedFunc(func() {
		done(textArea.GetText(), true)
	})
	cancelBtn.SetSelectedFunc(func() {
		done("", false)
	})

	// Tab cycles focus; Escape cancels.
	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			cycleFocus(true)
			return nil
		case tcell.KeyBacktab:
			cycleFocus(false)
			return nil
		case tcell.KeyEscape:
			done("", false)
			return nil
		}
		return event
	})

	modal := centeredBox(frame, width, height)
	pages.AddPage(textEditorPageName, modal, true, true)
	app.SetFocus(textArea)
}
