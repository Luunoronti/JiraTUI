package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/config"
	"jiratui/jira"
)

type App struct {
	tapp             *tview.Application
	mainWindow       *mainWindow
	issueList        *IssueList
	navPanel         *NavPanel
	jqlBar           *JqlBar
	cfg              *config.AppConfig
	client           jira.Client
	version          string
	currentJql       string
	navVisible       bool
	detailVisible    bool
	detailFullscreen bool
	jqlVisible       bool
	currentIssue     *jira.Issue
}

func Run(cfg *config.AppConfig, client jira.Client, version string) error {
	tapp := tview.NewApplication()

	history := config.NewJqlHistory()
	_ = history.Load() // best-effort; missing file is fine

	il := NewIssueList(tapp, cfg.Columns)
	nav := NewNavPanel()
	jqlBar := NewJqlBar(history)
	mw := newMainWindow(tapp, il, nav, jqlBar)

	app := &App{
		tapp:       tapp,
		mainWindow: mw,
		issueList:  il,
		navPanel:   nav,
		jqlBar:     jqlBar,
		cfg:        cfg,
		client:     client,
		version:    version,
		currentJql: cfg.Behavior.DefaultJql,
	}

	// When the terminal goes too-small the BeforeDrawFunc hides overlays;
	// sync our flags without queuing extra draws.
	mw.onNavClose = func() {
		app.navVisible = false
		app.jqlVisible = false
	}

	// Nav callbacks.
	nav.OnSelect = func(jql string) {
		app.closeNav()
		app.loadIssues(jql)
	}
	nav.OnClose = func() {
		app.closeNav()
	}

	// JQL bar callbacks.
	jqlBar.OnSubmit = func(jql string) {
		app.closeJql()
		app.loadIssues(jql)
	}
	jqlBar.OnClose = func() {
		app.closeJql()
	}

	// Issue list selection change → update detail panel.
	il.OnSelectionChange = func(issue jira.Issue) {
		app.showIssueInDetail(issue)
	}

	// Load initial issues.
	jqlBar.SetText(cfg.Behavior.DefaultJql)
	app.loadIssues(cfg.Behavior.DefaultJql)

	// Load nav data in background so the UI is immediately responsive.
	go func() {
		projects, _ := client.GetProjects()
		filters, _ := client.GetSavedFilters()
		tapp.QueueUpdateDraw(func() {
			nav.Populate(projects, filters)
		})
	}()

	// Global key handler.
	tapp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			tapp.Stop()
			return nil
		case tcell.KeyCtrlB:
			app.toggleNav()
			return nil
		case tcell.KeyCtrlD:
			app.toggleDetail()
			return nil
		case tcell.KeyCtrlJ:
			app.toggleJql()
			return nil
		case tcell.KeyEnter:
			// Enter opens fullscreen detail when issue list has focus and
			// we are not already in fullscreen or JQL bar.
			if !app.detailFullscreen && !app.jqlVisible {
				app.openDetailFull()
				return nil
			}
		case tcell.KeyEscape:
			if app.detailFullscreen {
				app.closeDetailFull()
				return nil
			}
			if app.detailVisible {
				app.closeDetail()
				return nil
			}
			if app.jqlVisible {
				// JQL bar handles Escape internally (clear vs close);
				// don't intercept here — let it reach the input field.
			}
		}
		return event
	})

	return tapp.SetRoot(mw.Primitive(), true).EnableMouse(false).Run()
}

// ─── detail ───────────────────────────────────────────────────────────────────

func (app *App) showIssueInDetail(issue jira.Issue) {
	cp := issue
	app.currentIssue = &cp
	if app.detailVisible {
		_, _, termW, _ := app.mainWindow.pages.GetRect()
		tier := SizeTier(termW)
		panelW := DetailWidth(tier, termW)
		app.mainWindow.detailPanel.SetIssue(app.currentIssue, panelW)
	}
	if app.detailFullscreen {
		_, _, termW, _ := app.mainWindow.pages.GetRect()
		app.mainWindow.detailFull.SetIssue(app.currentIssue, termW)
	}
}

func (app *App) toggleDetail() {
	_, _, termW, _ := app.mainWindow.pages.GetRect()
	tier := SizeTier(termW)
	if tier == Compact {
		if app.detailFullscreen {
			app.closeDetailFull()
		} else {
			app.openDetailFull()
		}
		return
	}
	if app.detailVisible {
		app.closeDetail()
	} else {
		app.openDetailSide()
	}
}

func (app *App) openDetailSide() {
	_, _, termW, _ := app.mainWindow.pages.GetRect()
	tier := SizeTier(termW)
	panelW := DetailWidth(tier, termW)
	if panelW <= 0 {
		app.openDetailFull()
		return
	}
	app.mainWindow.detailPanel.SetIssue(app.currentIssue, panelW)
	app.detailVisible = true
	app.mainWindow.showDetailSide()
	app.tapp.SetFocus(app.issueList)
}

func (app *App) closeDetail() {
	app.detailVisible = false
	app.mainWindow.hideDetailSide()
	app.tapp.SetFocus(app.issueList)
}

func (app *App) openDetailFull() {
	_, _, termW, _ := app.mainWindow.pages.GetRect()
	app.mainWindow.detailFull.SetIssue(app.currentIssue, termW)
	if app.navVisible {
		app.closeNav()
	}
	app.detailFullscreen = true
	app.mainWindow.showDetailFull()
	app.tapp.SetFocus(app.mainWindow.detailFull)
}

func (app *App) closeDetailFull() {
	app.detailFullscreen = false
	app.mainWindow.hideDetailFull()
	app.tapp.SetFocus(app.issueList)
}

// ─── nav ──────────────────────────────────────────────────────────────────────

func (app *App) toggleNav() {
	if app.navVisible {
		app.closeNav()
	} else {
		app.openNav()
	}
}

func (app *App) openNav() {
	app.navVisible = true
	app.mainWindow.showNav()
	app.tapp.SetFocus(app.navPanel)
}

func (app *App) closeNav() {
	app.navVisible = false
	app.mainWindow.hideNav()
	app.tapp.SetFocus(app.issueList)
}

// ─── JQL bar ──────────────────────────────────────────────────────────────────

func (app *App) toggleJql() {
	if app.jqlVisible {
		app.closeJql()
	} else {
		app.openJql()
	}
}

func (app *App) openJql() {
	app.jqlVisible = true
	app.mainWindow.showJql()
	app.jqlBar.FocusAndSelectAll(app.tapp)
}

func (app *App) closeJql() {
	app.jqlVisible = false
	app.mainWindow.hideJql()
	app.tapp.SetFocus(app.issueList)
}

// ─── issues ───────────────────────────────────────────────────────────────────

func (app *App) loadIssues(jql string) {
	issues, _, err := app.client.SearchIssues(jql, app.cfg.Behavior.PageSize)
	if err != nil {
		return
	}
	app.issueList.SetIssues(issues)
	app.mainWindow.currentJql = jql
	app.currentJql = jql
}
