package jira

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RealClient struct {
	baseURL    string
	authHeader string
	http       *http.Client
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
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(v)
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

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(b))
	}
	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
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

type searchPayload struct {
	Jql        string   `json:"jql"`
	MaxResults int      `json:"maxResults"`
	Fields     []string `json:"fields"`
}

type searchResponse struct {
	Issues []rawIssue `json:"issues"`
	Total  int        `json:"total"`
}

type rawIssue struct {
	ID     string    `json:"id"`
	Key    string    `json:"key"`
	Fields rawFields `json:"fields"`
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
	Created     time.Time       `json:"created"`
	Updated     time.Time       `json:"updated"`
	Comment     *struct {
		Comments []rawComment `json:"comments"`
	} `json:"comment"`
	Sprint json.RawMessage `json:"sprint"`
}

type rawComment struct {
	ID      string          `json:"id"`
	Author  JiraUser        `json:"author"`
	Body    json.RawMessage `json:"body"`
	Created time.Time       `json:"created"`
	Updated time.Time       `json:"updated"`
}

func (c *RealClient) SearchIssues(jql string, maxResults int) ([]Issue, int, error) {
	payload := searchPayload{
		Jql:        jql,
		MaxResults: maxResults,
		Fields: []string{
			"*navigable", "description", "comment", "priority", "status",
			"assignee", "reporter", "issuetype", "project", "summary",
			"labels", "sprint",
		},
	}
	var result searchResponse
	if err := c.post("/rest/api/3/issue/search", payload, &result); err != nil {
		return nil, 0, err
	}
	issues := make([]Issue, len(result.Issues))
	for i, ri := range result.Issues {
		issues[i] = convertIssue(ri)
	}
	return issues, result.Total, nil
}

func convertIssue(ri rawIssue) Issue {
	f := ri.Fields
	issue := Issue{
		ID:      ri.ID,
		Key:     ri.Key,
		Summary: f.Summary,
		Labels:  f.Labels,
		Created: f.Created,
		Updated: f.Updated,
	}
	if f.Description != nil {
		issue.Description = string(f.Description)
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
	if f.Comment != nil {
		for _, rc := range f.Comment.Comments {
			issue.Comments = append(issue.Comments, Comment{
				ID:      rc.ID,
				Author:  rc.Author,
				Body:    string(rc.Body),
				Created: rc.Created,
				Updated: rc.Updated,
			})
		}
	}
	return issue
}

func (c *RealClient) GetIssue(key string) (*Issue, error) {
	var ri rawIssue
	if err := c.get("/rest/api/3/issue/"+key+"?fields=*all", &ri); err != nil {
		return nil, err
	}
	issue := convertIssue(ri)
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
