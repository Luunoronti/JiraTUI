using System;
using System.Collections.Generic;

namespace JiraTUI.Jira.Models
{
    public class Issue
    {
        public string Key { get; set; }
        public string Summary { get; set; }
        public string Status { get; set; }
        public string Priority { get; set; }
        public string Assignee { get; set; }
        public string Reporter { get; set; }
        public string IssueType { get; set; }
        public string ProjectKey { get; set; }
        public string Sprint { get; set; }
        public List<string> Labels { get; set; } = new List<string>();
        public string Description { get; set; }
        public DateTime Updated { get; set; }
        public List<Comment> Comments { get; set; } = new List<Comment>();
    }

    public class Comment
    {
        public string Author { get; set; }
        public DateTime Created { get; set; }
        public string Body { get; set; }
    }

    public class Project
    {
        public string Key { get; set; }
        public string Name { get; set; }
    }

    public class SavedFilter
    {
        public string Name { get; set; }
        public string Jql { get; set; }
    }

    public class Transition
    {
        public string Id { get; set; }
        public string Name { get; set; }
        public string ToStatus { get; set; }
    }

    public class JiraUser
    {
        public string AccountId { get; set; }
        public string DisplayName { get; set; }
        public string EmailAddress { get; set; }

        public override string ToString()
        {
            if (string.IsNullOrEmpty(EmailAddress)) return DisplayName ?? AccountId ?? "(unknown)";
            return (DisplayName ?? "") + " <" + EmailAddress + ">";
        }
    }
}
