package ai

import (
	"strings"

	"jiratui/jira"
)

// BuildSystemPrompt constructs a system prompt that instructs the AI to convert
// natural language queries to JQL, using the available Jira metadata for context.
func BuildSystemPrompt(
	projects []jira.Project,
	statuses []jira.Status,
	priorities []jira.Priority,
	issueTypes []jira.IssueType,
) string {
	var sb strings.Builder

	sb.WriteString("You are a Jira JQL expert. Convert natural language queries to valid JQL.\n\n")

	if len(projects) > 0 {
		keys := make([]string, len(projects))
		for i, p := range projects {
			keys[i] = p.Key
		}
		sb.WriteString("Available project keys: ")
		sb.WriteString(strings.Join(keys, ", "))
		sb.WriteString("\n")
	}

	if len(statuses) > 0 {
		names := make([]string, len(statuses))
		for i, s := range statuses {
			names[i] = s.Name
		}
		sb.WriteString("Available statuses: ")
		sb.WriteString(strings.Join(names, ", "))
		sb.WriteString("\n")
	}

	if len(priorities) > 0 {
		names := make([]string, len(priorities))
		for i, p := range priorities {
			names[i] = p.Name
		}
		sb.WriteString("Available priorities: ")
		sb.WriteString(strings.Join(names, ", "))
		sb.WriteString("\n")
	}

	if len(issueTypes) > 0 {
		names := make([]string, len(issueTypes))
		for i, t := range issueTypes {
			names[i] = t.Name
		}
		sb.WriteString("Available issue types: ")
		sb.WriteString(strings.Join(names, ", "))
		sb.WriteString("\n")
	}

	sb.WriteString("\nRules:\n")
	sb.WriteString("- Use currentUser() for the current user.\n")
	sb.WriteString("- Use relative dates like -7d, -30d for recency.\n")
	sb.WriteString("- Prefer statusCategory for broad status matching (e.g. statusCategory != Done).\n")
	sb.WriteString("- Quote values containing spaces.\n\n")
	sb.WriteString("Respond with ONLY the JQL query, no explanation.")

	return sb.String()
}
