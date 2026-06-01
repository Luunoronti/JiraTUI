using System;
using System.Collections.Generic;
using System.Data;
using JiraTUI.Config;
using JiraTUI.Jira.Models;
using Terminal.Gui;

namespace JiraTUI.UI
{
    public class IssueListView : FrameView
    {
        private readonly TableView _table;
        private DataTable _data;
        private List<Issue> _current = new List<Issue>();
        private int _lastSummaryWidth = -1;
        private ColumnVisibilityConfig _columns = new ColumnVisibilityConfig();
        private List<ColumnDef> _activeColumns = new List<ColumnDef>();

        private const int SummaryMinWidth = 20;
        private const int SummaryMaxWidth = 300;

        public event Action<Issue> SelectionChanged;

        public IssueListView() : base("Issues")
        {
            X = 0; Y = 0;
            Width = Dim.Fill();
            Height = Dim.Fill();

            _table = new TableView(null)
            {
                X = 0,
                Y = 0,
                Width = Dim.Fill(),
                Height = Dim.Fill(),
                FullRowSelect = true,
            };

            _table.Style.AlwaysShowHeaders = true;
            _table.Style.ExpandLastColumn = true;
            _table.Style.ShowHorizontalHeaderOverline = false;
            _table.Style.ShowHorizontalHeaderUnderline = true;
            _table.Style.ShowVerticalCellLines = false;
            _table.Style.ShowVerticalHeaderLines = false;

            _table.SelectedCellChanged += (e) => RaiseSelection();

            Add(_table);

            BuildDataTable();
        }

        public void SetColumns(ColumnVisibilityConfig columns)
        {
            _columns = columns ?? new ColumnVisibilityConfig();
            BuildDataTable();
            RebuildRows();
            Title = $"Issues ({_current.Count})";
            _table.SetNeedsDisplay();
        }

        public void SetIssues(IList<Issue> issues)
        {
            _current = new List<Issue>(issues);
            RebuildRows();

            if (_current.Count > 0)
            {
                _table.SelectedRow = 0;
                _table.SelectedColumn = 0;
            }

            Title = $"Issues ({_current.Count})";
            _table.SetNeedsDisplay();
            RaiseSelection();
        }

        public Issue Selected
        {
            get
            {
                int row = _table.SelectedRow;
                if (row < 0 || row >= _current.Count) return null;
                return _current[row];
            }
        }

        /// <summary>
        /// Replace an in-memory issue (matched by Key) and repaint its row so the
        /// list reflects updates without re-running the entire JQL search.
        /// </summary>
        public void UpdateIssue(Issue updated)
        {
            if (updated == null || string.IsNullOrEmpty(updated.Key)) return;
            for (int idx = 0; idx < _current.Count; idx++)
            {
                if (_current[idx].Key == updated.Key)
                {
                    _current[idx] = updated;
                    int w = _lastSummaryWidth > 0 ? _lastSummaryWidth : 80;
                    for (int c = 0; c < _activeColumns.Count; c++)
                    {
                        _data.Rows[idx][c] = _activeColumns[c].Get(updated, w);
                    }
                    _table.SetNeedsDisplay();
                    return;
                }
            }
        }

        public override void LayoutSubviews()
        {
            base.LayoutSubviews();
            // Re-truncate Summary to current available width whenever the panel
            // is laid out (toggle details/nav, terminal resize, column changes).
            if (_current == null || _current.Count == 0 || _activeColumns == null) return;
            int sumIdx = _activeColumns.FindIndex(c => c.Id == "Summary");
            if (sumIdx < 0) return;

            int w = ComputeSummaryWidth();
            if (w == _lastSummaryWidth) return;
            _lastSummaryWidth = w;
            for (int i = 0; i < _current.Count; i++)
                _data.Rows[i][sumIdx] = Truncate(_current[i].Summary, w);
            _table.SetNeedsDisplay();
        }

        public void FocusTable() => _table.SetFocus();

        // =========================================================================
        // Internals
        // =========================================================================

