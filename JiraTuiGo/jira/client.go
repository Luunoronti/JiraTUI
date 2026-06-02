package jira

type Client interface {
	GetCurrentUser() (*JiraUser, error)
	SearchIssues(jql string, maxResults int) ([]Issue, int, error)
	GetIssue(key string) (*Issue, error)
	GetProjects() ([]Project, error)
	GetPriorities() ([]Priority, error)
	GetStatuses() ([]Status, error)
	GetIssueTypes() ([]IssueType, error)
	GetSavedFilters() ([]SavedFilter, error)
	GetTransitions(key string) ([]Transition, error)
	DoTransition(key, transitionID string) error
	SearchAssignableUsers(query, projectKey string) ([]JiraUser, error)
	UpdatePriority(key, priorityName string) error
	UpdateAssignee(key, accountID string) error
	UpdateDescription(key, description string) error
	AddComment(key, body string) error
	SaveFilter(name, description, jql string) (*SavedFilter, error)
}
