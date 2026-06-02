package jira

import (
	"fmt"
	"strings"
	"time"
)

type MockClient struct {
	issues  []Issue
	filters []SavedFilter
}

var mockUsers = []JiraUser{
	{AccountID: "u1", DisplayName: "Alice Johnson", EmailAddress: "alice@example.com"},
	{AccountID: "u2", DisplayName: "Bob Smith", EmailAddress: "bob@example.com"},
	{AccountID: "u3", DisplayName: "Carol White", EmailAddress: "carol@example.com"},
	{AccountID: "u4", DisplayName: "Dave Brown", EmailAddress: "dave@example.com"},
}

var mockProjects = []Project{
	{ID: "1", Key: "PROJ", Name: "Main Project"},
	{ID: "2", Key: "INFRA", Name: "Infrastructure"},
	{ID: "3", Key: "DEV", Name: "Development"},
	{ID: "4", Key: "QA", Name: "Quality Assurance"},
}

var mockPriorities = []Priority{
	{ID: "1", Name: "Highest"},
	{ID: "2", Name: "High"},
	{ID: "3", Name: "Medium"},
	{ID: "4", Name: "Low"},
	{ID: "5", Name: "Lowest"},
}

var mockStatuses = []Status{
	{ID: "1", Name: "To Do", StatusCategory: StatusCategory{Key: "new", Name: "To Do"}},
	{ID: "2", Name: "In Progress", StatusCategory: StatusCategory{Key: "indeterminate", Name: "In Progress"}},
	{ID: "3", Name: "In Review", StatusCategory: StatusCategory{Key: "indeterminate", Name: "In Progress"}},
	{ID: "4", Name: "Done", StatusCategory: StatusCategory{Key: "done", Name: "Done"}},
	{ID: "5", Name: "Blocked", StatusCategory: StatusCategory{Key: "indeterminate", Name: "In Progress"}},
}

var mockTypes = []IssueType{
	{ID: "1", Name: "Bug"},
	{ID: "2", Name: "Task"},
	{ID: "3", Name: "Story"},
	{ID: "4", Name: "Epic"},
	{ID: "5", Name: "Sub-task", Subtask: true},
	{ID: "6", Name: "Improvement"},
	{ID: "7", Name: "New Feature"},
}

type issueTemplate struct {
	summary     string
	issueType   string
	priority    string
	status      string
	assigneeIdx int
	reporterIdx int
	labels      []string
	sprint      string
	description string
}