        private class ColumnDef
        {
            public string Id;                            // logical id (Key/Type/Priority/...)
            public string Header;                        // displayed header text
            public Func<Issue, int, object> Get;         // (issue, summaryWidth) -> cell value
            public int MaxWidth;                         // 0 = unbounded
            public int EstimatedWidth;                   // for ComputeSummaryWidth math
        }

        private void BuildDataTable()
        {
            _data = new DataTable();
            _activeColumns = new List<ColumnDef>();

            if (_columns.Key)
                AddCol(new ColumnDef
                {
                    Id = "Key", Header = "Key",
                    Get = (i, _) => i.Key, EstimatedWidth = 10
                });

            if (_columns.Type)
                AddCol(new ColumnDef
                {
                    Id = "Type", Header = " ",
                    Get = (i, _) => IssueGlyphs.TypeGlyph(i.IssueType) + " ", EstimatedWidth = 2
                });

            if (_columns.Priority)
                AddCol(new ColumnDef
                {
                    Id = "Priority", Header = "  ",
                    Get = (i, _) => IssueGlyphs.PriorityGlyph(i.Priority) + " ", EstimatedWidth = 2
                });

            if (_columns.Status)
                AddCol(new ColumnDef
                {
                    // Glyph-only column — the legend (Ctrl-L) explains the meaning.
                    // Full status name is still visible in the details panel.
                    // Header is NBSP (U+00A0) so it renders blank but stays distinct
                    // from Type's regular-space header in the underlying DataTable.
                    Id = "Status", Header = " ",
                    Get = (i, _) => IssueGlyphs.StatusGlyph(i.Status) + " ", EstimatedWidth = 2
                });

            if (_columns.Assignee)
                AddCol(new ColumnDef
                {
                    Id = "Assignee", Header = "Assignee",
                    Get = (i, _) => Truncate(i.Assignee ?? "Unassigned", 20),
                    MaxWidth = 12, EstimatedWidth = 12,
                });

            if (_columns.Summary)
                AddCol(new ColumnDef
                {
                    Id = "Summary", Header = "Summary",
                    Get = (i, w) => Truncate(i.Summary, w),
                    EstimatedWidth = 30, // not used (Summary is the expanded column)
                });

            _table.Table = _data;

            // Apply ColumnStyle.MaxWidth for the columns that need it.
            foreach (var def in _activeColumns)
            {
                if (def.MaxWidth > 0)
                {
                    var style = _table.Style.GetOrCreateColumnStyle(_data.Columns[def.Header]);
                    style.MaxWidth = def.MaxWidth;
                }
            }

            // Glyph-column colouring — foreground only, background inherited from the
            // row's live scheme so focus/selection highlight still works. Status is
            // intentionally left uncoloured: its glyph alone conveys state, and a
            // hard-coded foreground clashes with some themes' row colours.
            ApplyGlyphColor("Type",     IssueGlyphs.TypeColor);
            ApplyGlyphColor("Priority", IssueGlyphs.PriorityColor);

            _lastSummaryWidth = -1; // force recompute on next layout/rebuild
        }

        private void AddCol(ColumnDef def)
        {
            _data.Columns.Add(def.Header);
            _activeColumns.Add(def);
        }

