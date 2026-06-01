using System;
using System.Collections.Generic;
using JiraTUI.Jira.Models;
using Terminal.Gui;
using Terminal.Gui.Trees;

namespace JiraTUI.UI
{
    public class NavigationView : FrameView
    {
        private readonly TreeView _tree;

        public event Action<string, string> NavSelected; // (path, jql)

        public NavigationView() : base("Navigation")
        {
            X = 0; Y = 0;
            Width = Dim.Fill();
            Height = Dim.Fill();

            _tree = new TreeView
            {
                X = 0,
                Y = 0,
                Width = Dim.Fill(),
                Height = Dim.Fill(),
            };

            _tree.SelectionChanged += (s, e) =>
            {
                var node = _tree.SelectedObject as NavNode;
                if (node == null) return;

                var path = node.BuildPath();
                var h = NavSelected;
                if (h != null) h(path, node.Jql);
            };

            Add(_tree);
        }

        public void Populate(IList<Project> projects, IList<SavedFilter> filters)
        {
            _tree.ClearObjects();

            var quick = new NavNode("Quick views", null);
            quick.AddChild(new NavNode("My open issues",
                "assignee = currentUser() AND statusCategory != Done ORDER BY updated DESC"));
            quick.AddChild(new NavNode("Reported by me",
                "reporter = currentUser() ORDER BY updated DESC"));
            quick.AddChild(new NavNode("Recently updated",
                "updated >= -7d ORDER BY updated DESC"));
            quick.AddChild(new NavNode("All issues",
                "ORDER BY updated DESC"));

            var projRoot = new NavNode("Projects", null);
            if (projects != null)
            {
                foreach (var p in projects)
                {
                    var pn = new NavNode(p.Key + " — " + p.Name, "project = " + p.Key + " ORDER BY updated DESC");
                    pn.AddChild(new NavNode("Backlog",
                        "project = " + p.Key + " AND statusCategory = \"To Do\" ORDER BY priority DESC"));
                    pn.AddChild(new NavNode("In progress",
                        "project = " + p.Key + " AND statusCategory = \"In Progress\" ORDER BY updated DESC"));
                    pn.AddChild(new NavNode("Done",
                        "project = " + p.Key + " AND statusCategory = Done ORDER BY resolved DESC"));
                    projRoot.AddChild(pn);
                }
            }

            var filterRoot = new NavNode("Saved filters", null);
            if (filters != null)
            {
                foreach (var f in filters)
                    filterRoot.AddChild(new NavNode(f.Name, f.Jql));
            }

            _tree.AddObject(quick);
            _tree.AddObject(projRoot);
            _tree.AddObject(filterRoot);

            _tree.Expand(quick);
            _tree.Expand(projRoot);
        }

        public void FocusTree() => _tree.SetFocus();

        private class NavNode : ITreeNode
        {
            public string Label { get; }
            public string Jql { get; }
            public string Text { get => Label; set { } }
            public object Tag { get; set; }
            public IList<ITreeNode> Children { get; } = new List<ITreeNode>();
            public NavNode Parent { get; private set; }

            public NavNode(string label, string jql)
            {
                Label = label;
                Jql = jql;
            }

            public void AddChild(NavNode child)
            {
                child.Parent = this;
                Children.Add(child);
            }

            public string BuildPath()
            {
                var parts = new List<string>();
                for (var n = this; n != null; n = n.Parent) parts.Insert(0, n.Label);
                return string.Join(" › ", parts);
            }

            public override string ToString() => Label;
        }
    }
}
