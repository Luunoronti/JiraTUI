package dialogs

import (
	"jiratui/jira"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const assigneePageName = "dialog-assignee"

// ShowAssigneeDialog shows a search-and-select dialog for picking an assignee.
// onDone is called with (accountId, saved). accountId="" means unassign.
func ShowAssigneeDialog(
	app *tview.Application,
	pages *tview.Pages,
	client jira.Client,
	issueKey string,
	currentAssignee string,
	onDone func(accountId string, saved bool),
) {
	const width = 60
	const height = 18

	// Extract project key from issue key (e.g. "PROJ-123" → "PROJ").
	projectKey := ""
	if idx := strings.Index(issueKey, "-"); idx > 0 {
		projectKey = issueKey[:idx]
	}

	var users []jira.JiraUser

	searchField := tview.NewInputField()
	searchField.SetLabel("Search: ")
	searchField.SetBorder(false)

	resultList := tview.NewList()
	resultList.ShowSecondaryText(false)
	resultList.SetBorder(false)

	done := func(accountId string, saved bool) {
		pages.RemovePage(assigneePageName)
		if onDone != nil {
			onDone(accountId, saved)
		}
	}

	populateList := func(results []jira.JiraUser) {
		users = results
		resultList.Clear()
		resultList.AddItem("(Unassign)", "", 0, nil)
		for _, u := range results {
			resultList.AddItem(u.DisplayName, u.AccountID, 0, nil)
		}
	}

	// Initial load with empty query.
	go func() {
		results, err := client.SearchAssignableUsers("", projectKey)
		app.QueueUpdateDraw(func() {
			if err == nil {
				populateList(results)
			} else {
				populateList(nil)
			}
		})
	}()

	// Search on each keystroke.
	searchField.SetChangedFunc(func(text string) {
		go func() {
			results, err := client.SearchAssignableUsers(text, projectKey)
			app.QueueUpdateDraw(func() {
				if err == nil {
					populateList(results)
				}
			})
		}()
	})

	// Down from search field → move to list.
	searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyTab:
			app.SetFocus(resultList)
			return nil
		case tcell.KeyEnter:
			// Search explicitly on Enter.
			query := searchField.GetText()
			go func() {
				results, err := client.SearchAssignableUsers(query, projectKey)
				app.QueueUpdateDraw(func() {
					if err == nil {
						populateList(results)
					}
					app.SetFocus(resultList)
				})
			}()
			return nil
		case tcell.KeyEscape:
			done("", false)
			return nil
		}
		return event
	})

	resultList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if resultList.GetCurrentItem() == 0 {
				app.SetFocus(searchField)
				return nil
			}
		case tcell.KeyBacktab:
			app.SetFocus(searchField)
			return nil
		case tcell.KeyEscape:
			done("", false)
			return nil
		}
		return event
	})

	resultList.SetSelectedFunc(func(idx int, mainText, secondaryText string, _ rune) {
		if idx == 0 {
			// "(Unassign)" — empty account ID.
			done("", true)
			return
		}
		// idx-1 because index 0 is "(Unassign)".
		if idx-1 < len(users) {
			done(users[idx-1].AccountID, true)
		}
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(searchField, 1, 0, true).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(resultList, 0, 1, false)

	frame := tview.NewFrame(flex).SetBorders(1, 1, 1, 1, 2, 2)
	frame.SetBorder(true).SetTitle(" Change Assignee ")

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			done("", false)
			return nil
		}
		return event
	})

	modal := centeredBox(frame, width, height)
	pages.AddPage(assigneePageName, modal, true, true)
	app.SetFocus(searchField)
}
