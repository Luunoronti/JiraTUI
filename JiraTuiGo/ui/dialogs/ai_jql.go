package dialogs

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const aiJqlPageName = "dialog-ai-jql"

// ShowAiJqlDialog shows a dialog that lets the user describe what they want to
// find in natural language and have the AI generate a JQL query.
//
//   - generate  – called (in a goroutine) with the user's prompt; returns JQL or error
//   - onUse     – called when the user clicks Use with the generated JQL
//   - onClose   – called when the dialog is dismissed without using the JQL
func ShowAiJqlDialog(
	app *tview.Application,
	pages *tview.Pages,
	initialJql string,
	generate func(prompt string) (string, error),
	onUse func(jql string),
	onClose func(),
) {
	const width = 70
	const height = 18

	var generatedJql string

	// ── helpers ──────────────────────────────────────────────────────────────

	closeDialog := func() {
		pages.RemovePage(aiJqlPageName)
		if onClose != nil {
			onClose()
		}
	}

	useDialog := func() {
		pages.RemovePage(aiJqlPageName)
		if onUse != nil {
			onUse(generatedJql)
		}
	}

	// ── widgets ───────────────────────────────────────────────────────────────

	promptInput := tview.NewInputField()
	promptInput.SetLabel("")
	promptInput.SetFieldWidth(0) // fills available width

	jqlView := tview.NewTextView()
	jqlView.SetDynamicColors(false)
	jqlView.SetScrollable(false)
	jqlView.SetWrap(true)
	jqlView.SetText(initialJql)

	statusView := tview.NewTextView()
	statusView.SetDynamicColors(true)
	statusView.SetText("")
	statusView.SetScrollable(false)

	generateBtn := tview.NewButton("Generate")
	useBtn := tview.NewButton("Use")
	closeBtn := tview.NewButton("Close")

	// Disable Use until JQL is generated.
	useBtn.SetDisabled(true)

	// ── focus order ───────────────────────────────────────────────────────────
	// Tab cycles: promptInput → generateBtn → useBtn → closeBtn → promptInput
	focusOrder := []tview.Primitive{promptInput, generateBtn, useBtn, closeBtn}

	focusNext := func(current tview.Primitive) {
		for i, p := range focusOrder {
			if p == current {
				app.SetFocus(focusOrder[(i+1)%len(focusOrder)])
				return
			}
		}
		app.SetFocus(focusOrder[0])
	}

	focusPrev := func(current tview.Primitive) {
		for i, p := range focusOrder {
			if p == current {
				n := (i - 1 + len(focusOrder)) % len(focusOrder)
				app.SetFocus(focusOrder[n])
				return
			}
		}
		app.SetFocus(focusOrder[len(focusOrder)-1])
	}

	tabCapture := func(current tview.Primitive) func(*tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTab:
				focusNext(current)
				return nil
			case tcell.KeyBacktab:
				focusPrev(current)
				return nil
			case tcell.KeyEscape:
				closeDialog()
				return nil
			}
			return event
		}
	}

	// ── Generate action ───────────────────────────────────────────────────────

	doGenerate := func() {
		prompt := promptInput.GetText()
		if prompt == "" {
			return
		}
		app.QueueUpdateDraw(func() {
			statusView.SetText("[yellow]Generating…[-]")
			useBtn.SetDisabled(true)
		})
		go func() {
			jql, err := generate(prompt)
			app.QueueUpdateDraw(func() {
				if err != nil {
					statusView.SetText("[red]Error: " + err.Error() + "[-]")
					return
				}
				generatedJql = jql
				jqlView.SetText(jql)
				statusView.SetText("[green]Done[-]")
				useBtn.SetDisabled(false)
			})
		}()
	}

	generateBtn.SetSelectedFunc(doGenerate)
	useBtn.SetSelectedFunc(useDialog)
	closeBtn.SetSelectedFunc(closeDialog)

	// Allow Enter in prompt field to trigger Generate.
	promptInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			doGenerate()
		}
	})

	// Apply tab/escape capture to each focusable widget.
	promptInput.SetInputCapture(tabCapture(promptInput))
	generateBtn.SetInputCapture(tabCapture(generateBtn))
	useBtn.SetInputCapture(tabCapture(useBtn))
	closeBtn.SetInputCapture(tabCapture(closeBtn))

	// ── layout ────────────────────────────────────────────────────────────────

	// Button row (right-aligned via a Flex).
	btnRow := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false). // spacer
		AddItem(generateBtn, 12, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(useBtn, 7, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(closeBtn, 9, 0, false)

	// Wrap prompt input in a bordered box.
	promptBox := tview.NewFrame(promptInput).
		SetBorders(1, 1, 0, 0, 1, 1)

	// Wrap JQL view in a bordered box.
	jqlBox := tview.NewFrame(jqlView).
		SetBorders(1, 1, 0, 0, 1, 1)

	inner := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Describe what you want to find:"), 1, 0, false).
		AddItem(promptBox, 3, 0, true).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(tview.NewTextView().SetText("Generated JQL:"), 1, 0, false).
		AddItem(jqlBox, 4, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(statusView, 1, 0, false).
		AddItem(btnRow, 1, 0, false)

	inner.SetBorder(true).SetTitle(" AI JQL ")

	modal := centeredBox(inner, width, height)
	pages.AddPage(aiJqlPageName, modal, true, true)
	app.SetFocus(promptInput)
}
