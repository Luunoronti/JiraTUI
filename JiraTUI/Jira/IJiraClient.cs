using System;
using System.Collections.Generic;
using JiraTUI.Jira.Models;

namespace JiraTUI.Jira
{
    public interface IJiraClient : IDisposable
    {
        bool TestConnection(out string error);
        IList<Project> GetProjects();
        IList<Issue> SearchIssues(string jql, int maxResults);
        Issue GetIssue(string key);
        IList<SavedFilter> GetSavedFilters();
        string CurrentUserDisplay { get; }
        string ServerLabel { get; }

        // Metadata
        IList<string> GetPriorityNames();
        IList<string> GetStatusNames();
        IList<string> GetIssueTypeNames();
        IList<Transition> GetTransitions(string issueKey);
        IList<JiraUser> SearchAssignableUsers(string issueKey, string query);

        // Mutations
        void SetPriority(string issueKey, string priorityName);
        void SetAssignee(string issueKey, string accountIdOrNull);
        void TransitionIssue(string issueKey, string transitionId);
        void UpdateDescription(string issueKey, string plainTextDescription);
        void AddComment(string issueKey, string plainTextComment);

        // Filters
        SavedFilter SaveFilter(string name, string description, string jql);
    }
}
