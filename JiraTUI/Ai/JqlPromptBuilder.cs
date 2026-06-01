using System;
using System.Linq;
using System.Text;
using JiraTUI.Jira;

namespace JiraTUI.Ai
{
    /// <summary>
    /// Builds the system prompt sent to Claude. Includes live context fetched from
    /// the user's Jira instance (projects, statuses, priorities, types) so the
    /// generated JQL uses values that actually exist for them.
    /// </summary>
    public static class JqlPromptBuilder
    {
        public static string BuildSystemPrompt(IJiraClient jira)
        {
            var sb = new StringBuilder();

            sb.AppendLine("You are a Jira JQL (Jira Query Language) expert.");
            sb.AppendLine("Given a natural-language request, output ONLY the corresponding JQL query.");
            sb.AppendLine("Do not include explanations, markdown, code fences, language tags, or any commentary — just the raw JQL on a single response.");
            sb.AppendLine();
            sb.AppendLine("Context for this Jira instance:");
            sb.AppendLine("- Current user display name: " + (jira.CurrentUserDisplay ?? "(unknown)"));
            sb.AppendLine("  Use currentUser() for any reference to \"me\", \"my\", \"mine\".");

            TryAppendLine(sb, () =>
            {
                var projects = jira.GetProjects();
                if (projects == null || projects.Count == 0) return null;
                return "- Projects: " + string.Join(", ",
                    projects.Select(p => p.Key + " (" + (p.Name ?? "") + ")"));
            });

            TryAppendLine(sb, () =>
            {
                var statuses = jira.GetStatusNames();
                if (statuses == null || statuses.Count == 0) return null;
                return "- Statuses: " + string.Join(", ", statuses);
            });

            TryAppendLine(sb, () =>
            {
                var priorities = jira.GetPriorityNames();
                if (priorities == null || priorities.Count == 0) return null;
                return "- Priorities: " + string.Join(", ", priorities);
            });

            TryAppendLine(sb, () =>
            {
                var types = jira.GetIssueTypeNames();
                if (types == null || types.Count == 0) return null;
                return "- Issue types: " + string.Join(", ", types);
            });

            sb.AppendLine();
            sb.AppendLine("JQL syntax rules:");
            sb.AppendLine("- Quote any value containing spaces or special chars: status = \"In Progress\"");
            sb.AppendLine("- Operators: =, !=, >, >=, <, <=, IN (...), NOT IN (...), ~ (contains text), !~");
            sb.AppendLine("- Combine with AND, OR, NOT; group with parentheses");
            sb.AppendLine("- IS EMPTY / IS NOT EMPTY for null checks (e.g. assignee IS EMPTY)");
            sb.AppendLine("- Functions: currentUser(), now(), startOfDay(), endOfDay(), startOfWeek(), endOfWeek(), startOfMonth()");
            sb.AppendLine("- ORDER BY field [ASC|DESC], comma-separated for multiple sort keys");
            sb.AppendLine("- Dates: 'yyyy-MM-dd' literal, or relative like -7d, -1w, -2M, -1y");
            sb.AppendLine();
            sb.AppendLine("Common fields: project, issuetype, status, priority, assignee, reporter, creator,");
            sb.AppendLine("created, updated, resolved, due, labels, sprint, fixVersion, affectedVersion, summary,");
            sb.AppendLine("description, comment, text (full-text), parent, epic.");

            return sb.ToString();
        }

        private static void TryAppendLine(StringBuilder sb, Func<string> fn)
        {
            try
            {
                var line = fn();
                if (!string.IsNullOrEmpty(line)) sb.AppendLine(line);
            }
            catch { /* skip if data not available */ }
        }
    }
}