var issueTemplates = []issueTemplate{
	{"Fix login timeout on mobile devices", "Bug", "High", "In Progress", 0, 1, []string{"mobile", "auth"}, "Sprint 42", "Login sessions expire too quickly on mobile. Token refresh doesn't account for background state on iOS."},
	{"Critical payment gateway error", "Bug", "Highest", "To Do", 1, 0, []string{"payment", "critical"}, "Sprint 42", "Payment processing fails intermittently with error code 500. Affects 2% of transactions."},
	{"Add dark mode to settings", "Story", "Medium", "In Review", 0, 2, []string{"ui", "settings"}, "Sprint 42", "Users have requested dark mode. Implement theme switching in settings panel."},
	{"Database connection pool exhaustion", "Bug", "Highest", "In Progress", 2, 1, []string{"database", "performance"}, "Sprint 41", "Connection pool runs out under high load. Need to tune pool parameters or add connection recycling."},
	{"Implement OAuth 2.0 login", "New Feature", "High", "To Do", 3, 0, []string{"auth", "oauth"}, "Sprint 43", "Add OAuth 2.0 support for enterprise customers. Must support Google, Microsoft, and Okta providers."},
	{"Update dependencies to latest versions", "Task", "Low", "Done", 1, 2, []string{"maintenance"}, "Sprint 40", "Quarterly dependency update. Run npm audit and resolve any security vulnerabilities."},
	{"Performance regression in search API", "Bug", "High", "In Progress", 0, 3, []string{"performance", "api"}, "Sprint 42", "Search API response time degraded by 40% after last deployment. Profile and fix."},
	{"Add pagination to user list", "Improvement", "Medium", "Done", 2, 1, []string{"ui", "pagination"}, "Sprint 41", "User list page loads all users at once, causing slow page loads for large organizations."},
	{"Fix broken image uploads on Firefox", "Bug", "Medium", "To Do", 3, 0, []string{"firefox", "upload"}, "Sprint 42", "Image uploads fail silently on Firefox 118+. Works on Chrome and Safari."},
	{"Implement real-time notifications", "New Feature", "High", "In Progress", 1, 2, []string{"notifications", "websocket"}, "Sprint 43", "Add WebSocket-based real-time notifications for new assignments and mentions."},
	{"Refactor authentication middleware", "Task", "Medium", "In Review", 0, 1, []string{"refactoring", "auth"}, "Sprint 42", "Current auth middleware is hard to test. Refactor to use dependency injection."},
	{"Add export to CSV feature", "Story", "Low", "To Do", 2, 3, []string{"export", "csv"}, "Sprint 43", "Allow users to export filtered issue lists to CSV format for reporting."},
	{"Fix memory leak in data processor", "Bug", "High", "In Progress", 3, 0, []string{"memory", "performance"}, "Sprint 42", "Data processing service shows growing memory usage over time. Heap dump shows unclosed streams."},
	{"Improve error messages for API consumers", "Improvement", "Medium", "Done", 1, 2, []string{"api", "dx"}, "Sprint 41", "API error messages are cryptic. Add field-level validation errors and more descriptive messages."},
	{"Set up CI/CD pipeline for mobile app", "Task", "High", "In Progress", 0, 3, []string{"ci", "mobile", "devops"}, "Sprint 42", "Configure GitHub Actions to build, test, and deploy iOS and Android apps automatically."},
	{"Add two-factor authentication", "New Feature", "High", "To Do", 2, 1, []string{"security", "auth"}, "Sprint 43", "Implement TOTP-based 2FA. Support authenticator apps and SMS fallback."},
	{"Optimize database queries in reporting module", "Improvement", "Medium", "In Review", 3, 0, []string{"database", "performance", "reporting"}, "Sprint 42", "Reporting queries take 5-10 seconds. Add indexes and optimize JOINs."},
	{"Fix CORS configuration for staging environment", "Bug", "Low", "Done", 1, 2, []string{"cors", "staging"}, "Sprint 41", "CORS headers missing on OPTIONS requests in staging. Frontend team is blocked."},
	{"Implement data retention policies", "Task", "Medium", "To Do", 0, 3, []string{"compliance", "data"}, "Sprint 43", "GDPR requires data retention limits. Implement automatic purging of old user data."},
	{"Add keyboard shortcuts to issue list", "Story", "Low", "Done", 2, 1, []string{"ui", "accessibility"}, "Sprint 40", "Power users have requested keyboard navigation. Add arrow keys, enter, and common shortcuts."},
	{"Fix timezone handling in calendar view", "Bug", "Medium", "In Progress", 3, 0, []string{"timezone", "calendar"}, "Sprint 42", "Calendar shows incorrect times for users in non-UTC timezones. Parse and display in local time."},
	{"Implement file attachments for issues", "New Feature", "Medium", "To Do", 1, 2, []string{"attachments", "files"}, "Sprint 43", "Allow attaching files, screenshots, and documents to issues. Max 10MB per file."},
	{"Add audit log for admin actions", "Story", "High", "In Review", 0, 3, []string{"audit", "security", "admin"}, "Sprint 42", "Track all admin actions for compliance. Store who changed what and when."},
	{"Fix race condition in task scheduler", "Bug", "Highest", "In Progress", 2, 1, []string{"concurrency", "scheduler"}, "Sprint 42", "Scheduler occasionally runs tasks twice. Investigate and fix distributed locking."},
	{"Update API documentation", "Task", "Low", "Done", 3, 0, []string{"documentation", "api"}, "Sprint 41", "API docs are outdated. Update OpenAPI spec and regenerate client libraries."},
	{"Implement search autocomplete", "Improvement", "Medium", "To Do", 1, 2, []string{"search", "ux"}, "Sprint 43", "Add autocomplete suggestions to search box to help users find issues faster."},
	{"Fix email notification delivery delays", "Bug", "High", "In Progress", 0, 3, []string{"email", "notifications"}, "Sprint 42", "Email notifications delayed by up to 30 minutes. Check mail queue and SMTP configuration."},
	{"Add user role management", "New Feature", "High", "In Review", 2, 1, []string{"rbac", "admin"}, "Sprint 43", "Implement granular role-based access control. Support custom roles with per-resource permissions."},
	{"Refactor frontend build system", "Task", "Medium", "Done", 3, 0, []string{"build", "frontend", "webpack"}, "Sprint 41", "Switch from Webpack 4 to Vite for faster builds and better developer experience."},
	{"Fix broken links in help documentation", "Bug", "Low", "To Do", 1, 2, []string{"documentation"}, "Sprint 43", "Several help articles have broken links after recent site restructure. Audit and fix."},
	{"Implement rate limiting for API", "Story", "High", "In Progress", 0, 3, []string{"api", "security", "rate-limiting"}, "Sprint 42", "Add per-user and per-IP rate limiting to prevent abuse. Target: 100 req/min for authenticated users."},
	{"Add integration tests for payment flow", "Task", "High", "To Do", 2, 1, []string{"testing", "payment"}, "Sprint 42", "Critical payment paths lack integration tests. Add end-to-end tests using test payment processor."},
	{"Fix responsive layout on tablet screen sizes", "Bug", "Medium", "In Review", 3, 0, []string{"ui", "responsive", "tablet"}, "Sprint 42", "Layout breaks on iPad-sized screens (768-1024px width). Fix grid breakpoints."},
	{"Implement custom dashboard widgets", "Story", "Medium", "To Do", 1, 2, []string{"dashboard", "widgets"}, "Sprint 43", "Allow users to customize their dashboard with draggable, configurable widgets."},
	{"Set up database replication", "Task", "High", "Done", 0, 3, []string{"database", "ha", "devops"}, "Sprint 41", "Configure PostgreSQL streaming replication for high availability. Primary in us-east-1, replica in eu-west-1."},
	{"Fix XSS vulnerability in comment editor", "Bug", "Highest", "Done", 2, 1, []string{"security", "xss"}, "Sprint 41", "Comment editor allows script injection. Sanitize HTML input with DOMPurify."},
	{"Add dark mode support for email templates", "Improvement", "Low", "To Do", 3, 0, []string{"email", "dark-mode"}, "Sprint 43", "Email clients that support dark mode should show appropriate colors. Add @media prefers-color-scheme."},
	{"Implement SSO for enterprise customers", "Epic", "High", "In Progress", 1, 2, []string{"sso", "enterprise", "auth"}, "Sprint 42", "SAML 2.0 SSO integration for enterprise customers. Support Azure AD, Okta, and OneLogin."},
	{"Fix incorrect totals in analytics dashboard", "Bug", "High", "In Review", 0, 3, []string{"analytics", "bug", "data"}, "Sprint 42", "Analytics dashboard shows incorrect totals for multi-tenant queries. Investigate query isolation."},
	{"Add webhook support for issue events", "New Feature", "Medium", "To Do", 2, 1, []string{"webhooks", "integration"}, "Sprint 43", "Allow customers to register webhooks for issue created/updated/deleted events."},
	{"Optimize image loading with lazy loading", "Improvement", "Low", "Done", 3, 0, []string{"performance", "images", "ui"}, "Sprint 41", "Page loads are slow due to eager image loading. Implement IntersectionObserver-based lazy loading."},
	{"Fix session handling for concurrent logins", "Bug", "Medium", "In Progress", 1, 2, []string{"auth", "session"}, "Sprint 42", "Users logged in from multiple devices get unexpected logouts. Fix session store concurrency."},
	{"Implement activity feed", "Story", "Medium", "To Do", 0, 3, []string{"activity", "social"}, "Sprint 43", "Show a chronological feed of recent changes made by team members."},
	{"Add Slack integration", "New Feature", "Medium", "In Review", 2, 1, []string{"slack", "integration", "notifications"}, "Sprint 43", "Send notifications to Slack channels when issues are assigned, transitioned, or commented on."},
	{"Fix CSV import for special characters", "Bug", "Low", "Done", 3, 0, []string{"import", "csv", "encoding"}, "Sprint 41", "CSV imports fail when data contains non-ASCII characters. Fix encoding detection and UTF-8 handling."},
	{"Implement issue linking", "Story", "Medium", "In Progress", 1, 2, []string{"issues", "linking"}, "Sprint 42", "Allow linking issues as 'blocks', 'is blocked by', 'relates to', 'duplicates'."},
	{"Add mobile push notifications", "New Feature", "High", "To Do", 0, 3, []string{"mobile", "push", "notifications"}, "Sprint 43", "Implement push notifications for iOS and Android using FCM/APNs."},
	{"Fix incorrect status transitions", "Bug", "Medium", "In Review", 2, 1, []string{"workflow", "status"}, "Sprint 42", "Some status transitions bypass required fields validation. Fix workflow validation logic."},
	{"Implement issue templates", "Story", "Low", "To Do", 3, 0, []string{"templates", "productivity"}, "Sprint 43", "Allow project admins to create issue templates with pre-filled fields and checklists."},
	{"Migrate to PostgreSQL 16", "Task", "Medium", "Done", 1, 2, []string{"database", "migration", "postgresql"}, "Sprint 41", "Upgrade from PostgreSQL 14 to 16 for improved performance and new features."},
	{"Fix OAuth token refresh race condition", "Bug", "High", "In Progress", 0, 3, []string{"oauth", "auth", "concurrency"}, "Sprint 42", "Concurrent API requests can trigger multiple token refreshes simultaneously. Implement mutex."},
	{"Add in-app onboarding tour", "Story", "Low", "To Do", 2, 1, []string{"onboarding", "ux"}, "Sprint 43", "Guide new users through key features with an interactive tour. Use React Joyride."},
	{"Implement comment threading", "Improvement", "Medium", "In Progress", 3, 0, []string{"comments", "ui"}, "Sprint 42", "Allow replying to specific comments to create threaded discussions."},
	{"Fix pagination in project list API", "Bug", "Low", "Done", 1, 2, []string{"api", "pagination"}, "Sprint 41", "Project list API ignores maxResults parameter. Returns all projects regardless of limit."},
	{"Add GDPR data export", "New Feature", "High", "In Review", 0, 3, []string{"gdpr", "compliance", "export"}, "Sprint 43", "Allow users to export all their personal data in machine-readable format (JSON/CSV)."},
	{"Improve search indexing performance", "Improvement", "High", "In Progress", 2, 1, []string{"search", "elasticsearch", "performance"}, "Sprint 42", "Full-text search indexing falls behind during peak hours. Optimize indexer and add backpressure."},
	{"Fix broken API versioning", "Bug", "Medium", "Done", 3, 0, []string{"api", "versioning"}, "Sprint 41", "API v1 endpoints return v2 format when Accept-Version header is missing. Fix version negotiation."},
	{"Implement team workload view", "Story", "Medium", "To Do", 1, 2, []string{"dashboard", "workload", "team"}, "Sprint 43", "Show assigned issues per team member with capacity indicators and sprint burndown."},
	{"Add IP allowlist for admin panel", "New Feature", "High", "To Do", 0, 3, []string{"security", "admin", "access-control"}, "Sprint 43", "Restrict admin panel access to whitelisted IP ranges. Support CIDR notation."},
	{"Fix broken link preview in rich text editor", "Bug", "Low", "In Progress", 2, 1, []string{"editor", "links"}, "Sprint 42", "Link previews don't render when URLs contain query parameters with special characters."},
	{"Implement project archiving", "Story", "Low", "To Do", 3, 0, []string{"projects", "archive"}, "Sprint 43", "Allow archiving completed projects. Archived projects hidden by default but searchable."},
	{"Add OpenTelemetry tracing", "Task", "Medium", "In Review", 1, 2, []string{"observability", "tracing", "otel"}, "Sprint 42", "Instrument backend services with OpenTelemetry for distributed tracing. Export to Jaeger."},
	{"Fix incorrect permissions for shared filters", "Bug", "Medium", "Done", 0, 3, []string{"permissions", "filters"}, "Sprint 41", "Shared filters visible to users without read access to referenced projects."},
	{"Implement batch issue updates", "New Feature", "Medium", "To Do", 2, 1, []string{"bulk", "productivity"}, "Sprint 43", "Allow updating multiple issues at once (status, assignee, priority, labels)."},
}

