package jira

import "time"

type JiraUser struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

type Priority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Status struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	StatusCategory StatusCategory `json:"statusCategory"`
}

type StatusCategory struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type IssueType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Subtask bool   `json:"subtask"`
}

type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type Comment struct {
	ID      string    `json:"id"`
	Author  JiraUser  `json:"author"`
	Body    string    `json:"body"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Issue struct {
	ID          string     `json:"id"`
	Key         string     `json:"key"`
	Summary     string     `json:"summary"`
	Description string     `json:"description"`
	Priority    Priority   `json:"priority"`
	Status      Status     `json:"status"`
	IssueType   IssueType  `json:"issuetype"`
	Project     Project    `json:"project"`
	Assignee    *JiraUser  `json:"assignee"`
	Reporter    *JiraUser  `json:"reporter"`
	Labels      []string   `json:"labels"`
	Sprint      string     `json:"sprint"`
	Comments    []Comment  `json:"comments"`
	Created     time.Time  `json:"created"`
	Updated     time.Time  `json:"updated"`
}

type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   Status `json:"to"`
}

type SavedFilter struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	JQL  string `json:"jql"`
}

type SearchResult struct {
	Issues []Issue `json:"issues"`
	Total  int     `json:"total"`
}
