using System.Collections.Generic;
using Terminal.Gui;

namespace JiraTUI.UI.Dialogs
{
    /// <summary>
    /// Read-only modal that explains the glyphs used across the UI for issue
    /// type, priority and status. Each row paints its own glyph in the colour
    /// that the table uses, so the dialog doubles as a colour reference.
    /// Two-column layout keeps it short enough for typical terminals.
    /// </summary>
    public class LegendDialog : Dialog
    {
        private const int LeftCol  = 2;
        private const int RightCol = 34;

        public LegendDialog() : base("Legend  (Esc to close)", 64, 20)
        {
            // Left column: issue types (the longest section).
            AddSection("ISSUE TYPE", IssueGlyphs.AllTypes(), LeftCol, 0);

            // Right column: priority then status, stacked.
            int y = AddSection("PRIORITY", IssueGlyphs.AllPriorities(), RightCol, 0);
            AddSection("STATUS", IssueGlyphs.AllStatuses(), RightCol, y + 1);

            var close = new Button("Close", true);
            close.Clicked += () => Application.RequestStop();
            AddButton(close);
        }

        /// <summary>
        /// Renders one titled block at (<paramref name="x"/>, <paramref name="y"/>)
        /// and returns the next free Y below it.
        /// </summary>
        private int AddSection(string title, IList<IssueGlyphs.Entry> entries, int x, int y)
        {
            Add(new Label(title) { X = x, Y = y, Width = 28 });
            y++;
            Add(new Label(new string('─', title.Length)) { X = x, Y = y, Width = 28 });
            y++;

            foreach (var e in entries)
            {
                var glyphLabel = new Label(e.Glyph) { X = x, Y = y, Width = 2 };
                if (e.Color.HasValue)
                    glyphLabel.ColorScheme = BuildFgScheme(e.Color.Value);
                Add(glyphLabel);

                Add(new Label(e.Name) { X = x + 3, Y = y, Width = 25 });
                y++;
            }
            return y;
        }

        /// <summary>
        /// ColorScheme whose every entry uses the given foreground over the
        /// dialog's normal background. Mirrors what the table's ColorGetter
        /// does for glyph columns, so colours line up visually.
        /// </summary>
        private static ColorScheme BuildFgScheme(Color fg)
        {
            var bg = Colors.Dialog.Normal.Background;
            var drv = Application.Driver;
            var attr = drv != null
                ? drv.MakeAttribute(fg, bg)
                : Terminal.Gui.Attribute.Make(fg, bg);
            return new ColorScheme
            {
                Normal    = attr,
                Focus     = attr,
                HotNormal = attr,
                HotFocus  = attr,
                Disabled  = attr,
            };
        }
    }
}
