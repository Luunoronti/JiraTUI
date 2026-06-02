package main

import "time"

type Issue struct {
	Key      string
	Summary  string
	Type     string
	Priority string
	Status   string
	Assignee string
	Updated  time.Time
}

var dummyIssues = []Issue{
	{Key: "ERO-2084", Summary: "Roboto | Design | Demo | Suburbs level is available in the demo", Type: "Bug", Priority: "Highest", Status: "To Do", Assignee: "Wiktoryn Żerebecki", Updated: time.Now().Add(-2 * time.Hour)},
	{Key: "ERO-2079", Summary: "Fix login timeout on mobile devices", Type: "Bug", Priority: "High", Status: "In Progress", Assignee: "Anna Nowak", Updated: time.Now().Add(-5 * time.Hour)},
	{Key: "ERO-2071", Summary: "Add dark mode to settings panel", Type: "Story", Priority: "Medium", Status: "In Review", Assignee: "Piotr Wiśniewski", Updated: time.Now().Add(-1 * 24 * time.Hour)},
	{Key: "ERO-2065", Summary: "Database connection pool exhaustion under high load", Type: "Bug", Priority: "Highest", Status: "In Progress", Assignee: "Maria Lewandowska", Updated: time.Now().Add(-2 * 24 * time.Hour)},
	{Key: "ERO-2058", Summary: "Implement OAuth 2.0 login for enterprise customers", Type: "New Feature", Priority: "High", Status: "Ready For Build", Assignee: "Wiktoryn Żerebecki", Updated: time.Now().Add(-3 * 24 * time.Hour)},
	{Key: "ERO-2051", Summary: "Update dependencies to latest versions", Type: "Task", Priority: "Low", Status: "Done", Assignee: "Anna Nowak", Updated: time.Now().Add(-5 * 24 * time.Hour)},
	{Key: "ERO-2044", Summary: "Performance regression in search API after last deployment", Type: "Bug", Priority: "High", Status: "In Progress", Assignee: "Piotr Wiśniewski", Updated: time.Now().Add(-6 * 24 * time.Hour)},
	{Key: "ERO-2038", Summary: "Add pagination to user list — loads all users at once", Type: "Improvement", Priority: "Medium", Status: "Done", Assignee: "Maria Lewandowska", Updated: time.Now().Add(-7 * 24 * time.Hour)},
	{Key: "ERO-2031", Summary: "Fix broken image uploads on Firefox 118+", Type: "Bug", Priority: "Medium", Status: "To Do", Assignee: "Wiktoryn Żerebecki", Updated: time.Now().Add(-8 * 24 * time.Hour)},
	{Key: "ERO-2024", Summary: "Implement real-time notifications via WebSocket", Type: "New Feature", Priority: "High", Status: "In Progress", Assignee: "Anna Nowak", Updated: time.Now().Add(-9 * 24 * time.Hour)},
	{Key: "ERO-2017", Summary: "Refactor authentication middleware for testability", Type: "Task", Priority: "Medium", Status: "In Review", Assignee: "Piotr Wiśniewski", Updated: time.Now().Add(-10 * 24 * time.Hour)},
	{Key: "ERO-2010", Summary: "Add export to CSV feature for filtered issue lists", Type: "Story", Priority: "Low", Status: "To Do", Assignee: "Maria Lewandowska", Updated: time.Now().Add(-11 * 24 * time.Hour)},
	{Key: "ERO-2003", Summary: "Fix memory leak in data processor — growing heap", Type: "Bug", Priority: "High", Status: "Blocked", Assignee: "Wiktoryn Żerebecki", Updated: time.Now().Add(-12 * 24 * time.Hour)},
	{Key: "ERO-1996", Summary: "Fix race condition in task scheduler", Type: "Bug", Priority: "Highest", Status: "In Progress", Assignee: "Anna Nowak", Updated: time.Now().Add(-13 * 24 * time.Hour)},
	{Key: "ERO-1989", Summary: "Add audit log for all admin actions", Type: "Story", Priority: "High", Status: "In Review", Assignee: "Piotr Wiśniewski", Updated: time.Now().Add(-14 * 24 * time.Hour)},
	{Key: "ERO-1982", Summary: "Implement SAML 2.0 SSO for enterprise customers", Type: "Epic", Priority: "High", Status: "In Progress", Assignee: "Maria Lewandowska", Updated: time.Now().Add(-15 * 24 * time.Hour)},
	{Key: "ERO-1975", Summary: "Fix CORS configuration for staging environment", Type: "Bug", Priority: "Low", Status: "Done", Assignee: "Wiktoryn Żerebecki", Updated: time.Now().Add(-16 * 24 * time.Hour)},
	{Key: "ERO-1968", Summary: "Add webhook support for issue created/updated events", Type: "New Feature", Priority: "Medium", Status: "To Do", Assignee: "Anna Nowak", Updated: time.Now().Add(-17 * 24 * time.Hour)},
	{Key: "ERO-1961", Summary: "Migrate to PostgreSQL 16 for improved performance", Type: "Task", Priority: "Medium", Status: "Done", Assignee: "Piotr Wiśniewski", Updated: time.Now().Add(-18 * 24 * time.Hour)},
	{Key: "ERO-1954", Summary: "Add GDPR data export for personal data compliance", Type: "New Feature", Priority: "High", Status: "In Review", Assignee: "Maria Lewandowska", Updated: time.Now().Add(-19 * 24 * time.Hour)},
}

// Glyphs
func typeGlyph(t string) string {
	switch t {
	case "Bug":
		return "⊘"
	case "Task":
		return "✓"
	case "Story":
		return "★"
	case "Epic":
		return "⬢"
	case "Sub-task", "Subtask":
		return "↳"
	case "Improvement":
		return "⚒"
	case "New Feature":
		return "✦"
	default:
		return "?"
	}
}

func priGlyph(p string) string {
	switch p {
	case "Highest", "Critical", "Blocker":
		return "⇈"
	case "High", "Major":
		return "▲"
	case "Medium":
		return "─"
	case "Low", "Minor":
		return "▼"
	case "Lowest", "Trivial":
		return "⇊"
	default:
		return "─"
	}
}

func statusGlyph(s string) string {
	switch s {
	case "Done", "Closed", "Resolved":
		return "✓"
	case "Blocked", "On Hold":
		return "✕"
	case "In Review", "Testing", "QA":
		return "◑"
	case "In Progress":
		return "◐"
	case "Ready For Build", "Ready For Dev", "Selected For Development":
		return "▷"
	default:
		return "○"
	}
}
