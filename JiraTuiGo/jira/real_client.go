package jira

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// debugLog is a file logger active when JIRATUI_DEBUG=1 is set.
var debugLog *log.Logger

func init() {
	if os.Getenv("JIRATUI_DEBUG") != "1" {
		return
	}
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "jiratui-debug.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	debugLog = log.New(f, "", log.LstdFlags)
	debugLog.Printf("=== JiraTUI debug log opened: %s ===", path)
}

func dbg(format string, args ...any) {
	if debugLog != nil {
		debugLog.Printf(format, args...)
	}
}

// Dbg is the exported version for use by other packages.
func Dbg(format string, args ...any) { dbg(format, args...) }

// parseSprint extracts the sprint name from the raw sprint field.
// Jira returns it as a single object {"name":"Sprint 9",...} or an array.
// We prefer the active sprint, then future, then the last entry.
func parseSprint(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	type sprintObj struct {
		Name  string `json:"name"`
		State string `json:"state"`
	}

	// Try array form first.
	var arr []sprintObj
	if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
		var active, future *sprintObj
		for i := range arr {
			switch strings.ToLower(arr[i].State) {
			case "active":
				if active == nil {
					active = &arr[i]
				}
			case "future":
				if future == nil {
					future = &arr[i]
				}
			}
		}
		pick := active
		if pick == nil {
			pick = future
		}
		if pick == nil {
			pick = &arr[len(arr)-1]
		}
		return pick.Name
	}

	// Try single-object form.
	var obj sprintObj
	if json.Unmarshal(raw, &obj) == nil && obj.Name != "" {
		return obj.Name
	}
	return ""
}

type RealClient struct {
	baseURL       string
	authHeader    string
	http          *http.Client
	sprintFieldID string // cached once we find it among customfield_* keys
}

