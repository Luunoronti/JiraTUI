package dialogs

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/config"
)

const settingsPageName = "dialog-settings"

// ShowSettingsDialog displays a multi-section settings dialog.
//
// Parameters:
//   - app        – tview Application
//   - pages      – top-level Pages used for layering
//   - cfg        – current configuration (read-only; a new copy is passed to onSave)
//   - themes     – list of available theme names (from themes.AvailableThemes())
//   - currentTheme – name of the currently active theme
//   - testConn   – callback to test a Jira connection; returns display name or error
//   - onSave     – called with the updated config when the user clicks Save
//   - onClose    – called when the dialog is closed without saving
func ShowSettingsDialog(
	app *tview.Application,
	pages *tview.Pages,
	cfg *config.AppConfig,
	themeList []string,
	currentTheme string,
	testConn func(baseURL, email, token string) (string, error),
	onSave func(newCfg *config.AppConfig),
	onClose func(),
) {
	_, _, termW, termH := pages.GetRect()
	if termW < 40 {
		termW = 80
	}
	if termH < 20 {
		termH = 30
	}
	width := min(80, termW-4)
	height := min(30, termH-4)

	// ── shared close helpers ──────────────────────────────────────────────────
	closeDialog := func() {
		pages.RemovePage(settingsPageName)
		if onClose != nil {
			onClose()
		}
	}

	// ── section names ─────────────────────────────────────────────────────────
	sections := []string{"Connection", "Appearance", "Behavior", "AI"}

	// ── mutable copies of every field ─────────────────────────────────────────
	// Connection
	connBaseURL := cfg.Conn.BaseURL
	connEmail := cfg.Conn.Email
	connToken := config.Unprotect(cfg.Conn.TokenProtected)

	// Appearance
	appearTheme := currentTheme

	// Behavior
	behavJql := cfg.Behavior.DefaultJql
	behavPageSize := fmt.Sprintf("%d", cfg.Behavior.PageSize)

	// AI
	aiAdapter := cfg.AI.Adapter
	aiBaseURL := cfg.AI.BaseURL
	aiModel := cfg.AI.Model
	aiKey := config.Unprotect(cfg.AI.ApiKeyProtected)

	// ── test-connection status label (Connection page) ─────────────────────
	testStatusView := tview.NewTextView()
	testStatusView.SetDynamicColors(true)
	testStatusView.SetText("")

	// ── build forms ──────────────────────────────────────────────────────────

	// Connection form
	connForm := tview.NewForm()
	connForm.SetBorder(false)
	connForm.AddInputField("Base URL", connBaseURL, 0, nil, func(v string) { connBaseURL = v })
	connForm.AddInputField("Email", connEmail, 0, nil, func(v string) { connEmail = v })
	connForm.AddPasswordField("API Token", connToken, 0, '*', func(v string) { connToken = v })
	// tokenFieldIdx is the form-item index of the API Token field (0-based, after Base URL and Email).
	const tokenFieldIdx = 2
	tokenVisible := false
	connForm.AddButton("Show Token", func() {
		item := connForm.GetFormItem(tokenFieldIdx)
		field, ok := item.(*tview.InputField)
		if !ok {
			return
		}
		tokenVisible = !tokenVisible
		if tokenVisible {
			field.SetMaskCharacter(0)
			connForm.GetButton(0).SetLabel("Hide Token") // button index 0 in connForm
		} else {
			field.SetMaskCharacter('*')
			connForm.GetButton(0).SetLabel("Show Token")
		}
	})
	connForm.AddButton("Test Connection", func() {
		go func() {
			name, err := testConn(connBaseURL, connEmail, connToken)
			app.QueueUpdateDraw(func() {
				if err != nil {
					testStatusView.SetText("[red]✗ Error: " + err.Error() + "[-]")
				} else {
					testStatusView.SetText("[green]✓ Connected as " + name + "[-]")
				}
			})
		}()
	})

	connPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(connForm, 0, 1, true).
		AddItem(testStatusView, 1, 0, false)

	// Appearance form
	appearForm := tview.NewForm()
	appearForm.SetBorder(false)
	// Find index of currentTheme in themeList
	themeIdx := 0
	for i, t := range themeList {
		if t == currentTheme {
			themeIdx = i
			break
		}
	}
	appearForm.AddDropDown("Theme", themeList, themeIdx, func(option string, _ int) {
		appearTheme = option
	})

	// Behavior form
	behavForm := tview.NewForm()
	behavForm.SetBorder(false)
	behavForm.AddInputField("Default JQL", behavJql, 0, nil, func(v string) { behavJql = v })
	behavForm.AddInputField("Page Size", behavPageSize, 10, tview.InputFieldInteger, func(v string) { behavPageSize = v })

	// AI form
	aiAdapterOptions := []string{"anthropic", "openai-compatible"}
	aiAdapterIdx := 0
	for i, a := range aiAdapterOptions {
		if a == aiAdapter {
			aiAdapterIdx = i
			break
		}
	}
	aiForm := tview.NewForm()
	aiForm.SetBorder(false)
	aiForm.AddDropDown("Adapter", aiAdapterOptions, aiAdapterIdx, func(option string, _ int) {
		aiAdapter = option
	})
	aiForm.AddInputField("Base URL", aiBaseURL, 0, nil, func(v string) { aiBaseURL = v })
	aiForm.AddInputField("Model", aiModel, 0, nil, func(v string) { aiModel = v })
	aiForm.AddPasswordField("API Key", aiKey, 0, '*', func(v string) { aiKey = v })
	// aiKeyFieldIdx: DropDown(0) + Base URL(1) + Model(2) + API Key(3)
	const aiKeyFieldIdx = 3
	aiKeyVisible := false
	aiForm.AddButton("Show Key", func() {
		item := aiForm.GetFormItem(aiKeyFieldIdx)
		field, ok := item.(*tview.InputField)
		if !ok {
			return
		}
		aiKeyVisible = !aiKeyVisible
		if aiKeyVisible {
			field.SetMaskCharacter(0)
			aiForm.GetButton(0).SetLabel("Hide Key")
		} else {
			field.SetMaskCharacter('*')
			aiForm.GetButton(0).SetLabel("Show Key")
		}
	})

	// ── right-side pages (one per section) ───────────────────────────────────
	rightPages := tview.NewPages()
	rightPages.AddPage("Connection", connPage, true, true)
	rightPages.AddPage("Appearance", appearForm, true, false)
	rightPages.AddPage("Behavior", behavForm, true, false)
	rightPages.AddPage("AI", aiForm, true, false)

	// ── left sidebar list ─────────────────────────────────────────────────────
	sideList := tview.NewList()
	sideList.ShowSecondaryText(false)
	sideList.SetBorder(false)
	for _, s := range sections {
		name := s // capture
		sideList.AddItem(name, "", 0, nil)
	}
	sideList.SetChangedFunc(func(idx int, _ string, _ string, _ rune) {
		rightPages.SwitchToPage(sections[idx])
	})

	// ── horizontal content area ───────────────────────────────────────────────
	contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(sideList, 16, 0, true).
		AddItem(tview.NewBox(), 1, 0, false). // spacer
		AddItem(rightPages, 0, 1, false)

	// ── Save / Cancel buttons ─────────────────────────────────────────────────
	saveBtn := tview.NewButton("Save")
	cancelBtn := tview.NewButton("Cancel")

	saveBtn.SetSelectedFunc(func() {
		// Parse page size
		ps, err := strconv.Atoi(behavPageSize)
		if err != nil || ps <= 0 {
			ps = cfg.Behavior.PageSize
		}

		// Protect secrets only if non-empty
		var protectedToken string
		if connToken != "" {
			protectedToken = config.Protect(connToken)
		}
		var protectedAIKey string
		if aiKey != "" {
			protectedAIKey = config.ProtectAIKey(aiKey)
		}

		newCfg := &config.AppConfig{
			Conn: config.ConnectionConfig{
				BaseURL:        connBaseURL,
				Email:          connEmail,
				TokenProtected: protectedToken,
				AuthType:       cfg.Conn.AuthType,
			},
			Appearance: config.AppearanceConfig{
				ThemeName: appearTheme,
			},
			Behavior: config.BehaviorConfig{
				DefaultJql:         behavJql,
				PageSize:           ps,
				AutoRefreshSeconds: cfg.Behavior.AutoRefreshSeconds,
			},
			AI: config.AiConfig{
				Adapter:         aiAdapter,
				BaseURL:         aiBaseURL,
				Model:           aiModel,
				ApiKeyProtected: protectedAIKey,
			},
			Columns: cfg.Columns,
		}

		pages.RemovePage(settingsPageName)
		if onSave != nil {
			onSave(newCfg)
		}
	})

	cancelBtn.SetSelectedFunc(func() {
		closeDialog()
	})

	btnFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false). // push buttons right
		AddItem(saveBtn, 8, 0, false).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(cancelBtn, 10, 0, false).
		AddItem(tview.NewBox(), 2, 0, false)

	// ── outer frame ───────────────────────────────────────────────────────────
	outerFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(contentFlex, 0, 1, true).
		AddItem(btnFlex, 1, 0, false)

	frame := tview.NewFrame(outerFlex).SetBorders(1, 1, 1, 1, 2, 2)
	frame.SetBorder(true).SetTitle(" Settings ")

	// ── focus helpers ────────────────────────────────────────────────────────
	activeForm := func() *tview.Form {
		name, _ := rightPages.GetFrontPage()
		switch name {
		case "Connection":
			return connForm
		case "Appearance":
			return appearForm
		case "Behavior":
			return behavForm
		case "AI":
			return aiForm
		}
		return connForm
	}

	// itemHasFocus checks whether a tview.FormItem (interface) currently has
	// focus by type-asserting to the concrete types that implement HasFocus().
	itemHasFocus := func(item tview.FormItem) bool {
		switch v := item.(type) {
		case *tview.InputField:
			return v.HasFocus()
		case *tview.DropDown:
			return v.HasFocus()
		case *tview.Checkbox:
			return v.HasFocus()
		case *tview.TextView:
			return v.HasFocus()
		}
		return false
	}

	// isLastFormControlFocused returns true when the very last control of the
	// active form has focus — used to decide when Tab should leave the form.
	// "Last control" means the last button when buttons exist, otherwise the
	// last form item (Behavior and Appearance have no buttons).
	isLastFormControlFocused := func() bool {
		f := activeForm()
		if n := f.GetButtonCount(); n > 0 {
			return f.GetButton(n - 1).HasFocus()
		}
		if n := f.GetFormItemCount(); n > 0 {
			return itemHasFocus(f.GetFormItem(n - 1))
		}
		return false
	}

	// focusLastFormControl focuses the last button (or last item if no buttons)
	// of the active form so that Shift-Tab from Save lands on the final control.
	focusLastFormControl := func() {
		f := activeForm()
		if n := f.GetButtonCount(); n > 0 {
			app.SetFocus(f.GetButton(n - 1))
			return
		}
		n := f.GetFormItemCount()
		if n == 0 {
			app.SetFocus(f)
			return
		}
		item := f.GetFormItem(n - 1)
		switch v := item.(type) {
		case *tview.InputField:
			app.SetFocus(v)
		case *tview.DropDown:
			app.SetFocus(v)
		default:
			app.SetFocus(f)
		}
	}

	// ── Tab / Shift-Tab routing ───────────────────────────────────────────────
	// Rules (forward Tab):
	//   sideList → active form (first field)
	//   form internals → handled by tview.Form naturally
	//   form last button → Save
	//   Save → Cancel
	//   Cancel → sideList
	//
	// Rules (Shift-Tab):
	//   sideList → Cancel
	//   Cancel → Save
	//   Save → active form last button
	//   form internals → handled by tview.Form naturally
	//   (when form wraps from first to last internally, that is also fine)
	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if sideList.HasFocus() {
				app.SetFocus(activeForm())
				return nil
			}
			if isLastFormControlFocused() {
				app.SetFocus(saveBtn)
				return nil
			}
			if saveBtn.HasFocus() {
				app.SetFocus(cancelBtn)
				return nil
			}
			if cancelBtn.HasFocus() {
				app.SetFocus(sideList)
				return nil
			}
			// Form has focus on a non-last control — let tview.Form handle Tab.
			return event

		case tcell.KeyBacktab:
			if sideList.HasFocus() {
				app.SetFocus(cancelBtn)
				return nil
			}
			if cancelBtn.HasFocus() {
				app.SetFocus(saveBtn)
				return nil
			}
			if saveBtn.HasFocus() {
				focusLastFormControl()
				return nil
			}
			// Form has focus — let tview.Form handle Shift-Tab internally.
			return event

		case tcell.KeyEscape:
			closeDialog()
			return nil
		}
		return event
	})

	modal := centeredBox(frame, width, height)
	pages.AddPage(settingsPageName, modal, true, true)
	app.SetFocus(sideList)
}