func NewMockClient() *MockClient {
	c := &MockClient{}
	c.buildIssues()
	c.buildFilters()
	return c
}

func (c *MockClient) buildIssues() {
	base := time.Date(2025, 5, 1, 9, 0, 0, 0, time.UTC)
	projects := mockProjects

	for i, tmpl := range issueTemplates {
		proj := projects[i%len(projects)]
		issueNum := i + 1
		key := fmt.Sprintf("%s-%d", proj.Key, issueNum)

		assignee := &mockUsers[tmpl.assigneeIdx%len(mockUsers)]
		reporter := &mockUsers[tmpl.reporterIdx%len(mockUsers)]

		created := base.Add(time.Duration(i*-3) * time.Hour * 24)
		updated := created.Add(time.Duration(i%5+1) * time.Hour * 12)

		issue := Issue{
			ID:          fmt.Sprintf("%d", 1000+i),
			Key:         key,
			Summary:     tmpl.summary,
			Description: adfFromText(tmpl.description),
			Priority:    priorityByName(tmpl.priority),
			Status:      statusByName(tmpl.status),
			IssueType:   typeByName(tmpl.issueType),
			Project:     proj,
			Assignee:    assignee,
			Reporter:    reporter,
			Labels:      tmpl.labels,
			Sprint:      tmpl.sprint,
			Comments:    buildComments(i, assignee, reporter, updated),
			Created:     created,
			Updated:     updated,
		}
		c.issues = append(c.issues, issue)
	}
}

