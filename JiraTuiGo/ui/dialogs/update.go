package dialogs

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/updater"
)

const updatePageName = "dialog-update"

// ShowUpdateDialog shows a centered update dialog that drives the download/verify/apply flow.
func ShowUpdateDialog(
	app *tview.Application,
	pages *tview.Pages,
	currentVersion string,
	release *updater.Release,
	assetName string,
	onClose func(),
) {
	const (
		dialogWidth  = 62
		dialogHeight = 16
		barWidth     = 24
	)

	// ── state ──────────────────────────────────────────────────────────────
	type phase int
	const (
		phaseIdle phase = iota
		phaseRunning
		phaseDone
	)
	currentPhase := phaseIdle
	pct := 0
	_ = pct

	close := func() {
		pages.RemovePage(updatePageName)
		if onClose != nil {
			onClose()
		}
	}

	// ── widgets ────────────────────────────────────────────────────────────
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(" Update Available ")
	form.SetBorderPadding(1, 1, 2, 2)

	// Version info lines (read-only text fields rendered as plain TextViews
	// inside a Flex so they look like labels).
	versionText := tview.NewTextView()
	versionText.SetDynamicColors(true)
	versionText.SetText(
		fmt.Sprintf("  Current:   [::b]%s[::-]\n  Available: [::b]%s[::-]",
			currentVersion, release.Version),
	)

	progressText := tview.NewTextView()
	progressText.SetDynamicColors(true)
	progressText.SetText("")

	statusText := tview.NewTextView()
	statusText.SetDynamicColors(true)
	statusText.SetText("")

	updateBtn := tview.NewButton("  Update  ")
	cancelBtn := tview.NewButton("  Cancel  ")

	// Focus ring: updateBtn <-> cancelBtn
	updateBtn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if currentPhase == phaseRunning {
			return nil // swallow keys while running
		}
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyRight:
			app.SetFocus(cancelBtn)
			return nil
		case tcell.KeyBacktab, tcell.KeyLeft:
			app.SetFocus(cancelBtn)
			return nil
		case tcell.KeyEscape:
			close()
			return nil
		}
		return event
	})

	cancelBtn.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyRight:
			if currentPhase != phaseRunning {
				app.SetFocus(updateBtn)
			}
			return nil
		case tcell.KeyBacktab, tcell.KeyLeft:
			if currentPhase != phaseRunning {
				app.SetFocus(updateBtn)
			}
			return nil
		case tcell.KeyEscape:
			close()
			return nil
		}
		return event
	})

	// setStatus safely updates status text from any goroutine.
	setProgress := func(downloaded, total int64) {
		var line string
		if total > 0 {
			p := int(downloaded * 100 / total)
			if p > 100 {
				p = 100
			}
			line = fmt.Sprintf("  Progress: %s %d%%", progressBar(p, barWidth), p)
		} else {
			kb := downloaded / 1024
			line = fmt.Sprintf("  Progress: %d KB downloaded", kb)
		}
		app.QueueUpdateDraw(func() {
			progressText.SetText(line)
		})
	}

	setStatus := func(msg string) {
		app.QueueUpdateDraw(func() {
			statusText.SetText("  Status: " + msg)
		})
	}

	// runUpdate is kicked off when the user clicks [Update].
	runUpdate := func() {
		currentPhase = phaseRunning
		// Hide buttons, show progress area.
		app.QueueUpdateDraw(func() {
			updateBtn.SetLabel("          ")
			cancelBtn.SetLabel("  Close   ")
			app.SetFocus(cancelBtn)
		})

		go func() {
			// Download.
			setStatus("Downloading...")
			data, err := updater.Download(release.DownloadURL, setProgress)
			if err != nil {
				setStatus("[red]✗ Error: " + err.Error() + "[-]")
				currentPhase = phaseDone
				return
			}

			// Verify checksum (only if we have a checksum URL).
			if release.ChecksumURL != "" && assetName != "" {
				setStatus("Verifying...")
				ok, err := updater.VerifyChecksum(data, release.ChecksumURL, assetName)
				if err != nil {
					setStatus("[red]✗ Checksum error: " + err.Error() + "[-]")
					currentPhase = phaseDone
					return
				}
				if !ok {
					setStatus("[red]✗ Checksum mismatch — aborting[-]")
					currentPhase = phaseDone
					return
				}
			}

			// Apply.
			setStatus("Applying...")
			if err := updater.Apply(data); err != nil {
				setStatus("[red]✗ Apply error: " + err.Error() + "[-]")
				currentPhase = phaseDone
				return
			}

			currentPhase = phaseDone
			if runtime.GOOS == "windows" {
				setStatus("[green]✓ Written to jiratui.exe.new — restart with new binary[-]")
			} else {
				setStatus("[green]✓ Update applied. Please restart.[-]")
			}
		}()
	}

	updateBtn.SetSelectedFunc(func() {
		if currentPhase != phaseIdle {
			return
		}
		runUpdate()
	})

	cancelBtn.SetSelectedFunc(func() {
		if currentPhase == phaseRunning {
			return // can't cancel mid-download in this implementation
		}
		close()
	})

	// ── layout ─────────────────────────────────────────────────────────────
	// We build a Flex layout manually since tview.Form's AddItem doesn't give us
	// enough control over spacing and read-only text areas.
	btnFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(updateBtn, 12, 0, true).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(cancelBtn, 12, 0, false).
		AddItem(tview.NewBox(), 0, 1, false)

	_ = form // not used directly; we build our own Flex

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(versionText, 2, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(progressText, 1, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(statusText, 1, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(btnFlex, 1, 0, true)

	frame := tview.NewFrame(layout).
		SetBorders(1, 1, 1, 1, 2, 2)
	frame.SetBorder(true)
	frame.SetTitle(" Update Available ")

	// Global Escape to dismiss from anywhere in the dialog.
	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			if currentPhase != phaseRunning {
				close()
			}
			return nil
		}
		return event
	})

	modal := centeredBox(frame, dialogWidth, dialogHeight)
	pages.AddPage(updatePageName, modal, true, true)
	app.SetFocus(updateBtn)
}

// progressBar renders a simple ASCII progress bar.
// e.g. progressBar(42, 20) → "[████████░░░░░░░░░░░░]"
func progressBar(pct int, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := width * pct / 100
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}
