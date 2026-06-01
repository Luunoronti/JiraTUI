using System;
using JiraTUI.Ai;
using JiraTUI.Jira;
using Terminal.Gui;

namespace JiraTUI.UI.Dialogs
{
    /// <summary>
    /// Modal that takes a natural-language prompt, sends it to Claude with Jira
    /// context, and offers the resulting JQL for review / edit / accept.
    /// </summary>
    public class AiJqlDialog : Dialog
    {
        public bool Accepted { get; private set; }
        public string GeneratedJql { get; private set; }

        private readonly AiClient _ai;
        private readonly IJiraClient _jira;
        private readonly TextView _prompt;
        private readonly TextView _result;
        private readonly Label _status;
        private readonly Button _generateBtn;

        public AiJqlDialog(AiClient ai, IJiraClient jira, string startingJql)
            : base("Ask AI for JQL", ComputeWidth(), ComputeHeight())
        {
            _ai = ai;
            _jira = jira;

            Add(new Label("Prompt (Polish or English, multi-line):") { X = 1, Y = 0 });
            _prompt = new TextView
            {
                X = 1, Y = 1,
                Width = Dim.Fill() - 2,
                Height = 6,
                WordWrap = true,
            };
            Add(_prompt);

            _generateBtn = new Button("Generate (Ctrl-Enter)") { X = 1, Y = 8 };
            _generateBtn.Clicked += GenerateClick;
            Add(_generateBtn);

            _status = new Label("idle")
            {
                X = Pos.Right(_generateBtn) + 2,
                Y = 8,
                Width = Dim.Fill() - 2,
            };
            Add(_status);

            Add(new Label("Generated JQL (editable — fine-tune before Use):") { X = 1, Y = 10 });
            _result = new TextView
            {
                X = 1, Y = 11,
                Width = Dim.Fill() - 2,
                Height = Dim.Fill() - 13,
                WordWrap = true,
                Text = startingJql ?? "",
            };
            Add(_result);

            var cancel = new Button("Cancel");
            cancel.Clicked += () => { Accepted = false; Application.RequestStop(); };
            var use = new Button("Use");
            use.Clicked += () =>
            {
                GeneratedJql = (_result.Text != null ? _result.Text.ToString() : "").Trim();
                Accepted = true;
                Application.RequestStop();
            };
            AddButton(cancel);
            AddButton(use);

            // Ctrl-Enter generates (KeyDown bubbles top-down, captured here before children)
            KeyDown += args =>
            {
                if (args.KeyEvent.Key == (Key.CtrlMask | Key.Enter) ||
                    args.KeyEvent.Key == (Key.CtrlMask | Key.J))   // some terminals map Ctrl-Enter → Ctrl-J
                {
                    GenerateClick();
                    args.Handled = true;
                }
            };

            Loaded += () => _prompt.SetFocus();
        }

        private void GenerateClick()
        {
            var promptText = _prompt.Text != null ? _prompt.Text.ToString().Trim() : "";
            if (string.IsNullOrEmpty(promptText))
            {
                MessageBox.ErrorQuery("Prompt required", "Wpisz prompt w górnym polu.", "OK");
                return;
            }

            // Visual feedback before the sync call freezes the UI.
            _generateBtn.Text = "Generating…";
            _generateBtn.Enabled = false;
            _status.Text = "calling Anthropic…";
            _generateBtn.SetNeedsDisplay();
            _status.SetNeedsDisplay();
            Application.Refresh();

            try
            {
                var systemPrompt = JqlPromptBuilder.BuildSystemPrompt(_jira);
                var raw = _ai.Generate(systemPrompt, promptText);
                var clean = AiClient.StripMarkdownFences(raw);
                _result.Text = clean;
                _result.TopRow = 0;
                _status.Text = "ok";
            }
            catch (Exception ex)
            {
                _status.Text = "ERROR";
                MessageBox.ErrorQuery("AI error", ex.Message, "OK");
            }
            finally
            {
                _generateBtn.Text = "Generate (Ctrl-Enter)";
                _generateBtn.Enabled = true;
            }
        }

        private static int ComputeWidth()
        {
            int cols = Application.Driver != null ? Application.Driver.Cols : 100;
            return System.Math.Min(System.Math.Max(80, cols - 6), 110);
        }

        private static int ComputeHeight()
        {
            int rows = Application.Driver != null ? Application.Driver.Rows : 30;
            return System.Math.Min(System.Math.Max(22, rows - 4), 30);
        }
    }
}