func buildComments(seed int, assignee, reporter *JiraUser, base time.Time) []Comment {
	commentTexts := [][]string{
		{"Reproduced on latest build. Working on a fix.", "Thanks, assigned to sprint."},
		{"Investigation started. Root cause identified.", "Please prioritize this."},
		{"PR #" + fmt.Sprintf("%d", 800+seed) + " opened for review.", "LGTM, merging."},
		{"This is blocked by another issue.", "Unblocked, resuming work."},
		{},
	}
	texts := commentTexts[seed%len(commentTexts)]

	var comments []Comment
	for i, text := range texts {
		author := reporter
		if i%2 == 0 {
			author = assignee
		}
		comments = append(comments, Comment{
			ID:      fmt.Sprintf("c%d%d", seed, i),
			Author:  *author,
			Body:    text,
			Created: base.Add(time.Duration(i+1) * time.Hour * 24),
			Updated: base.Add(time.Duration(i+1) * time.Hour * 24),
		})
	}
	return comments
}

func priorityByName(name string) Priority {
	for _, p := range mockPriorities {
		if p.Name == name {
			return p
		}
	}
	return Priority{ID: "3", Name: "Medium"}
}

func statusByName(name string) Status {
	for _, s := range mockStatuses {
		if s.Name == name {
			return s
		}
	}
	return mockStatuses[0]
}

