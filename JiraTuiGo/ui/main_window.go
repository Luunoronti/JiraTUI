package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/themes"
)

type mainWindow struct {
	pages        *tview.Pages
	issueList    *IssueList
	nav          *NavPanel
	detailPanel  *DetailPanel
	detailFull   *DetailFullView
	jqlBar       *JqlBar
	statusBar    *tview.TextView
	menuBar      *tview.TextView
	app          *tview.Application
	currentJql   string
	updateVersion string // "" = none; "v0.2.1" = update available; "restart" = restart needed

	// onNavClose is called when the too-small transition hides the nav page,
	// so the App can sync its navVisible flag without queuing extra draws.
	onNavClose func()
}

func newMainWindow(app *tview.Application, il *IssueList, nav *NavPanel, jqlBar *JqlBar) *mainWindow {
	mw := &mainWindow{
		pages:       tview.NewPages(),
		app:         app,
		issueList:   il,
		nav:         nav,
		detailPanel: NewDetailPanel(),
		detailFull:  NewDetailFullView(),
		jqlBar:      jqlBar,
	}
	mw.build()
	return mw
}

func (mw *mainWindow) showNav() {
	mw.pages.ShowPage("nav")
}

func (mw *mainWindow) hideNav() {
	mw.pages.HidePage("nav")
}

func (mw *mainWindow) showDetailSide() {
	mw.pages.ShowPage("detail")
}

func (mw *mainWindow) hideDetailSide() {
	mw.pages.HidePage("detail")
}

func (mw *mainWindow) showDetailFull() {
	mw.pages.ShowPage("detail-full")
}

func (mw *mainWindow) hideDetailFull() {
	mw.pages.HidePage("detail-full")
}

func (mw *mainWindow) showJql() {
	mw.pages.ShowPage("jql")
}

func (mw *mainWindow) hideJql() {
	mw.pages.HidePage("jql")
}

func (mw *mainWindow) build() {
	t := themes.Current()
	bg := themes.C(t.Background)
	fg := themes.C(t.TextNormal)
	keyFg := themes.C(t.StatusKeyFg)
	statusBg := themes.C(t.StatusBg)
	statusFg := themes.C(t.StatusFg)
	_ = keyFg

	// Menu bar
	mw.menuBar = tview.NewTextView()
	mw.menuBar.SetDynamicColors(true)
	mw.menuBar.SetBackgroundColor(bg)
	mw.menuBar.SetTextColor(fg)
	menuText := fmt.Sprintf("[::b]F2[::-]:Settings  [::b]Ctrl-B[::-]:Nav  [::b]Ctrl-D[::-]:Detail  [::b]Ctrl-J[::-]:JQL  [::b]Ctrl-R[::-]:Refresh  [::b]Ctrl-Q[::-]:Quit")
	_ = menuText
	mw.menuBar.SetText(" JiraTUI")

	// Status bar — 2 rows allocated; second row used only when hints overflow.
	mw.statusBar = tview.NewTextView()
	mw.statusBar.SetDynamicColors(true)
	mw.statusBar.SetScrollable(false)
	mw.statusBar.SetWrap(false)
	mw.statusBar.SetBackgroundColor(statusBg)
	mw.statusBar.SetTextColor(statusFg)
	mw.updateStatusBar(120)

	// Main flex layout — no separate menu bar row; "JiraTUI" title is drawn
	// right-aligned inside the issue list header row (saves one screen row).
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mw.issueList, 0, 1, true).
		AddItem(mw.statusBar, 2, 0, false)

	// Too small view
	tooSmall := tview.NewTextView()
	tooSmall.SetTextAlign(tview.AlignCenter)
	tooSmall.SetDynamicColors(true)
	tooSmall.SetBackgroundColor(bg)
	tooSmall.SetTextColor(fg)

	mw.pages.AddPage("main", mainFlex, true, true)
	mw.pages.AddPage("toosmall", tooSmall, true, false)
	// nav page: resize=false (panel manages its own rect), initially hidden.
	mw.pages.AddPage("nav", mw.nav, false, false)
	// detail side panel: resize=false (panel manages its own rect), initially hidden.
	mw.pages.AddPage("detail", mw.detailPanel, false, false)
	// detail fullscreen: resize=true (occupies full terminal), initially hidden.
	mw.pages.AddPage("detail-full", mw.detailFull, true, false)
	// JQL bar: resize=false (self-positions at bottom in Draw), initially hidden.
	mw.pages.AddPage("jql", mw.jqlBar, false, false)

	// Guard state — calling SetText / SwitchToPage / ShowPage / HidePage inside
	// BeforeDrawFunc triggers SetNeedsDisplay, which queues another draw →
	// infinite redraw loop. Only call the mutating methods when values change.
	lastWidth := -1
	wasTooSmall := false
	lastMsg := ""
	colorDetected := false

	mw.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		// On the first draw, tcell's screen has been fully initialised and
		// screen.Colors() reflects the real terminal capability (it checks
		// COLORTERM, terminfo Tc/RGB, TERM_PROGRAM, etc.).  Override our
		// env-var pre-detection with tcell's authoritative answer.
		if !colorDetected {
			colorDetected = true
			if themes.DetectFromScreen(screen.Colors()) {
				themes.Apply(themes.Current())
				mw.refreshColors()
			}
		}

		w, h := screen.Size()
		if TooSmall(w, h) {
			msg := TooSmallMsg(w, h)
			if msg != lastMsg {
				lastMsg = msg
				tooSmall.SetText(msg)
			}
			if !wasTooSmall {
				wasTooSmall = true
				mw.pages.HidePage("main")
				mw.pages.HidePage("nav")
				mw.pages.HidePage("detail")
				mw.pages.HidePage("detail-full")
				mw.pages.HidePage("jql")
				mw.pages.ShowPage("toosmall")
				// Notify app to reset navVisible flag (no tview calls here).
				if mw.onNavClose != nil {
					mw.onNavClose()
				}
			}
		} else {
			if wasTooSmall {
				wasTooSmall = false
				mw.pages.HidePage("toosmall")
				mw.pages.ShowPage("main")
				// nav stays hidden; user re-opens with Ctrl-B if desired.
			}
			if w != lastWidth {
				lastWidth = w
				mw.updateStatusBar(w)
			}
		}
		return false
	})
}

