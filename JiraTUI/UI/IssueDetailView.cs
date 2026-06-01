using System.Text;
using JiraTUI.Jira.Models;
using Terminal.Gui;

namespace JiraTUI.UI
{
    public class IssueDetailView : FrameView
    {
        private readonly TextView _body;
        private readonly Label _hints;
        // Persistent scheme — mutated in place by RefreshColors().
        // TG 1.x caches the ColorScheme reference given at construction time;
        // replacing it with a new object doesn't propagate to the renderer.
        private readonly ColorScheme _bodyScheme = new ColorScheme();

        public IssueDetailView() : base("Details")
        {
            X = 0; Y = 0;
            Width = Dim.Fill();
            Height = Dim.Fill();

            _body = new TextView
            {
                X = 0,
                Y = 0,
                Width = Dim.Fill(),
                Height = Dim.Fill() - 1,
                ReadOnly = true,
                WordWrap = true,
                ColorScheme = _bodyScheme,
            };
            Add(_body);

            _hints = new Label("Ctrl-O open  ·  Alt-I:  P riority · S tatus · A ssignee · D escription · C omment")
            {
                X = 0,
                Y = Pos.AnchorEnd(1),
                Width = Dim.Fill(),
            };
            Add(_hints);

            RefreshColors();
        }

        /// <summary>
        /// All entries in _bodyScheme are pinned to Colors.Base.Normal so that
        /// whichever scheme entry TG 1.x picks for a ReadOnly TextView (Normal,
        /// Focus, Disabled, HotNormal) the rendered color matches the rest of the
        /// Base-colored UI. Mutates _bodyScheme in place — same pattern as
        /// ThemeManager.Apply(), which mutates Colors.Base rather than replacing
        /// the reference (TG 1.x caches the reference given at construction time).
        /// Must be called after every theme switch because Attribute is a value type.
        /// </summary>
        public void RefreshColors()
        {
            this.ColorScheme = Colors.Base;
            var n = Colors.Base.Normal;
            _bodyScheme.Normal    = n;
            _bodyScheme.Focus     = n;
            _bodyScheme.HotNormal = n;
            _bodyScheme.HotFocus  = n;
            _bodyScheme.Disabled  = n;
            _body.SetNeedsDisplay();
            this.SetNeedsDisplay();
        }

        public void ShowError(Issue rowIssue, string error)
        {
            Title = "Details — error";
            var sb = new StringBuilder();
            sb.AppendLine("Nie udało się pobrać pełnych danych issue.");
            sb.AppendLine();
            sb.AppendLine(error ?? "");
            sb.AppendLine();
            sb.AppendLine("── Z wyników wyszukiwania: ──");
            if (rowIssue != null)
            {
                sb.AppendLine(rowIssue.Key + "  " + (rowIssue.IssueType ?? ""));
                sb.AppendLine("Summary  : " + (rowIssue.Summary ?? ""));
                sb.AppendLine("Status   : " + (rowIssue.Status ?? ""));
                sb.AppendLine("Priority : " + (rowIssue.Priority ?? ""));
                sb.AppendLine("Assignee : " + (rowIssue.Assignee ?? "Unassigned"));
            }
            _body.Text = sb.ToString();
            _body.TopRow = 0;
        }

        public void ShowIssue(Issue i)
        {
            if (i == null)
            {
                Title = "Details";
                _body.Text = "(no issue selected)";
                return;
            }

            Title = $"Details — {i.Key}";

            var sb = new StringBuilder();
            sb.AppendLine(IssueGlyphs.TypeGlyph(i.IssueType) + "  " + i.Key + "   " + (i.IssueType ?? ""));
            sb.AppendLine(new string('─', 60));
            sb.AppendLine("Summary  : " + (i.Summary ?? ""));
            sb.AppendLine("Status   : " + IssueGlyphs.StatusGlyph(i.Status) + " " + (i.Status ?? ""));
            sb.AppendLine("Priority : " + IssueGlyphs.PriorityGlyph(i.Priority) + " " + (i.Priority ?? ""));
            sb.AppendLine("Assignee : " + (i.Assignee ?? "Unassigned"));
            sb.AppendLine("Reporter : " + (i.Reporter ?? ""));
            sb.AppendLine("Sprint   : " + (i.Sprint ?? ""));
            sb.AppendLine("Updated  : " + i.Updated.ToLocalTime().ToString("yyyy-MM-dd HH:mm"));
            sb.AppendLine("Labels   : " + ((i.Labels != null && i.Labels.Count > 0) ? string.Join(", ", i.Labels) : "—"));
            sb.AppendLine();
            sb.AppendLine("── Description ──");
            sb.AppendLine(string.IsNullOrEmpty(i.Description) ? "(empty)" : i.Description);
            sb.AppendLine();
            sb.AppendLine("── Comments (" + (i.Comments != null ? i.Comments.Count : 0) + ") ──");
            if (i.Comments != null)
            {
                foreach (var c in i.Comments)
                {
                    sb.AppendLine();
                    sb.AppendLine("[" + c.Author + " · " + c.Created.ToLocalTime().ToString("yyyy-MM-dd HH:mm") + "]");
                    sb.AppendLine(c.Body);
                }
            }

            _body.Text = sb.ToString();
            _body.TopRow = 0;
        }

    }
}
