package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/config"
	"jiratui/jira"
)

type App struct {
	tapp              *tview.Application
	mainWindow        *mainWindow
	issueList         *IssueList
	navPanel          *NavPanel
	cfg               *config.AppConfig
	client            jira.Client
	version           string
	currentJql        string
	navVisible        bool
	detailVisible     bool
	detailFullscreen  bool
	currentIssue      *jira.Issue
}

func Run(cfg *config.AppConfig, client jira.Client, version string) error {
	tapp := tview.NewApplication()

	il := NewIssueList(tapp, cfg.Columns)
	nav := NewNavPanel()
	mw := newMainWindow(tapp, il, nav)

	app := &App{
		tapp:       tapp,
		mainWindow: mw,
		issueList:  il,
		navPanel:   nav,
		cfg:        cfg,
		client:     client,
		version:    version,
		currentJql: cfg.Behavior.DefaultJql,
	}

	// When the terminal goes too-small the BeforeDrawFunc hides the nav page;
	// sync our flag without queuing extra draws.
	mw.onNavClose = func() {
		app.navVisible = false
	}

	// Nav callbacks.
	nav.OnSelect = func(jql string) {
		app.closeNav()
		app.loadIssues(jql)
	}
	nav.OnClose = func() {
		app.closeNav()
	}

	// Issue list selection change → update detail panel.
	il.OnSelectionChange = func(issue jira.Issue) {
		app.showIssueInDetail(issue)
	}

	// Load initial issues.
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
		case tcell.KeyEnter:
			// Enter opens fullscreen detail when issue list has focus and
			// we are not already in fullscreen.
			if !app.detailFullscreen {
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
		}
		return event
	})

	return tapp.SetRoot(mw.Primitive(), true).EnableMouse(false).Run()
}

// showIssueInDetail caches the current issue and pushes it to whichever
// detail view is currently visible (side panel or fullscreen).
// Safe to call from outside Draw.
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
		// In compact tier, Ctrl-D opens fullscreen instead.
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
		// Compact: fall through to fullscreen.
		app.openDetailFull()
		return
	}
	app.mainWindow.detailPanel.SetIssue(app.currentIssue, panelW)
	app.detailVisible = true
	app.mainWindow.showDetailSide()
	// ShowPage can transfer focus to the new visible page — explicitly keep
	// focus on the issue list so the side panel remains passive.
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
	// Close nav panel if open.
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

func (app *App) loadIssues(jql string) {
	issues, _, err := app.client.SearchIssues(jql, app.cfg.Behavior.PageSize)
	if err != nil {
		return
	}
	app.issueList.SetIssues(issues)
	app.mainWindow.currentJql = jql
	app.currentJql = jql
}

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
