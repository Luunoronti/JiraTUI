using System;
using System.Collections.Generic;
using JiraTUI.Jira;
using JiraTUI.Jira.Models;
using Terminal.Gui;

namespace JiraTUI.UI.Dialogs
{
    /// <summary>
    /// Pick a user (or unassign). Returns AccountId, or "" sentinel for unassign,
    /// or null if cancelled.
    /// </summary>
    public class AssigneeDialog : Dialog
    {
        public const string UnassignSentinel = "";

        public string SelectedAccountId { get; private set; } // null = cancel
        public bool Saved { get; private set; }

        private readonly IJiraClient _jira;
        private readonly string _issueKey;
        private readonly TextField _query;
        private readonly ListView _list;
        private List<object> _items = new List<object>(); // first row is "Unassign", rest JiraUser

        public AssigneeDialog(IJiraClient jira, string issueKey, string currentAssignee)
            : base("Assign — " + issueKey, 70, 18)
        {
            _jira = jira;
            _issueKey = issueKey;

            Add(new Label("Search:") { X = 1, Y = 0 });
            _query = new TextField("") { X = 9, Y = 0, Width = Dim.Fill() - 2 };
            _query.KeyDown += e =>
            {
                if (e.KeyEvent.Key == Key.Enter)
                {
                    RunSearch();
                    if (_items.Count > 0) _list.SetFocus();
                    e.Handled = true;
                }
                else if (e.KeyEvent.Key == Key.CursorDown)
                {
                    // Down arrow from search → jump straight into the list.
                    _list.SetFocus();
                    e.Handled = true;
                }
            };
            Add(_query);

            _list = new ListView(_items)
            {
                X = 1,
                Y = 2,
                Width = Dim.Fill() - 2,
                Height = Dim.Fill() - 2,
            };
            _list.OpenSelectedItem += _ => Commit();
            Add(_list);

            var ok = new Button("OK", true);
            ok.Clicked += Commit;
            var cancel = new Button("Cancel");
            cancel.Clicked += () => { Saved = false; SelectedAccountId = null; Application.RequestStop(); };
            AddButton(cancel);
            AddButton(ok);

            RunSearch(); // initial load with empty query

            Loaded += () => _list.SetFocus();
        }

        private void RunSearch()
        {
            try
            {
                var users = _jira.SearchAssignableUsers(_issueKey, _query.Text != null ? _query.Text.ToString() : "");
                _items = new List<object>();
                _items.Add("(Unassign)");
                foreach (var u in users) _items.Add(u);
                _list.SetSource(_items);
                if (_items.Count > 0) _list.SelectedItem = 0;
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Search failed", ex.Message, "OK");
            }
        }

        private void Commit()
        {
            int idx = _list.SelectedItem;
            if (idx < 0 || idx >= _items.Count)
            {
                Saved = false;
                SelectedAccountId = null;
                Application.RequestStop();
                return;
            }

            var pick = _items[idx];
            if (pick is JiraUser u) SelectedAccountId = u.AccountId;
            else SelectedAccountId = UnassignSentinel; // first row

            Saved = true;
            Application.RequestStop();
        }
    }
}
