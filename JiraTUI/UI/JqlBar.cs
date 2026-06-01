using System;
using JiraTUI.Config;
using Terminal.Gui;

namespace JiraTUI.UI
{
    public class JqlBar : FrameView
    {
        private readonly TextField _input;
        private JqlHistory _history;
        private int _historyIndex = -1;
        private JqlHistoryEntry _recalled;
        private bool _suppressChange;

        public event Action<string> Submitted;

        /// <summary>
        /// The history entry currently displayed via ↑/↓ recall, or null if the
        /// user is typing freely. MainWindow reads this after Submitted fires to
        /// decide whether to substitute the AI-translated effective JQL.
        /// </summary>
        public JqlHistoryEntry RecalledEntry => _recalled;

        public string Jql
        {
            get => _input.Text != null ? _input.Text.ToString() : "";
            set
            {
                _suppressChange = true;
                _input.Text = value ?? "";
                _input.CursorPosition = (value ?? "").Length;
                _suppressChange = false;
            }
        }

        public JqlBar() : base("JQL  (Enter to run · ↑↓ history · Ctrl-G AI)")
        {
            X = 0;
            Width = Dim.Fill();
            Height = 3;

            _input = new TextField("")
            {
                X = 0,
                Y = 0,
                Width = Dim.Fill(),
                Height = 1,
            };

            _input.TextChanged += _ =>
            {
                // Programmatic updates (recall, MainWindow assigning Jql) flip this
                // flag so they don't break the recall association. Real user typing
                // doesn't, so it breaks the association and a subsequent Enter is
                // treated as a fresh submission.
                if (_suppressChange) return;
                _recalled = null;
                _historyIndex = -1;
            };

            _input.KeyDown += e =>
            {
                var k = e.KeyEvent.Key;
                if (k == Key.Enter)
                {
                    var h = Submitted;
                    if (h != null) h(Jql);
                    _historyIndex = -1;
                    // NB: _recalled deliberately left as-is until MainWindow reads it.
                    // It'll be cleared on next text change.
                    e.Handled = true;
                }
                else if (k == Key.CursorUp)
                {
                    BrowseHistory(+1);
                    e.Handled = true;
                }
                else if (k == Key.CursorDown)
                {
                    BrowseHistory(-1);
                    e.Handled = true;
                }
            };

            Add(_input);
        }

        public void SetHistory(JqlHistory history) => _history = history;

        public void FocusInput() => _input.SetFocus();

        /// <summary>
        /// Select every character in the input. Typing then replaces the whole
        /// query, as users expect after re-opening the bar with Ctrl-J.
        /// </summary>
        public void SelectAll()
        {
            int len = _input.Text != null ? _input.Text.Length : 0;
            if (len <= 0) return;
            _input.SelectedStart = 0;
            _input.CursorPosition = len;
        }

        private void BrowseHistory(int delta)
        {
            if (_history == null || _history.Count == 0) return;

            int newIdx = _historyIndex + delta;
            if (newIdx < 0)
            {
                // walked past the most-recent entry — back to a clean bar
                _historyIndex = -1;
                _recalled = null;
                SetTextSilently("");
                return;
            }

            var entry = _history.GetByRecentIndex(newIdx);
            if (entry == null) return; // walked past the oldest entry; stay put

            _historyIndex = newIdx;
            _recalled = entry;
            SetTextSilently(entry.OriginalText ?? "");
        }

        private void SetTextSilently(string text)
        {
            _suppressChange = true;
            _input.Text = text ?? "";
            _input.CursorPosition = (text ?? "").Length;
            _suppressChange = false;
        }
    }
}
