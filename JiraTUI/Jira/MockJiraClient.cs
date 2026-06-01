using System;
using System.Collections.Generic;
using System.Linq;
using JiraTUI.Jira.Models;

namespace JiraTUI.Jira
{
    public class MockJiraClient : IJiraClient
    {
        private readonly List<Project> _projects;
        private readonly List<Issue> _issues;
        private readonly List<SavedFilter> _filters;

        public string CurrentUserDisplay => "demo.user@example.com";
        public string ServerLabel => "mock://localhost (offline demo)";

        public MockJiraClient()
        {
            _projects = new List<Project>
            {
                new Project { Key = "PROJ",  Name = "Main Product" },
                new Project { Key = "INFRA", Name = "Infrastructure" },
                new Project { Key = "DEV",   Name = "Developer Experience" },
                new Project { Key = "QA",    Name = "Quality Assurance" },
            };

            _filters = new List<SavedFilter>
            {
                new SavedFilter { Name = "My open issues",     Jql = "assignee = currentUser() AND statusCategory != Done" },
                new SavedFilter { Name = "Critical bugs",      Jql = "type = Bug AND priority = Highest" },
                new SavedFilter { Name = "Recently updated",   Jql = "updated >= -7d ORDER BY updated DESC" },
            };

            _issues = new List<Issue>();
            var statuses  = new[] { "Open", "In Progress", "In Review", "Done", "Blocked" };
            var prios     = new[] { "Highest", "High", "Medium", "Low", "Lowest" };
            var types     = new[] { "Bug", "Task", "Story", "Epic" };
            var assignees = new[] { "Jan Kowalski", "Anna Nowak", "Piotr Wiśniewski", "Maria Lewandowska", "demo.user@example.com" };
            var sprints   = new[] { "Sprint 41", "Sprint 42", "Sprint 43", "Backlog" };

            var rnd = new Random(1234);
            int counter = 100;
            foreach (var p in _projects)
            {
                int n = 15 + rnd.Next(10);
                for (int i = 0; i < n; i++)
                {
                    counter++;
                    var issue = new Issue
                    {
                        Key = $"{p.Key}-{counter}",
                        ProjectKey = p.Key,
                        Summary = SampleSummary(rnd),
                        Status = statuses[rnd.Next(statuses.Length)],
                        Priority = prios[rnd.Next(prios.Length)],
                        IssueType = types[rnd.Next(types.Length)],
                        Assignee = assignees[rnd.Next(assignees.Length)],
                        Reporter = assignees[rnd.Next(assignees.Length)],
                        Sprint = sprints[rnd.Next(sprints.Length)],
                        Updated = DateTime.UtcNow.AddHours(-rnd.Next(0, 24 * 14)),
                        Labels = SampleLabels(rnd),
                        Description = SampleDescription(rnd),
                    };

                    int cc = rnd.Next(0, 5);
                    for (int j = 0; j < cc; j++)
                    {
                        issue.Comments.Add(new Comment
                        {
                            Author = assignees[rnd.Next(assignees.Length)],
                            Created = issue.Updated.AddHours(-rnd.Next(0, 200)),
                            Body = SampleComment(rnd),
                        });
                    }

                    _issues.Add(issue);
                }
            }
        }

        public void Dispose() { }

        public bool TestConnection(out string error)
        {
            error = null;
            return true;
        }

        public IList<Project> GetProjects() => _projects;

        public IList<SavedFilter> GetSavedFilters() => _filters;

        public IList<Issue> SearchIssues(string jql, int maxResults)
        {
            IEnumerable<Issue> q = _issues;

            if (!string.IsNullOrWhiteSpace(jql))
            {
                var lower = jql.ToLowerInvariant();

                if (lower.Contains("assignee = currentuser()"))
                    q = q.Where(i => i.Assignee == CurrentUserDisplay);

                if (lower.Contains("statuscategory != done"))
                    q = q.Where(i => i.Status != "Done");

                if (lower.Contains("type = bug"))
                    q = q.Where(i => i.IssueType == "Bug");

                if (lower.Contains("priority = highest"))
                    q = q.Where(i => i.Priority == "Highest");

                var projMarker = "project = ";
                var idx = lower.IndexOf(projMarker, StringComparison.Ordinal);
                if (idx >= 0)
                {
                    var tail = lower.Substring(idx + projMarker.Length).TrimStart();
                    var end = tail.IndexOfAny(new[] { ' ', '\t', '\n' });
                    var proj = end > 0 ? tail.Substring(0, end) : tail;
                    proj = proj.Trim('"');
                    q = q.Where(i => string.Equals(i.ProjectKey, proj, StringComparison.OrdinalIgnoreCase));
                }
            }

            return q.OrderByDescending(i => i.Updated).Take(maxResults).ToList();
        }

        public Issue GetIssue(string key) => _issues.FirstOrDefault(i => i.Key == key);

        public IList<string> GetPriorityNames()
            => new[] { "Highest", "High", "Medium", "Low", "Lowest" };

        public IList<string> GetStatusNames()
            => new[] { "Open", "In Progress", "In Review", "Done", "Blocked" };

        public IList<string> GetIssueTypeNames()
            => new[] { "Bug", "Task", "Story", "Epic", "Sub-task" };

