using System.Collections.Generic;
using Terminal.Gui;

namespace JiraTUI.UI
{
    /// <summary>
    /// Single source of truth for the glyph + colour mapping of issue type,
    /// priority and status. Used by IssueListView (table column colouring),
    /// IssueDetailView (inline glyph prefix) and LegendDialog (Ctrl-L).
    /// Glyphs are picked from BMP ranges that Cascadia Mono renders with
    /// uniform width — no emoji-font fallback surprises in the table.
    /// </summary>
    public static class IssueGlyphs
    {
        public class Entry
        {
            public string Name;
            public string Glyph;
            public Color? Color; // null = inherit row foreground
        }

        // ---------- Type ----------

        public static string TypeGlyph(string t)
        {
            var e = TypeEntry(t);
            return e != null ? e.Glyph : "·";
        }

        public static Color? TypeColor(string glyph)
        {
            switch (glyph)
            {
                case "⊘": return Terminal.Gui.Color.Red;
                case "✓": return Terminal.Gui.Color.BrightCyan;
                case "★": return Terminal.Gui.Color.BrightGreen;
                case "⬢": return Terminal.Gui.Color.BrightMagenta;
                case "↳": return Terminal.Gui.Color.Gray;
                case "⚒": return Terminal.Gui.Color.BrightBlue;
                case "✦": return Terminal.Gui.Color.BrightYellow;
                case "⌬": return Terminal.Gui.Color.Brown;
                case "‼": return Terminal.Gui.Color.BrightRed;
                case "✉": return Terminal.Gui.Color.Cyan;
                default:  return null;
            }
        }

        private static Entry TypeEntry(string t)
        {
            if (string.IsNullOrEmpty(t)) return null;
            switch (t)
            {
                case "Bug":             return new Entry { Name = "Bug",             Glyph = "⊘", Color = Terminal.Gui.Color.Red };
                case "Task":            return new Entry { Name = "Task",            Glyph = "✓", Color = Terminal.Gui.Color.BrightCyan };
                case "Story":           return new Entry { Name = "Story",           Glyph = "★", Color = Terminal.Gui.Color.BrightGreen };
                case "Epic":            return new Entry { Name = "Epic",            Glyph = "⬢", Color = Terminal.Gui.Color.BrightMagenta };
                case "Sub-task":
                case "Subtask":         return new Entry { Name = "Sub-task",        Glyph = "↳", Color = Terminal.Gui.Color.Gray };
                case "Improvement":     return new Entry { Name = "Improvement",     Glyph = "⚒", Color = Terminal.Gui.Color.BrightBlue };
                case "New Feature":     return new Entry { Name = "New Feature",     Glyph = "✦", Color = Terminal.Gui.Color.BrightYellow };
                case "Test":            return new Entry { Name = "Test",            Glyph = "⌬", Color = Terminal.Gui.Color.Brown };
                case "Incident":        return new Entry { Name = "Incident",        Glyph = "‼", Color = Terminal.Gui.Color.BrightRed };
                case "Service Request": return new Entry { Name = "Service Request", Glyph = "✉", Color = Terminal.Gui.Color.Cyan };
                default: return null;
            }
        }

        public static IList<Entry> AllTypes()
        {
            return new List<Entry>
            {
                TypeEntry("Bug"),
                TypeEntry("Task"),
                TypeEntry("Story"),
                TypeEntry("Epic"),
                TypeEntry("Sub-task"),
                TypeEntry("Improvement"),
                TypeEntry("New Feature"),
                TypeEntry("Test"),
                TypeEntry("Incident"),
                TypeEntry("Service Request"),
            };
        }

        // ---------- Priority ----------

        public static string PriorityGlyph(string p)
        {
            if (string.IsNullOrEmpty(p)) return "·";
            switch (p.Trim())
            {
                case "Highest":
                case "Critical":
                case "Blocker":  return "⇈";
                case "High":
                case "Major":    return "▲";
                case "Medium":   return "─";
                case "Low":
                case "Minor":    return "▼";
                case "Lowest":
                case "Trivial":  return "⇊";
                default: return "·";
            }
        }