        private void RebuildRows()
        {
            if (_data == null || _activeColumns == null) return;
            int prevSel = _table.SelectedRow;
            int sumIdx  = _activeColumns.FindIndex(c => c.Id == "Summary");

            _data.Rows.Clear();

            // Pass 1: populate every column EXCEPT Summary so we can measure the
            // actual rendered widths of the fixed columns from real data.
            var tempRows = new object[_current.Count][];
            for (int i = 0; i < _current.Count; i++)
            {
                tempRows[i] = new object[_activeColumns.Count];
                for (int c = 0; c < _activeColumns.Count; c++)
                {
                    if (c == sumIdx) continue;
                    tempRows[i][c] = _activeColumns[c].Get(_current[i], 0);
                }
            }

            // Update EstimatedWidth for each fixed column to the actual max content
            // width (capped at MaxWidth). This means Status="To Do" columns won't
            // steal space reserved for MaxWidth=15 that the data never uses.
            for (int c = 0; c < _activeColumns.Count; c++)
            {
                if (c == sumIdx) continue;
                var def = _activeColumns[c];
                int maxLen = def.Header.Length; // column is at least as wide as header
                for (int i = 0; i < tempRows.Length; i++)
                {
                    int len = tempRows[i][c]?.ToString().Length ?? 0;
                    if (len > maxLen) maxLen = len;
                }
                if (def.MaxWidth > 0) maxLen = Math.Min(maxLen, def.MaxWidth);
                def.EstimatedWidth = maxLen;
            }

            // Pass 2: now that EstimatedWidths are accurate, compute Summary width
            // and fill in all rows.
            int w = ComputeSummaryWidth();
            _lastSummaryWidth = w;

            for (int i = 0; i < _current.Count; i++)
            {
                if (sumIdx >= 0)
                    tempRows[i][sumIdx] = _activeColumns[sumIdx].Get(_current[i], w);
                _data.Rows.Add(tempRows[i]);
            }

            // Preserve selection across rebuilds.
            if (_current.Count > 0)
            {
                int newSel = prevSel < 0 ? 0 : Math.Min(prevSel, _current.Count - 1);
                _table.SelectedRow = newSel;
                _table.SelectedColumn = 0;
            }
        }

        private int ComputeSummaryWidth()
        {
            int total = _table.Bounds.Width;
            if (total <= 0) total = Bounds.Width - 2;
            if (total <= 0) total = Application.Driver != null ? Application.Driver.Cols : 80;

            int fixedSum = 0;
            foreach (var def in _activeColumns)
            {
                if (def.Id == "Summary") continue;
                fixedSum += def.EstimatedWidth;
            }
            int separators = _activeColumns.Count > 1 ? _activeColumns.Count - 1 : 0;

            int remaining = total - fixedSum - separators;
            if (remaining < SummaryMinWidth) remaining = SummaryMinWidth;
            if (remaining > SummaryMaxWidth) remaining = SummaryMaxWidth;
            return remaining;
        }

        private void RaiseSelection()
        {
            var h = SelectionChanged;
            if (h != null) h(Selected);
        }

        // Returns a ColorScheme that keeps the priority foreground colour but inherits
        // backgrounds from the row's live scheme — works correctly with all themes and
        // with the focus/selection highlight.
        private static ColorScheme BuildPriorityScheme(Color fg, ColorScheme row)
        {
            var drv = Application.Driver;
            Terminal.Gui.Attribute A(Color f, Color b) =>
                drv != null ? drv.MakeAttribute(f, b) : Terminal.Gui.Attribute.Make(f, b);
            return new ColorScheme
            {
                Normal    = A(fg, row.Normal.Background),
                Focus     = A(fg, row.Focus.Background),
                HotNormal = A(fg, row.HotNormal.Background),
                HotFocus  = A(fg, row.HotFocus.Background),
                Disabled  = A(Color.DarkGray, row.Disabled.Background),
            };
        }

        /// <summary>
        /// Wire a foreground-only ColorGetter on a glyph column. <paramref name="resolver"/>
        /// receives the rendered glyph (the cell's Representation) and returns the
        /// foreground colour to apply, or null to inherit the row scheme.
        /// </summary>
        private void ApplyGlyphColor(string columnId, Func<string, Color?> resolver)
        {
            int idx = _activeColumns.FindIndex(c => c.Id == columnId);
            if (idx < 0) return;

            var style = _table.Style.GetOrCreateColumnStyle(
                _data.Columns[_activeColumns[idx].Header]);

            style.ColorGetter = args =>
            {
                var glyph = args.Representation?.TrimEnd();
                var fg = resolver(glyph);
                return fg.HasValue
                    ? BuildPriorityScheme(fg.Value, args.RowScheme)
                    : null;
            };
        }

        private static string Truncate(string s, int max)
        {
            if (string.IsNullOrEmpty(s)) return "";
            if (max <= 0) return "";
            return s.Length <= max ? s : s.Substring(0, max - 1) + "…";
        }
    }
}