func (mw *mainWindow) updateStatusBar(width int) {
	allHints := []Hint{
		{"F2", "Settings"},
		{"Ctrl-R", "Refresh"},
		{"Ctrl-B", "Nav"},
		{"Ctrl-D", "Detail"},
		{"Ctrl-J", "JQL"},
		{"Ctrl-\\", "Columns"},
		{"Ctrl-G", "AI"},
		{"Ctrl-L", "Legend"},
		{"Ctrl-Q", "Quit"},
	}

	issueCount := ""
	if mw.issueList != nil {
		issueCount = "  " + mw.issueList.statusLine()
	}

	updateIndicator := ""
	if mw.updateVersion != "" {
		if mw.updateVersion == "restart" {
			updateIndicator = "  ↑ Restart to apply update"
		} else {
			updateIndicator = "  ↑ " + mw.updateVersion + " — Ctrl-U"
		}
	}

	hints := StatusBarHints(width, allHints)
	meta := issueCount + updateIndicator

	// Try fitting everything on one line (status bar allocated 2 rows;
	// leave second row empty when one is enough).
	line1 := " " + hints + meta
	if len([]rune(line1)) <= width {
		mw.statusBar.SetText(line1 + "\n")
		return
	}

	// Overflow — split: shortcuts on line 1, meta (count + update) on line 2.
	// Re-abbreviate hints to leave room for the meta on line 2.
	mw.statusBar.SetText(" " + hints + "\n " + meta)
}

// SetUpdateAvailable sets the update indicator string and refreshes the status bar.
// Safe to call from any goroutine via QueueUpdateDraw.
func (mw *mainWindow) SetUpdateAvailable(version string) {
	mw.updateVersion = version
	_, _, w, _ := mw.pages.GetRect()
	if w <= 0 {
		w = 120 // fallback width before first draw
	}
	mw.updateStatusBar(w)
}

// refreshColors re-applies theme colors to tview-managed primitives (menu bar,
// status bar) after the color tier is updated from screen.Colors() on first draw.
// Custom primitives (IssueList, NavPanel, …) read themes.C() in their own
// Draw() calls so they update automatically without this.
func (mw *mainWindow) refreshColors() {
	t := themes.Current()
	mw.menuBar.SetBackgroundColor(themes.C(t.Background))
	mw.menuBar.SetTextColor(themes.C(t.TextNormal))
	mw.statusBar.SetBackgroundColor(themes.C(t.StatusBg))
	mw.statusBar.SetTextColor(themes.C(t.StatusFg))
}

func (mw *mainWindow) Primitive() tview.Primitive {
	return mw.pages
}