        public static Color? PriorityColor(string glyph)
        {
            switch (glyph)
            {
                case "⇈": return Terminal.Gui.Color.BrightRed;
                case "▲": return Terminal.Gui.Color.Red;
                case "─": return null;
                case "▼": return Terminal.Gui.Color.BrightCyan;
                case "⇊": return Terminal.Gui.Color.DarkGray;
                default:  return null;
            }
        }

        public static IList<Entry> AllPriorities()
        {
            return new List<Entry>
            {
                new Entry { Name = "Highest / Blocker", Glyph = "⇈", Color = Terminal.Gui.Color.BrightRed },
                new Entry { Name = "High / Major",      Glyph = "▲", Color = Terminal.Gui.Color.Red },
                new Entry { Name = "Medium",            Glyph = "─", Color = null },
                new Entry { Name = "Low / Minor",       Glyph = "▼", Color = Terminal.Gui.Color.BrightCyan },
                new Entry { Name = "Lowest / Trivial",  Glyph = "⇊", Color = Terminal.Gui.Color.DarkGray },
            };
        }

        // ---------- Status ----------
        //
        // Jira lets administrators define arbitrary status names, so we match on
        // common patterns by category rather than exact strings. Order matters —
        // first match wins.

        public static string StatusGlyph(string s)
        {
            var e = StatusEntryFor(s);
            return e != null ? e.Glyph : "?";
        }

        // Status glyphs intentionally have no colour override — some themes paint
        // the row in colours that clash badly with a hard-coded foreground, so we
        // let status inherit the row's normal foreground and rely on the glyph
        // shape alone for at-a-glance recognition.

        private static Entry StatusEntryFor(string s)
        {
            if (string.IsNullOrEmpty(s)) return null;
            var u = s.Trim().ToUpperInvariant();

            if (u == "DONE" || u == "CLOSED" || u == "RESOLVED" || u == "COMPLETE"
                || u == "COMPLETED" || u.Contains("DEPLOYED") || u == "FIXED")
                return new Entry { Name = "Done / Closed / Resolved", Glyph = "✓" };

            if (u.Contains("CANCEL") || u.Contains("WON'T") || u.Contains("WONT")
                || u == "REJECTED" || u == "DUPLICATE")
                return new Entry { Name = "Cancelled / Won't do", Glyph = "⊘" };

            if (u.Contains("BLOCK") || u.Contains("HOLD") || u == "WAITING" || u.Contains("STALLED"))
                return new Entry { Name = "Blocked / On Hold", Glyph = "✕" };

            if (u.Contains("REVIEW") || u == "QA" || u.Contains("TEST") || u.Contains("VERIFY")
                || u.Contains("CODE REVIEW"))
                return new Entry { Name = "In Review / Testing", Glyph = "◑" };

            if (u.Contains("PROGRESS") || u == "DOING" || u == "DEVELOPING" || u.Contains("WIP"))
                return new Entry { Name = "In Progress", Glyph = "◐" };

            // To do / backlog / open / new — empty circle, fills as work progresses.
            if (u == "TO DO" || u == "TODO" || u == "OPEN" || u == "NEW"
                || u == "BACKLOG" || u == "SELECTED FOR DEVELOPMENT" || u == "READY")
                return new Entry { Name = "To Do / Open / Backlog", Glyph = "○" };

            return new Entry { Name = "Other", Glyph = "?" };
        }

        public static IList<Entry> AllStatuses()
        {
            // Canonical examples, deduplicated. The matcher above maps many
            // synonyms to the same glyph; here we list the buckets.
            return new List<Entry>
            {
                new Entry { Name = "To Do / Open / Backlog",     Glyph = "○" },
                new Entry { Name = "In Progress",                Glyph = "◐" },
                new Entry { Name = "In Review / Testing / QA",   Glyph = "◑" },
                new Entry { Name = "Blocked / On Hold",          Glyph = "✕" },
                new Entry { Name = "Done / Closed / Resolved",   Glyph = "✓" },
                new Entry { Name = "Cancelled / Won't do",       Glyph = "⊘" },
                new Entry { Name = "Other / unrecognised",       Glyph = "?" },
            };
        }
    }
}
