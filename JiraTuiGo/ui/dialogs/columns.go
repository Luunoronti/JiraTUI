package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/config"
)

const columnsPageName = "dialog-columns"

// ShowColumnsDialog shows a checkbox dialog for toggling issue list column
// visibility. At least one column must stay checked. onSave is called with
// the new config if the user confirms; onClose is called on cancel.
func ShowColumnsDialog(
	app *tview.Application,
	pages *tview.Pages,
	current config.ColumnVisibilityConfig,
	onSave func(cfg config.ColumnVisibilityConfig),
	onClose func(),
) {
	// Work on a mutable copy.
	cfg := current

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(" Column Visibility ")

	form.AddCheckbox("Key", cfg.Key, func(v bool) { cfg.Key = v })
	form.AddCheckbox("Type glyph", cfg.Type, func(v bool) { cfg.Type = v })
	form.AddCheckbox("Priority glyph", cfg.Priority, func(v bool) { cfg.Priority = v })
	form.AddCheckbox("Status", cfg.Status, func(v bool) { cfg.Status = v })
	form.AddCheckbox("Assignee", cfg.Assignee, func(v bool) { cfg.Assignee = v })
	form.AddCheckbox("Summary", cfg.Summary, func(v bool) { cfg.Summary = v })

	closeDialog := func() {
		pages.RemovePage(columnsPageName)
		if onClose != nil {
			onClose()
		}
	}

	form.AddButton("Apply", func() {
		// Ensure at least one column remains visible.
		if !cfg.Key && !cfg.Type && !cfg.Priority && !cfg.Status && !cfg.Assignee && !cfg.Summary {
			return // silently ignore — all unchecked is meaningless
		}
		pages.RemovePage(columnsPageName)
		if onSave != nil {
			onSave(cfg)
		}
	})

	form.AddButton("Cancel", func() {
		closeDialog()
	})

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeDialog()
			return nil
		}
		return event
	})

	modal := centeredBox(form, 42, 20)
	pages.AddPage(columnsPageName, modal, true, true)
	app.SetFocus(form)
}
