package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"jiratui/config"
	"jiratui/jira"
)

type App struct {
	tapp        *tview.Application
	mainWindow  *mainWindow
	issueList   *IssueList
	navPanel    *NavPanel
	cfg         *config.AppConfig
	client      jira.Client
	version     string
	currentJql  string
	navVisible  bool
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
		}
		return event
	})

	return tapp.SetRoot(mw.Primitive(), true).EnableMouse(false).Run()
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