func typeByName(name string) IssueType {
	for _, t := range mockTypes {
		if t.Name == name {
			return t
		}
	}
	return IssueType{ID: "2", Name: "Task"}
}

func adfFromText(text string) string {
	escaped, _ := encodeJSON(text)
	return `{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":` + escaped + `}]}]}`
}

func encodeJSON(s string) (string, error) {
	import_json_encoder := func(s string) string {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		s = strings.ReplaceAll(s, "\n", `\n`)
		s = strings.ReplaceAll(s, "\r", `\r`)
		s = strings.ReplaceAll(s, "\t", `\t`)
		return `"` + s + `"`
	}
	return import_json_encoder(s), nil
}

func (c *MockClient) buildFilters() {
	c.filters = []SavedFilter{
		{ID: "f1", Name: "My open issues", JQL: "assignee = currentUser() AND statusCategory != Done"},
		{ID: "f2", Name: "Critical bugs", JQL: "type = Bug AND priority in (Highest, High) AND statusCategory != Done"},
		{ID: "f3", Name: "Sprint 42", JQL: "sprint = \"Sprint 42\" AND statusCategory != Done"},
	}
}

func (c *MockClient) GetCurrentUser() (*JiraUser, error) {
	u := mockUsers[0]
	return &u, nil
}

