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

	// Status bar
	mw.statusBar = tview.NewTextView()
	mw.statusBar.SetDynamicColors(true)
	mw.statusBar.SetBackgroundColor(statusBg)
	mw.statusBar.SetTextColor(statusFg)
	mw.updateStatusBar(120)

	// Main flex layout
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mw.menuBar, 1, 0, false).
		AddItem(mw.issueList, 0, 1, true).
		AddItem(mw.statusBar, 1, 0, false)

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

	mw.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
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
	hints := []Hint{
		{"F2", "Settings"},
		{"Ctrl-R", "Refresh"},
		{"Ctrl-B", "Nav"},
		{"Ctrl-D", "Detail"},
		{"Ctrl-J", "JQL"},
		{"Ctrl-G", "AI"},
		{"Ctrl-L", "Legend"},
		{"Ctrl-Q", "Quit"},
	}
	text := StatusBarHints(width, hints)

	issueCount := ""
	if mw.issueList != nil {
		issueCount = "  " + mw.issueList.statusLine()
	}

	mw.statusBar.SetText(" " + text + issueCount)
}

func (mw *mainWindow) Primitive() tview.Primitive {
	return mw.pages
}
