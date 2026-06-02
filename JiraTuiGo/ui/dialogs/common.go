package dialogs

import "github.com/rivo/tview"

// centeredBox wraps content in a grid that centers it on screen.
func centeredBox(content tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(content, 1, 1, 1, 1, 0, 0, true)
}