        public IList<Transition> GetTransitions(string issueKey)
        {
            return new List<Transition>
            {
                new Transition { Id = "11", Name = "To Do",       ToStatus = "Open" },
                new Transition { Id = "21", Name = "In Progress", ToStatus = "In Progress" },
                new Transition { Id = "31", Name = "In Review",   ToStatus = "In Review" },
                new Transition { Id = "41", Name = "Done",        ToStatus = "Done" },
                new Transition { Id = "51", Name = "Block",       ToStatus = "Blocked" },
            };
        }

        public IList<JiraUser> SearchAssignableUsers(string issueKey, string query)
        {
            var pool = new[]
            {
                new JiraUser { AccountId = "u-jan",   DisplayName = "Jan Kowalski",        EmailAddress = "jan@example.com" },
                new JiraUser { AccountId = "u-anna",  DisplayName = "Anna Nowak",          EmailAddress = "anna@example.com" },
                new JiraUser { AccountId = "u-piotr", DisplayName = "Piotr Wiśniewski",    EmailAddress = "piotr@example.com" },
                new JiraUser { AccountId = "u-maria", DisplayName = "Maria Lewandowska",   EmailAddress = "maria@example.com" },
                new JiraUser { AccountId = "u-demo",  DisplayName = "demo.user@example.com", EmailAddress = "demo.user@example.com" },
            };
            if (string.IsNullOrWhiteSpace(query)) return pool;
            return pool.Where(u =>
                (u.DisplayName ?? "").IndexOf(query, StringComparison.OrdinalIgnoreCase) >= 0 ||
                (u.EmailAddress ?? "").IndexOf(query, StringComparison.OrdinalIgnoreCase) >= 0
            ).ToList();
        }

        public void SetPriority(string issueKey, string priorityName)
        {
            var i = _issues.FirstOrDefault(x => x.Key == issueKey);
            if (i != null) i.Priority = priorityName;
        }

        public void SetAssignee(string issueKey, string accountIdOrNull)
        {
            var i = _issues.FirstOrDefault(x => x.Key == issueKey);
            if (i != null)
            {
                if (accountIdOrNull == null) { i.Assignee = null; return; }
                // Resolve to display name for nicer UI without a separate user lookup.
                var u = SearchAssignableUsers(issueKey, null).FirstOrDefault(x => x.AccountId == accountIdOrNull);
                i.Assignee = u != null ? u.DisplayName : accountIdOrNull;
            }
        }

        public void TransitionIssue(string issueKey, string transitionId)
        {
            var i = _issues.FirstOrDefault(x => x.Key == issueKey);
            if (i == null) return;
            var t = GetTransitions(issueKey).FirstOrDefault(x => x.Id == transitionId);
            if (t != null) i.Status = t.ToStatus;
        }

        public void UpdateDescription(string issueKey, string plainTextDescription)
        {
            var i = _issues.FirstOrDefault(x => x.Key == issueKey);
            if (i != null) i.Description = plainTextDescription;
        }

        public void AddComment(string issueKey, string plainTextComment)
        {
            var i = _issues.FirstOrDefault(x => x.Key == issueKey);
            if (i == null) return;
            i.Comments.Add(new Comment
            {
                Author = CurrentUserDisplay,
                Created = DateTime.UtcNow,
                Body = plainTextComment ?? "",
            });
        }

        public SavedFilter SaveFilter(string name, string description, string jql)
        {
            var f = new SavedFilter { Name = name, Jql = jql };
            _filters.Add(f);
            return f;
        }

        private static string SampleSummary(Random r)
        {
            var verbs = new[] { "Fix", "Refactor", "Add", "Remove", "Investigate", "Improve", "Update", "Migrate", "Document" };
            var nouns = new[] { "login flow", "export feature", "search performance", "OAuth callback", "DB schema",
                                "audit logging", "session timeout", "cache invalidation", "API pagination",
                                "error reporting", "telemetry pipeline", "user profile screen", "CI build matrix" };
            return $"{verbs[r.Next(verbs.Length)]} {nouns[r.Next(nouns.Length)]}";
        }

        private static List<string> SampleLabels(Random r)
        {
            var pool = new[] { "backend", "frontend", "db", "tech-debt", "security", "perf", "ux", "ci", "docs" };
            int n = r.Next(0, 3);
            var picked = new HashSet<string>();
            for (int i = 0; i < n; i++) picked.Add(pool[r.Next(pool.Length)]);
            return picked.ToList();
        }

        private static string SampleDescription(Random r)
        {
            return
                "## Context\r\n" +
                "Customer-reported issue affecting workflow.\r\n\r\n" +
                "## Steps to reproduce\r\n" +
                "1. Open the affected screen\r\n" +
                "2. Perform the action that triggers the bug\r\n" +
                "3. Observe the unexpected result\r\n\r\n" +
                "## Expected\r\n" +
                "Operation succeeds without error.\r\n\r\n" +
                "## Notes\r\n" +
                "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
                "Vestibulum sit amet pulvinar tellus. Quisque venenatis dignissim libero.";
        }

        private static string SampleComment(Random r)
        {
            var pool = new[]
            {
                "Reproduced locally on develop branch.",
                "PR opened, awaiting review.",
                "Blocked on dependency upgrade.",
                "Discussed with QA — needs additional acceptance criteria.",
                "Cannot reproduce on staging. Closing unless we get more details.",
                "Looks like a duplicate of an older ticket — investigating.",
            };
            return pool[r.Next(pool.Length)];
        }
    }
}