func NewRealClient(baseURL, email, token string) *RealClient {
	creds := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
	return &RealClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		authHeader: "Basic " + creds,
		http:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *RealClient) get(path string, v any) error {
	dbg("GET %s", c.baseURL+path)
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		dbg("GET request build error: %v", err)
		return err
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		dbg("GET http error: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	dbg("GET %s → %d (%d bytes)", path, resp.StatusCode, len(body))
	if resp.StatusCode >= 400 {
		dbg("GET error body: %.500s", string(body))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, v); err != nil {
		dbg("GET json decode error: %v | body[:200]: %.200s", err, string(body))
		return err
	}
	return nil
}

func (c *RealClient) post(path string, body, v any) error {
	return c.doJSON("POST", path, body, v)
}

func (c *RealClient) put(path string, body any) error {
	return c.doJSON("PUT", path, body, nil)
}

func (c *RealClient) doJSON(method, path string, body, v any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	dbg("%s %s", method, c.baseURL+path)
	resp, err := c.http.Do(req)
	if err != nil {
		dbg("%s http error: %v", method, err)
		return err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	dbg("%s %s → %d (%d bytes)", method, path, resp.StatusCode, len(b))
	if resp.StatusCode >= 400 {
		dbg("%s error body: %.500s", method, string(b))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	if v != nil {
		if err := json.Unmarshal(b, v); err != nil {
			dbg("%s json decode error: %v | body[:200]: %.200s", method, err, string(b))
			return err
		}
	}
	return nil
}

func (c *RealClient) GetCurrentUser() (*JiraUser, error) {
	var u JiraUser
	if err := c.get("/rest/api/3/myself", &u); err != nil {
		return nil, err
	}
	return &u, nil
}

type searchResponse struct {
	Issues []rawIssue `json:"issues"`
	Total  int        `json:"total"`
}

type rawIssue struct {
	ID     string          `json:"id"`
	Key    string          `json:"key"`
	Fields json.RawMessage `json:"fields"` // kept raw so we can scan customfield_* for sprint
}

// jiraTime handles Jira Cloud's date format "2006-01-02T15:04:05.000+0000"
// where the timezone offset has no colon — unlike RFC3339 which requires one.
type jiraTime struct{ time.Time }

func (jt *jiraTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		jt.Time = time.Time{}
		return nil
	}
	// Try formats in order. Go's -0700 token matches ±HHMM (no colon).
	for _, f := range []string{
		"2006-01-02T15:04:05.999-0700", // Jira Cloud: milliseconds, no colon in tz
		"2006-01-02T15:04:05-0700",     // no milliseconds, no colon
		time.RFC3339Nano,               // standard with colon + nanoseconds
		time.RFC3339,                   // standard with colon
	} {
		if t, err := time.Parse(f, s); err == nil {
			jt.Time = t
			return nil
		}
	}
	jt.Time = time.Time{} // zero value — don't fail the whole decode
	return nil
}

type rawFields struct {
	Summary     string          `json:"summary"`
	Description json.RawMessage `json:"description"`
	Priority    *Priority       `json:"priority"`
	Status      *Status         `json:"status"`
	IssueType   *IssueType      `json:"issuetype"`
	Project     *Project        `json:"project"`
	Assignee    *JiraUser       `json:"assignee"`
	Reporter    *JiraUser       `json:"reporter"`
	Labels      []string        `json:"labels"`
	Created     jiraTime        `json:"created"`
	Updated     jiraTime        `json:"updated"`
	Comment     *struct {
		Comments []rawComment `json:"comments"`
	} `json:"comment"`
	Sprint json.RawMessage `json:"sprint"`
}

type rawComment struct {
	ID      string          `json:"id"`
	Author  JiraUser        `json:"author"`
	Body    json.RawMessage `json:"body"`
	Created jiraTime        `json:"created"`
	Updated jiraTime        `json:"updated"`
}

type searchPayload struct {
	Jql        string   `json:"jql"`
	MaxResults int      `json:"maxResults"`
	Fields     []string `json:"fields"`
}

func (c *RealClient) SearchIssues(jql string, maxResults int) ([]Issue, int, error) {
	payload := searchPayload{
		Jql:        jql,
		MaxResults: maxResults,
		// "*navigable" returns all navigable fields (including custom sprint field);
		// description and comment are not navigable so we add them explicitly.
		Fields: []string{"*navigable", "description", "comment"},
	}

	// Try the newer /rest/api/3/search/jql endpoint first (Jira Cloud 2024+),
	// then fall back to the classic /rest/api/3/issue/search.
	var lastErr error
	for _, path := range []string{
		"/rest/api/3/search/jql",
		"/rest/api/3/issue/search",
	} {
		var result searchResponse
		err := c.post(path, payload, &result)
		if err == nil {
			issues := make([]Issue, len(result.Issues))
			for i, ri := range result.Issues {
				issues[i] = c.convertIssue(ri)
			}
			dbg("SearchIssues %s → %d issues", path, len(issues))
			return issues, result.Total, nil
		}
		dbg("SearchIssues %s error: %v", path, err)
		// Only try the next endpoint on 404/405 — other errors are real failures.
		if !strings.Contains(err.Error(), "HTTP 404") && !strings.Contains(err.Error(), "HTTP 405") {
			return nil, 0, err
		}
		lastErr = err
	}
	return nil, 0, lastErr
}

func (c *RealClient) convertIssue(ri rawIssue) Issue {
	var f rawFields
	_ = json.Unmarshal(ri.Fields, &f)

	issue := Issue{
		ID:      ri.ID,
		Key:     ri.Key,
		Summary: f.Summary,
		Labels:  f.Labels,
		Created: f.Created.Time,
		Updated: f.Updated.Time,
	}
	if f.Description != nil {
		issue.Description = RenderADF(string(f.Description))
	}
	if f.Priority != nil {
		issue.Priority = *f.Priority
	}
	if f.Status != nil {
		issue.Status = *f.Status
	}
	if f.IssueType != nil {
		issue.IssueType = *f.IssueType
	}
	if f.Project != nil {
		issue.Project = *f.Project
	}
	issue.Assignee = f.Assignee
	issue.Reporter = f.Reporter

	// Sprint lives in a customfield_* key that varies between Jira instances.
	// Scan all keys the first time, then use the cached ID — same approach as C#.
	issue.Sprint = c.extractSprint(ri.Fields)

	if f.Comment != nil {
		for _, rc := range f.Comment.Comments {
			issue.Comments = append(issue.Comments, Comment{
				ID:      rc.ID,
				Author:  rc.Author,
				Body:    RenderADF(string(rc.Body)),
				Created: rc.Created.Time,
				Updated: rc.Updated.Time,
			})
		}
	}
	return issue
}

// extractSprint scans all customfield_* keys for an array (or object) that
// looks like a Jira sprint, caches the field ID, and returns the sprint name.
func (c *RealClient) extractSprint(fieldsRaw json.RawMessage) string {
	var allFields map[string]json.RawMessage
	if err := json.Unmarshal(fieldsRaw, &allFields); err != nil {
		return ""
	}

	// Fast path: use cached field ID.
	if c.sprintFieldID != "" {
		if raw, ok := allFields[c.sprintFieldID]; ok {
			if name := parseSprint(raw); name != "" {
				return name
			}
		}
	}

	// Slow path: scan all customfield_* keys.
	for key, raw := range allFields {
		if !strings.HasPrefix(key, "customfield_") {
			continue
		}
		name := parseSprint(raw)
		if name == "" {
			continue
		}
		// Verify it looks like a sprint object (has state or boardId).
		type sprintCheck struct {
			State   string `json:"state"`
			BoardID any    `json:"boardId"`
		}
		var arr []sprintCheck
		var obj sprintCheck
		if json.Unmarshal(raw, &arr) == nil && len(arr) > 0 {
			if arr[0].State != "" || arr[0].BoardID != nil {
				c.sprintFieldID = key
				return name
			}
		} else if json.Unmarshal(raw, &obj) == nil {
			if obj.State != "" || obj.BoardID != nil {
				c.sprintFieldID = key
				return name
			}
		}
	}
	return ""
}

func (c *RealClient) GetIssue(key string) (*Issue, error) {
	var ri rawIssue
	if err := c.get("/rest/api/3/issue/"+key+"?fields=*all", &ri); err != nil {
		return nil, err
	}
	issue := c.convertIssue(ri)
	return &issue, nil
}

func (c *RealClient) GetProjects() ([]Project, error) {
	var result struct {
		Values []Project `json:"values"`
	}
	if err := c.get("/rest/api/3/project/search?maxResults=200", &result); err != nil {
		return nil, err
	}
	return result.Values, nil
}

func (c *RealClient) GetPriorities() ([]Priority, error) {
	var priorities []Priority
	return priorities, c.get("/rest/api/3/priority", &priorities)
}

func (c *RealClient) GetStatuses() ([]Status, error) {
	var statuses []Status
	return statuses, c.get("/rest/api/3/status", &statuses)
}

func (c *RealClient) GetIssueTypes() ([]IssueType, error) {
	var types []IssueType
	return types, c.get("/rest/api/3/issuetype", &types)
}

func (c *RealClient) GetSavedFilters() ([]SavedFilter, error) {
	var filters []SavedFilter
	return filters, c.get("/rest/api/3/filter/my?expand=jql", &filters)
}

func (c *RealClient) GetTransitions(key string) ([]Transition, error) {
	var result struct {
		Transitions []Transition `json:"transitions"`
	}
	if err := c.get("/rest/api/3/issue/"+key+"/transitions", &result); err != nil {
		return nil, err
	}
	return result.Transitions, nil
}

func (c *RealClient) DoTransition(key, transitionID string) error {
	return c.post("/rest/api/3/issue/"+key+"/transitions", map[string]any{
		"transition": map[string]string{"id": transitionID},
	}, nil)
}

func (c *RealClient) SearchAssignableUsers(query, projectKey string) ([]JiraUser, error) {
	path := "/rest/api/3/user/assignable/search?project=" + url.QueryEscape(projectKey)
	if query != "" {
		path += "&query=" + url.QueryEscape(query)
	}
	var users []JiraUser
	return users, c.get(path, &users)
}

func (c *RealClient) UpdatePriority(key, priorityName string) error {
	return c.put("/rest/api/3/issue/"+key, map[string]any{
		"fields": map[string]any{
			"priority": map[string]string{"name": priorityName},
		},
	})
}

func (c *RealClient) UpdateAssignee(key, accountID string) error {
	assignee := map[string]any{"accountId": accountID}
	if accountID == "" {
		assignee = nil
	}
	return c.put("/rest/api/3/issue/"+key, map[string]any{
		"fields": map[string]any{"assignee": assignee},
	})
}

func (c *RealClient) UpdateDescription(key, description string) error {
	adf := map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []any{
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]any{"type": "text", "text": description},
				},
			},
		},
	}
	return c.put("/rest/api/3/issue/"+key, map[string]any{
		"fields": map[string]any{"description": adf},
	})
}

func (c *RealClient) AddComment(key, body string) error {
	adf := map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []any{
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]any{"type": "text", "text": body},
				},
			},
		},
	}
	return c.post("/rest/api/3/issue/"+key+"/comment", map[string]any{"body": adf}, nil)
}

func (c *RealClient) SaveFilter(name, description, jql string) (*SavedFilter, error) {
	var f SavedFilter
	err := c.post("/rest/api/3/filter", map[string]any{
		"name":        name,
		"description": description,
		"jql":         jql,
	}, &f)
	if err != nil {
		return nil, err
	}
	return &f, nil
}