func (c *MockClient) SearchIssues(jql string, maxResults int) ([]Issue, int, error) {
	filtered := c.filterIssues(jql)
	total := len(filtered)
	if maxResults > 0 && len(filtered) > maxResults {
		filtered = filtered[:maxResults]
	}
	return filtered, total, nil
}

func (c *MockClient) filterIssues(jql string) []Issue {
	if jql == "" {
		return c.issues
	}
	lower := strings.ToLower(jql)

	var result []Issue
	for _, issue := range c.issues {
		if matchJQL(issue, lower) {
			result = append(result, issue)
		}
	}
	return result
}

func matchJQL(issue Issue, jql string) bool {
	if strings.Contains(jql, "project =") || strings.Contains(jql, "project=") {
		for _, proj := range mockProjects {
			needle := strings.ToLower(proj.Key)
			if strings.Contains(jql, "project = "+needle) || strings.Contains(jql, "project="+needle) ||
				strings.Contains(jql, "project = \""+needle+"\"") {
				if strings.ToLower(issue.Project.Key) != needle {
					return false
				}
				break
			}
		}
	}

	if strings.Contains(jql, "assignee = currentuser()") {
		if issue.Assignee == nil || issue.Assignee.AccountID != mockUsers[0].AccountID {
			return false
		}
	}

	if strings.Contains(jql, "issuetype =") || strings.Contains(jql, "type =") {
		for _, t := range mockTypes {
			needle := strings.ToLower(t.Name)
			if strings.Contains(jql, "issuetype = "+needle) || strings.Contains(jql, "type = "+needle) ||
				strings.Contains(jql, "issuetype = \""+needle+"\"") || strings.Contains(jql, "type = \""+needle+"\"") {
				if strings.ToLower(issue.IssueType.Name) != needle {
					return false
				}
				break
			}
		}
	}

	if strings.Contains(jql, "status =") {
		for _, s := range mockStatuses {
			needle := strings.ToLower(s.Name)
			if strings.Contains(jql, "status = \""+needle+"\"") || strings.Contains(jql, "status = "+needle) {
				if strings.ToLower(issue.Status.Name) != needle {
					return false
				}
				break
			}
		}
	}

	if strings.Contains(jql, "statuscategory != done") {
		if issue.Status.StatusCategory.Key == "done" {
			return false
		}
	}

	return true
}

