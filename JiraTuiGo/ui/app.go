package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
	"jiratui/ai"
	"jiratui/config"
	"jiratui/jira"
	"jiratui/themes"
	"jiratui/ui/dialogs"
	"jiratui/updater"
)

type App struct {
	tapp             *tview.Application
	mainWindow       *mainWindow
	issueList        *IssueList
	navPanel         *NavPanel
	jqlBar           *JqlBar
	cfg              *config.AppConfig
	client           jira.Client
	aiClient         ai.AiClient // nil if not configured
	version          string
	repoOwner        string
	repoName         string
	latestRelease    *updater.Release
	updateAssetName  string
	currentJql       string
	navVisible       bool
	detailVisible    bool
	detailFullscreen bool
	jqlVisible       bool
	currentIssue     *jira.Issue
	modalOpen        int // >0 when a dialog is on screen
	issueMeta        *config.IssueMetaStore
	changelog        string
}

func Run(cfg *config.AppConfig, client jira.Client, version, repoOwner, repoName, changelog string) error {
	tapp := tview.NewApplication()

	history := config.NewJqlHistory()
	_ = history.Load() // best-effort; missing file is fine

	il := NewIssueList(tapp, cfg.Columns)
	nav := NewNavPanel()
	jqlBar := NewJqlBar(history)
	mw := newMainWindow(tapp, il, nav, jqlBar)

	meta := config.NewIssueMetaStore()
	_ = meta.Load() // best-effort
	il.SetMeta(meta)

	app := &App{
		tapp:       tapp,
		mainWindow: mw,
		issueList:  il,
		navPanel:   nav,
		jqlBar:     jqlBar,
		cfg:        cfg,
		client:     client,
		version:    version,
		repoOwner:  repoOwner,
		repoName:   repoName,
		currentJql: cfg.Behavior.DefaultJql,
		issueMeta:  meta,
		changelog:  changelog,
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

	// Auto-show What's New on first launch after version change.
	//
	// Why this three-step dance:
	//   1. QueueUpdateDraw() before Run() → sends on nil channel → hangs.
	//   2. SetAfterDrawFunc() → runs inside the draw lock; calling pages.AddPage()
	//      or SetFocus() from it tries to re-acquire the same lock → deadlocks.
	//   3. Solution: AfterDrawFunc signals a channel, a goroutine outside the lock
	//      calls QueueUpdateDraw safely once Run() is known to be running.
	if version != "dev" && version != cfg.Behavior.LastSeenVersion {
		cfg.Behavior.LastSeenVersion = version
		_ = cfg.Save()
		ready := make(chan struct{}, 1)
		tapp.SetAfterDrawFunc(func(_ tcell.Screen) {
			select {
			case ready <- struct{}{}: // signal once; goroutine drains it
			default:
			}
		})
		go func() {
			<-ready
			tapp.SetAfterDrawFunc(nil)
			tapp.QueueUpdateDraw(func() {
				app.showWhatsNew()
			})
		}()
	}

	// Load nav data in background so the UI is immediately responsive.
	go func() {
		projects, _ := client.GetProjects()
		filters, _ := client.GetSavedFilters()
		tapp.QueueUpdateDraw(func() {
			nav.Populate(projects, filters)
		})
	}()

	// Daily update check — runs in background, never crashes the app.
	go func() {
		if version != "dev" && time.Since(cfg.Behavior.LastUpdateCheck) < 15*time.Minute {
			return
		}
		release, err := updater.Check(app.repoOwner, app.repoName, version)
		if err != nil || release == nil {
			return
		}
		cfg.Behavior.LastUpdateCheck = time.Now()
		_ = cfg.Save()
		app.latestRelease = release
		assetName := updater.AssetName()
		app.updateAssetName = assetName
		tapp.QueueUpdateDraw(func() {
			mw.SetUpdateAvailable(release.Version)
		})
	}()

	// Global key handler.
	tapp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// F1 — What's New, always available when no other modal is open.
		if event.Key() == tcell.KeyF1 && app.modalOpen == 0 {
			app.showWhatsNew()
			return nil
		}

		// F2 is always available — even when another modal is open we allow
		// re-opening settings (guard is handled inside openSettings via modalOpen).
		if event.Key() == tcell.KeyF2 && app.modalOpen == 0 {
			app.openSettings()
			return nil
		}

		// Ctrl-G opens the AI JQL dialog (only when no modal is open).
		if event.Key() == tcell.KeyCtrlG && app.modalOpen == 0 {
			app.openAiJql()
			return nil
		}

		// Ctrl-U opens the update dialog when an update is available.
		if event.Key() == tcell.KeyCtrlU && app.modalOpen == 0 && app.latestRelease != nil {
			app.openUpdateDialog()
			return nil
		}

		// When a modal dialog is open, pass ALL keys through to the focused
		// primitive.  Without this the global handler consumes Enter (opening
		// fullscreen detail) and Escape (closing detail) before the dialog sees
		// them, leaving dialogs that can never be confirmed or closed.
		if app.modalOpen > 0 {
			return event
		}

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
		case tcell.KeyCtrlL:
			app.showLegend()
			return nil
		case tcell.KeyCtrlY:
			app.toggleHiddenView()
			return nil
		case tcell.KeyCtrlR:
			app.loadIssues(app.currentJql)
			return nil
		case tcell.KeyCtrlBackslash:
			app.showColumns()
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

		// Mutation shortcuts — only when nav/jql are not active and an issue is selected.
		if !app.navVisible && !app.jqlVisible && app.currentIssue != nil {
			switch event.Key() {
			case tcell.KeyCtrlP:
				app.changePriority()
				return nil
			case tcell.KeyCtrlT:
				app.changeStatus()
				return nil
			case tcell.KeyCtrlA:
				app.changeAssignee()
				return nil
			case tcell.KeyCtrlE:
				app.editDescription()
				return nil
			case tcell.KeyCtrlK:
				app.addComment()
				return nil
			case tcell.KeyCtrlO:
				app.openInBrowser()
				return nil
			case tcell.KeyCtrlF:
				app.saveAsFilter()
				return nil
			case tcell.KeyCtrlH:
				app.toggleHideIssue()
				return nil
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
	app.issueList.SetError("")
	issues, total, err := app.client.SearchIssues(jql, app.cfg.Behavior.PageSize)
	if err != nil {
		jira.Dbg("loadIssues error for jql=%q: %v", jql, err)
		// If the JQL is invalid (HTTP 400) and AI is configured, try to fix it automatically.
		if strings.Contains(err.Error(), "400") && app.cfg.AI.Adapter != "" {
			app.tryAiFallback(jql)
			return
		}
		app.issueList.SetError(err.Error())
		return
	}
	jira.Dbg("loadIssues jql=%q → %d/%d issues", jql, len(issues), total)
	app.issueList.SetIssues(issues)
	app.mainWindow.currentJql = jql
	app.currentJql = jql
}

// refreshCurrentIssue re-fetches the current issue from Jira and updates all panels.
func (app *App) refreshCurrentIssue() {
	if app.currentIssue == nil {
		return
	}
	fresh, err := app.client.GetIssue(app.currentIssue.Key)
	if err != nil || fresh == nil {
		return
	}
	app.currentIssue = fresh
	app.issueList.UpdateIssue(*fresh)
	if app.detailVisible {
		_, _, termW, _ := app.mainWindow.pages.GetRect()
		app.mainWindow.detailPanel.SetIssue(app.currentIssue, DetailWidth(SizeTier(termW), termW))
	}
	if app.detailFullscreen {
		_, _, termW, _ := app.mainWindow.pages.GetRect()
		app.mainWindow.detailFull.SetIssue(app.currentIssue, termW)
	}
}

// restoreFocus returns focus to the appropriate panel after a dialog closes.
func (app *App) restoreFocus() {
	if app.detailFullscreen {
		app.tapp.SetFocus(app.mainWindow.detailFull)
	} else {
		app.tapp.SetFocus(app.issueList)
	}
}

// ─── mutation actions ─────────────────────────────────────────────────────────

func (app *App) changePriority() {
	priorities, err := app.client.GetPriorities()
	if err != nil || len(priorities) == 0 {
		return
	}

	names := make([]string, len(priorities))
	for i, p := range priorities {
		names[i] = p.Name
	}

	// Find current priority index.
	initialIdx := 0
	if app.currentIssue != nil {
		for i, p := range priorities {
			if p.Name == app.currentIssue.Priority.Name {
				initialIdx = i
				break
			}
		}
	}

	issueKey := app.currentIssue.Key
	app.modalOpen++
	dialogs.ShowChoiceDialog(
		app.tapp,
		app.mainWindow.pages,
		"Change Priority — "+issueKey,
		names,
		initialIdx,
		func(idx int) {
			app.modalOpen--
			app.restoreFocus()
			if idx < 0 {
				return
			}
			go func() {
				_ = app.client.UpdatePriority(issueKey, priorities[idx].Name)
				app.tapp.QueueUpdateDraw(func() {
					app.refreshCurrentIssue()
				})
			}()
		},
	)
}

func (app *App) changeStatus() {
	if app.currentIssue == nil {
		return
	}
	issueKey := app.currentIssue.Key

	transitions, err := app.client.GetTransitions(issueKey)
	if err != nil || len(transitions) == 0 {
		return
	}

	names := make([]string, len(transitions))
	for i, t := range transitions {
		names[i] = t.Name
	}

	app.modalOpen++
	dialogs.ShowChoiceDialog(
		app.tapp,
		app.mainWindow.pages,
		"Change Status — "+issueKey,
		names,
		0,
		func(idx int) {
			app.modalOpen--
			app.restoreFocus()
			if idx < 0 {
				return
			}
			go func() {
				_ = app.client.DoTransition(issueKey, transitions[idx].ID)
				app.tapp.QueueUpdateDraw(func() {
					app.refreshCurrentIssue()
				})
			}()
		},
	)
}

func (app *App) changeAssignee() {
	if app.currentIssue == nil {
		return
	}
	issueKey := app.currentIssue.Key
	currentAssignee := ""
	if app.currentIssue.Assignee != nil {
		currentAssignee = app.currentIssue.Assignee.DisplayName
	}

	app.modalOpen++
	dialogs.ShowAssigneeDialog(
		app.tapp,
		app.mainWindow.pages,
		app.client,
		issueKey,
		currentAssignee,
		func(accountId string, saved bool) {
			app.modalOpen--
			app.restoreFocus()
			if !saved {
				return
			}
			go func() {
				_ = app.client.UpdateAssignee(issueKey, accountId)
				app.tapp.QueueUpdateDraw(func() {
					app.refreshCurrentIssue()
				})
			}()
		},
	)
}

func (app *App) editDescription() {
	if app.currentIssue == nil {
		return
	}
	issueKey := app.currentIssue.Key
	initial := app.currentIssue.Description

	app.modalOpen++
	dialogs.ShowTextEditorDialog(
		app.tapp,
		app.mainWindow.pages,
		"Edit Description — "+issueKey,
		initial,
		"Save",
		func(text string, saved bool) {
			app.modalOpen--
			app.restoreFocus()
			if !saved {
				return
			}
			go func() {
				_ = app.client.UpdateDescription(issueKey, text)
				app.tapp.QueueUpdateDraw(func() {
					app.refreshCurrentIssue()
				})
			}()
		},
	)
}

func (app *App) addComment() {
	if app.currentIssue == nil {
		return
	}
	issueKey := app.currentIssue.Key

	app.modalOpen++
	dialogs.ShowTextEditorDialog(
		app.tapp,
		app.mainWindow.pages,
		"Add Comment — "+issueKey,
		"",
		"Submit",
		func(text string, saved bool) {
			app.modalOpen--
			app.restoreFocus()
			if !saved || text == "" {
				return
			}
			go func() {
				_ = app.client.AddComment(issueKey, text)
				app.tapp.QueueUpdateDraw(func() {
					app.refreshCurrentIssue()
				})
			}()
		},
	)
}

func (app *App) openInBrowser() {
	if app.currentIssue == nil {
		return
	}
	baseURL := app.cfg.Conn.BaseURL
	if baseURL == "" {
		return
	}
	url := baseURL + "/browse/" + app.currentIssue.Key
	_ = browser.OpenURL(url)
}

func (app *App) saveAsFilter() {
	jql := app.currentJql
	app.modalOpen++
	dialogs.ShowSaveFilterDialog(
		app.tapp,
		app.mainWindow.pages,
		jql,
		func(name, description string, saved bool) {
			app.modalOpen--
			app.restoreFocus()
			if !saved || name == "" {
				return
			}
			go func() {
				_, err := app.client.SaveFilter(name, description, jql)
				if err != nil {
					return
				}
				// Refresh nav panel with updated filters.
				projects, _ := app.client.GetProjects()
				filters, _ := app.client.GetSavedFilters()
				app.tapp.QueueUpdateDraw(func() {
					app.navPanel.Populate(projects, filters)
				})
			}()
		},
	)
}

func (app *App) showLegend() {
	app.modalOpen++
	dialogs.ShowLegendDialog(app.tapp, app.mainWindow.pages, func() {
		app.modalOpen--
		app.restoreFocus()
	})
}

func (app *App) showColumns() {
	app.modalOpen++
	dialogs.ShowColumnsDialog(
		app.tapp,
		app.mainWindow.pages,
		app.cfg.Columns,
		func(cols config.ColumnVisibilityConfig) {
			app.modalOpen--
			app.cfg.Columns = cols
			_ = app.cfg.Save()
			app.issueList.SetColumns(cols)
			app.restoreFocus()
		},
		func() {
			app.modalOpen--
			app.restoreFocus()
		},
	)
}

// ─── what's new ───────────────────────────────────────────────────────────────

func (app *App) showWhatsNew() {
	app.modalOpen++
	dialogs.ShowWhatsNewDialog(
		app.tapp,
		app.mainWindow.pages,
		app.changelog,
		func() {
			app.modalOpen--
			app.restoreFocus()
		},
	)
}

// ─── update ───────────────────────────────────────────────────────────────────

func (app *App) openUpdateDialog() {
	app.modalOpen++
	dialogs.ShowUpdateDialog(
		app.tapp,
		app.mainWindow.pages,
		app.version,
		app.latestRelease,
		app.updateAssetName,
		func() {
			app.modalOpen--
			// Mark restart needed after a successful apply (onClose is called
			// regardless of outcome; the dialog already shows the correct status
			// text, so we just persist the indicator in the status bar).
			app.mainWindow.SetUpdateAvailable("restart")
			app.restoreFocus()
		},
	)
}

// ─── AI JQL ───────────────────────────────────────────────────────────────────

// openAiJql opens the AI JQL dialog.
func (app *App) openAiJql() {
	aiCfg := &app.cfg.AI
	if aiCfg.Adapter == "" || aiCfg.Model == "" {
		app.issueList.SetError("AI not configured — open Settings (F2) → AI tab")
		go func() {
			time.Sleep(3 * time.Second)
			app.tapp.QueueUpdateDraw(func() { app.issueList.SetError("") })
		}()
		return
	}

	client, err := ai.NewAiClient(aiCfg, config.Unprotect)
	if err != nil {
		app.issueList.SetError("AI error: " + err.Error())
		go func() {
			time.Sleep(3 * time.Second)
			app.tapp.QueueUpdateDraw(func() { app.issueList.SetError("") })
		}()
		return
	}

	// Build context for system prompt (best-effort, ignore errors).
	var projects []jira.Project
	var statuses []jira.Status
	var priorities []jira.Priority
	var issueTypes []jira.IssueType
	projects, _ = app.client.GetProjects()
	statuses, _ = app.client.GetStatuses()
	priorities, _ = app.client.GetPriorities()
	issueTypes, _ = app.client.GetIssueTypes()

	system := ai.BuildSystemPrompt(projects, statuses, priorities, issueTypes)

	app.modalOpen++
	dialogs.ShowAiJqlDialog(
		app.tapp,
		app.mainWindow.pages,
		app.currentJql,
		func(prompt string) (string, error) {
			return client.Generate(system, prompt)
		},
		func(jql string) {
			app.modalOpen--
			app.jqlBar.SetText(jql)
			app.loadIssues(jql)
			app.restoreFocus()
		},
		func() {
			app.modalOpen--
			app.restoreFocus()
		},
	)
}

// tryAiFallback silently attempts to translate an invalid JQL query via AI and
// retries the search with the translated query.
func (app *App) tryAiFallback(originalJql string) {
	client, err := ai.NewAiClient(&app.cfg.AI, config.Unprotect)
	if err != nil {
		return
	}
	projects, _ := app.client.GetProjects()
	statuses, _ := app.client.GetStatuses()
	priorities, _ := app.client.GetPriorities()
	issueTypes, _ := app.client.GetIssueTypes()
	system := ai.BuildSystemPrompt(projects, statuses, priorities, issueTypes)
	go func() {
		jql, err := client.Generate(system, originalJql)
		if err != nil {
			return
		}
		jql = strings.TrimSpace(jql)
		// Strip markdown fences if the model wrapped the output.
		jql = strings.TrimPrefix(jql, "```jql")
		jql = strings.TrimPrefix(jql, "```")
		jql = strings.TrimSuffix(jql, "```")
		jql = strings.TrimSpace(jql)
		app.tapp.QueueUpdateDraw(func() {
			app.jqlBar.SetText(jql)
			app.loadIssues(jql)
		})
	}()
}

// ─── settings ─────────────────────────────────────────────────────────────────

func (app *App) openSettings() {
	app.modalOpen++
	dialogs.ShowSettingsDialog(
		app.tapp,
		app.mainWindow.pages,
		app.cfg,
		themes.AvailableThemes(),
		themes.CurrentThemeName(),
		func(baseURL, email, token string) (string, error) {
			c := jira.NewRealClient(baseURL, email, token)
			user, err := c.GetCurrentUser()
			if err != nil {
				return "", err
			}
			return user.DisplayName, nil
		},
		func(newCfg *config.AppConfig) {
			app.modalOpen--
			app.cfg = newCfg
			_ = newCfg.Save() // best-effort
			themes.Switch(newCfg.Appearance.ThemeName)
			themes.Apply(themes.Current())
			app.rebuildClient()
			app.restoreFocus()
		},
		func() {
			app.modalOpen--
			app.restoreFocus()
		},
	)
}

// ─── hide issue ───────────────────────────────────────────────────────────────

func (app *App) toggleHideIssue() {
	if app.currentIssue == nil {
		return
	}
	app.issueMeta.ToggleHidden(app.currentIssue.Key)
	_ = app.issueMeta.Save()
	app.issueList.RebuildVisible()
}

func (app *App) toggleHiddenView() {
	app.issueList.ToggleShowHidden()
}

// rebuildClient recreates the Jira client from the current config and reloads
// all data. Called after settings are saved.
func (app *App) rebuildClient() {
	if app.cfg.Conn.BaseURL != "" && app.cfg.Conn.Email != "" && app.cfg.Conn.TokenProtected != "" {
		app.client = jira.NewRealClient(
			app.cfg.Conn.BaseURL,
			app.cfg.Conn.Email,
			config.Unprotect(app.cfg.Conn.TokenProtected),
		)
	} else {
		app.client = jira.NewMockClient()
	}

	// Reload nav data.
	go func() {
		projects, _ := app.client.GetProjects()
		filters, _ := app.client.GetSavedFilters()
		app.tapp.QueueUpdateDraw(func() {
			app.navPanel.Populate(projects, filters)
		})
	}()

	app.loadIssues(app.cfg.Behavior.DefaultJql)
	app.jqlBar.SetText(app.cfg.Behavior.DefaultJql)
}
