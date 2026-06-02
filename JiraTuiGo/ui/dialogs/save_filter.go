package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const saveFilterPageName = "dialog-save-filter"

// ShowSaveFilterDialog shows a form for saving the current JQL as a named filter.
// onDone is called with (name, description, saved).
func ShowSaveFilterDialog(
	app *tview.Application,
	pages *tview.Pages,
	jql string,
	onDone func(name, description string, saved bool),
) {
	const width = 60
	const height = 12

	done := func(name, description string, saved bool) {
		pages.RemovePage(saveFilterPageName)
		if onDone != nil {
			onDone(name, description, saved)
		}
	}

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Save Filter ")

	form.AddInputField("Name", "", 40, nil, nil)
	form.AddInputField("Description", "", 40, nil, nil)

	form.AddButton("Save", func() {
		name := form.GetFormItemByLabel("Name").(*tview.InputField).GetText()
		desc := form.GetFormItemByLabel("Description").(*tview.InputField).GetText()
		done(name, desc, true)
	})
	form.AddButton("Cancel", func() {
		done("", "", false)
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			done("", "", false)
			return nil
		}
		return event
	})

	modal := centeredBox(form, width, height)
	pages.AddPage(saveFilterPageName, modal, true, true)
	app.SetFocus(form)
}