func (c *MockClient) GetIssue(key string) (*Issue, error) {
	for _, issue := range c.issues {
		if issue.Key == key {
			cp := issue
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("issue %s not found", key)
}

func (c *MockClient) GetProjects() ([]Project, error) {
	return mockProjects, nil
}

func (c *MockClient) GetPriorities() ([]Priority, error) {
	return mockPriorities, nil
}

func (c *MockClient) GetStatuses() ([]Status, error) {
	return mockStatuses, nil
}

func (c *MockClient) GetIssueTypes() ([]IssueType, error) {
	return mockTypes, nil
}

func (c *MockClient) GetSavedFilters() ([]SavedFilter, error) {
	return c.filters, nil
}

func (c *MockClient) GetTransitions(key string) ([]Transition, error) {
	return []Transition{
		{ID: "t1", Name: "Start Progress", To: statusByName("In Progress")},
		{ID: "t2", Name: "Send for Review", To: statusByName("In Review")},
		{ID: "t3", Name: "Done", To: statusByName("Done")},
		{ID: "t4", Name: "Reopen", To: statusByName("To Do")},
	}, nil
}

func (c *MockClient) DoTransition(key, transitionID string) error {
	transitions, _ := c.GetTransitions(key)
	for _, t := range transitions {
		if t.ID == transitionID {
			for i := range c.issues {
				if c.issues[i].Key == key {
					c.issues[i].Status = t.To
					c.issues[i].Updated = time.Now()
					return nil
				}
			}
		}
	}
	return fmt.Errorf("transition %s not found", transitionID)
}

func (c *MockClient) SearchAssignableUsers(query, _ string) ([]JiraUser, error) {
	lower := strings.ToLower(query)
	var result []JiraUser
	for _, u := range mockUsers {
		if query == "" || strings.Contains(strings.ToLower(u.DisplayName), lower) ||
			strings.Contains(strings.ToLower(u.EmailAddress), lower) {
			result = append(result, u)
		}
	}
	return result, nil
}

func (c *MockClient) UpdatePriority(key, priorityName string) error {
	for i := range c.issues {
		if c.issues[i].Key == key {
			c.issues[i].Priority = priorityByName(priorityName)
			c.issues[i].Updated = time.Now()
			return nil
		}
	}
	return fmt.Errorf("issue %s not found", key)
}

func (c *MockClient) UpdateAssignee(key, accountID string) error {
	for i := range c.issues {
		if c.issues[i].Key == key {
			for _, u := range mockUsers {
				if u.AccountID == accountID {
					cp := u
					c.issues[i].Assignee = &cp
					c.issues[i].Updated = time.Now()
					return nil
				}
			}
			c.issues[i].Assignee = nil
			c.issues[i].Updated = time.Now()
			return nil
		}
	}
	return fmt.Errorf("issue %s not found", key)
}

func (c *MockClient) UpdateDescription(key, description string) error {
	for i := range c.issues {
		if c.issues[i].Key == key {
			c.issues[i].Description = adfFromText(description)
			c.issues[i].Updated = time.Now()
			return nil
		}
	}
	return fmt.Errorf("issue %s not found", key)
}

func (c *MockClient) AddComment(key, body string) error {
	for i := range c.issues {
		if c.issues[i].Key == key {
			user := mockUsers[0]
			c.issues[i].Comments = append(c.issues[i].Comments, Comment{
				ID:      fmt.Sprintf("c_new_%d", len(c.issues[i].Comments)),
				Author:  user,
				Body:    body,
				Created: time.Now(),
				Updated: time.Now(),
			})
			c.issues[i].Updated = time.Now()
			return nil
		}
	}
	return fmt.Errorf("issue %s not found", key)
}

func (c *MockClient) SaveFilter(name, description, jql string) (*SavedFilter, error) {
	f := &SavedFilter{
		ID:   fmt.Sprintf("f%d", len(c.filters)+1),
		Name: name,
		JQL:  jql,
	}
	c.filters = append(c.filters, *f)
	return f, nil
}
