using System;
using System.Linq;
using JiraTUI.Config;
using JiraTUI.Themes;
using NStack;
using Terminal.Gui;

namespace JiraTUI.UI
{
    public class SettingsDialog : Dialog
    {
        public bool Saved { get; private set; }
        public AppConfig Result { get; private set; }

        private readonly AppConfig _working;
        private readonly string _originalThemeName;

        private TextField _baseUrl;
        private TextField _email;
        private TextField _token;
        private RadioGroup _authType;
        private ComboBox _themeCombo;
        private TextField _defaultJql;
        private TextField _pageSize;
        private TextField _autoRefresh;
        private TextField _aiProvider;
        private TextField _aiModel;
        private TextField _aiApiKey;

        public SettingsDialog(AppConfig current) : base("Settings", 76, 22)
        {
            _working = Clone(current);
            _originalThemeName = ThemeManager.CurrentThemeName;

            var tabs = new TabView
            {
                X = 0,
                Y = 0,
                Width = Dim.Fill(),
                Height = Dim.Fill() - 2,
            };

            tabs.AddTab(new TabView.Tab("Connection", BuildConnectionView()), true);
            tabs.AddTab(new TabView.Tab("Appearance", BuildAppearanceView()), false);
            tabs.AddTab(new TabView.Tab("Behavior",   BuildBehaviorView()), false);
            tabs.AddTab(new TabView.Tab("AI",         BuildAiView()), false);

            Add(tabs);

            var save = new Button("Save", true);
            save.Clicked += () =>
            {
                if (TryCommit(out var err))
                {
                    Saved = true;
                    Result = _working;
                    Application.RequestStop();
                }
                else
                {
                    MessageBox.ErrorQuery("Invalid settings", err, "OK");
                }
            };

            var cancel = new Button("Cancel");
            cancel.Clicked += () =>
            {
                // Revert any live-preview theme change.
                if (!string.Equals(ThemeManager.CurrentThemeName, _originalThemeName, StringComparison.OrdinalIgnoreCase))
                {
                    ThemeManager.Apply(_originalThemeName);
                    Application.Refresh();
                }
                Saved = false;
                Application.RequestStop();
            };

            var test = new Button("Test connection");
            test.Clicked += () =>
            {
                // Build a transient config from the in-dialog values, without
                // mutating _working or saving anything to disk.
                var probe = Clone(_working);
                probe.Connection.BaseUrl = CleanField(_baseUrl.Text);
                probe.Connection.Email = CleanField(_email.Text);
                var plain = CleanField(_token.Text);
                probe.Connection.TokenProtected =
                    string.IsNullOrEmpty(plain) ? "" : SecretProtector.Protect(plain);
                probe.Connection.AuthType = (AuthType)_authType.SelectedItem;

                if (!Jira.JiraClientFactory.HasCreds(probe))
                {
                    MessageBox.ErrorQuery("Test connection",
                        "Wypełnij Base URL, email i token.", "OK");
                    return;
                }

                using (var client = Jira.JiraClientFactory.Create(probe))
                {
                    if (client.TestConnection(out var err))
                    {
                        MessageBox.Query("Test connection",
                            "Połączenie OK.\r\nUżytkownik: " + client.CurrentUserDisplay,
                            "OK");
                    }
                    else
                    {
                        MessageBox.ErrorQuery("Test connection",
                            "Nieudane:\r\n" + err, "OK");
                    }
                }
            };

            AddButton(test);
            AddButton(cancel);
            AddButton(save);
        }

        private View BuildConnectionView()
        {
            var v = new View { X = 0, Y = 0, Width = Dim.Fill(), Height = Dim.Fill() };

            v.Add(new Label("Base URL:") { X = 1, Y = 1 });
            _baseUrl = new TextField(_working.Connection.BaseUrl ?? "")
                { X = 15, Y = 1, Width = Dim.Fill() - 2 };
            v.Add(_baseUrl);

            v.Add(new Label("Auth type:") { X = 1, Y = 3 });
            _authType = new RadioGroup(new ustring[] { "Basic", "API Token", "PAT" })
                { X = 15, Y = 3 };
            _authType.SelectedItem = (int)_working.Connection.AuthType;
            v.Add(_authType);

            v.Add(new Label("Email / User:") { X = 1, Y = 7 });
            _email = new TextField(_working.Connection.Email ?? "")
                { X = 15, Y = 7, Width = Dim.Fill() - 2 };
            v.Add(_email);

            v.Add(new Label("Token:") { X = 1, Y = 9 });
            _token = new TextField(SecretProtector.Unprotect(_working.Connection.TokenProtected))
            {
                X = 15,
                Y = 9,
                Width = Dim.Fill() - 2,
                Secret = true,
            };
            v.Add(_token);

            // Live character counter — paste through TextField key-events can
            // drop bytes; this is how the user sees it the moment it happens.
            var lenLabel = new Label(TokenLenText()) { X = 15, Y = 10, Width = 14 };
            v.Add(lenLabel);
            _token.TextChanged += _ => { lenLabel.Text = TokenLenText(); lenLabel.SetNeedsDisplay(); };

            // Toggle plaintext so a wrong paste is easy to spot/fix.
            var showCb = new CheckBox("Show") { X = 29, Y = 10 };
            showCb.Toggled += _ =>
            {
                _token.Secret = !showCb.Checked;
                _token.SetNeedsDisplay();
            };
            v.Add(showCb);

            // Bypass TextField's per-key paste path: shove the whole clipboard in at once.
            var pasteBtn = new Button("Paste from clipboard") { X = 38, Y = 10 };
            pasteBtn.Clicked += () =>
            {
                try
                {
                    var clip = Clipboard.Contents != null ? Clipboard.Contents.ToString() : "";
                    clip = TrimToken(clip);
                    _token.Text = clip;
                    _token.CursorPosition = clip.Length;
                    lenLabel.Text = TokenLenText();
                    lenLabel.SetNeedsDisplay();
                }
                catch (Exception ex)
                {
                    MessageBox.ErrorQuery("Clipboard error", ex.Message, "OK");
                }
            };
            v.Add(pasteBtn);

            v.Add(new Label("Token is stored encrypted (DPAPI, current user).")
                { X = 15, Y = 12 });

            return v;
        }

        private string TokenLenText()
        {
            int len = (_token != null && _token.Text != null) ? _token.Text.Length : 0;
            return "(" + len + " chars)";
        }

        private static string TrimToken(string s)
        {
            if (string.IsNullOrEmpty(s)) return "";
            return s.Trim(' ', '\t', '\r', '\n', ' ', '​', '﻿');
        }

        private View BuildAppearanceView()
        {
            var v = new View { X = 0, Y = 0, Width = Dim.Fill(), Height = Dim.Fill() };

            v.Add(new Label("Theme:") { X = 1, Y = 1 });
            _themeCombo = new ComboBox
            {
                X = 15,
                Y = 1,
                Width = 30,
                Height = 6,
                ReadOnly = true,
            };
            var names = ThemeManager.AvailableThemes.ToList();
            _themeCombo.SetSource(names);
            int idx = names.FindIndex(n => string.Equals(n, _working.Appearance.ThemeName, StringComparison.OrdinalIgnoreCase));
            if (idx < 0) idx = 0;
            _themeCombo.SelectedItem = idx;

            _themeCombo.SelectedItemChanged += (args) =>
            {
                int i = _themeCombo.SelectedItem;
                if (i < 0 || i >= names.Count) return;
                ThemeManager.Apply(names[i]);
                Application.Refresh();
            };
            v.Add(_themeCombo);

            v.Add(new Label("Podgląd na żywo. Cancel cofnie zmianę.")
                { X = 1, Y = 8 });

            return v;
        }

        private View BuildBehaviorView()
        {
            var v = new View { X = 0, Y = 0, Width = Dim.Fill(), Height = Dim.Fill() };

            v.Add(new Label("Default JQL:") { X = 1, Y = 1 });
            _defaultJql = new TextField(_working.Behavior.DefaultJql ?? "")
                { X = 18, Y = 1, Width = Dim.Fill() - 2 };
            v.Add(_defaultJql);

            v.Add(new Label("Page size:") { X = 1, Y = 3 });
            _pageSize = new TextField(_working.Behavior.PageSize.ToString())
                { X = 18, Y = 3, Width = 8 };
            v.Add(_pageSize);

            v.Add(new Label("Auto-refresh (s):") { X = 1, Y = 5 });
            _autoRefresh = new TextField(_working.Behavior.AutoRefreshSeconds.ToString())
                { X = 18, Y = 5, Width = 8 };
            v.Add(_autoRefresh);

            v.Add(new Label("0 = off") { X = 28, Y = 5 });

            return v;
        }

        private View BuildAiView()
        {
            var v = new View { X = 0, Y = 0, Width = Dim.Fill(), Height = Dim.Fill() };

            v.Add(new Label("Provider:") { X = 1, Y = 1 });
            _aiProvider = new TextField(_working.Ai.Provider ?? "Anthropic")
                { X = 15, Y = 1, Width = Dim.Fill() - 2 };
            v.Add(_aiProvider);

            v.Add(new Label("Model:") { X = 1, Y = 3 });
            _aiModel = new TextField(_working.Ai.Model ?? "claude-sonnet-4-5")
                { X = 15, Y = 3, Width = Dim.Fill() - 2 };
            v.Add(_aiModel);

            v.Add(new Label("API Key:") { X = 1, Y = 5 });
            _aiApiKey = new TextField(SecretProtector.Unprotect(_working.Ai.ApiKeyProtected))
            {
                X = 15, Y = 5, Width = Dim.Fill() - 2, Secret = true,
            };
            v.Add(_aiApiKey);

            var lenLabel = new Label(AiKeyLenText()) { X = 15, Y = 6, Width = 14 };
            v.Add(lenLabel);
            _aiApiKey.TextChanged += _ => { lenLabel.Text = AiKeyLenText(); lenLabel.SetNeedsDisplay(); };

            var showCb = new CheckBox("Show") { X = 29, Y = 6 };
            showCb.Toggled += _ =>
            {
                _aiApiKey.Secret = !showCb.Checked;
                _aiApiKey.SetNeedsDisplay();
            };
            v.Add(showCb);

            var pasteBtn = new Button("Paste from clipboard") { X = 38, Y = 6 };
            pasteBtn.Clicked += () =>
            {
                try
                {
                    var clip = Clipboard.Contents != null ? Clipboard.Contents.ToString() : "";
                    clip = TrimToken(clip);
                    _aiApiKey.Text = clip;
                    _aiApiKey.CursorPosition = clip.Length;
                    lenLabel.Text = AiKeyLenText();
                    lenLabel.SetNeedsDisplay();
                }
                catch (Exception ex)
                {
                    MessageBox.ErrorQuery("Clipboard error", ex.Message, "OK");
                }
            };
            v.Add(pasteBtn);

            v.Add(new Label("Anthropic key z console.anthropic.com → Settings → API Keys.") { X = 1, Y = 8 });
            v.Add(new Label("Klucz przechowywany szyfrowany (DPAPI, current user).") { X = 1, Y = 9 });
            v.Add(new Label("Skrót w JQL barze: Ctrl-G otwiera dialog generowania.") { X = 1, Y = 10 });

            return v;
        }

        private string AiKeyLenText()
        {
            int len = (_aiApiKey != null && _aiApiKey.Text != null) ? _aiApiKey.Text.Length : 0;
            return "(" + len + " chars)";
        }

        private bool TryCommit(out string error)
        {
            error = null;

            _working.Connection.BaseUrl = CleanField(_baseUrl.Text);
            _working.Connection.Email = CleanField(_email.Text);
            var plain = CleanField(_token.Text);
            _working.Connection.TokenProtected =
                string.IsNullOrEmpty(plain) ? "" : SecretProtector.Protect(plain);
            _working.Connection.AuthType = (AuthType)_authType.SelectedItem;

            var names = ThemeManager.AvailableThemes.ToList();
            int ti = _themeCombo.SelectedItem;
            if (ti >= 0 && ti < names.Count)
                _working.Appearance.ThemeName = names[ti];

            _working.Behavior.DefaultJql = (_defaultJql.Text ?? "").ToString();

            if (!int.TryParse((_pageSize.Text ?? "").ToString(), out int ps) || ps <= 0 || ps > 1000)
            {
                error = "Page size must be a number between 1 and 1000.";
                return false;
            }
            _working.Behavior.PageSize = ps;

            if (!int.TryParse((_autoRefresh.Text ?? "").ToString(), out int ar) || ar < 0)
            {
                error = "Auto-refresh must be 0 or a positive number of seconds.";
                return false;
            }
            _working.Behavior.AutoRefreshSeconds = ar;

            _working.Ai.Provider = CleanField(_aiProvider.Text);
            _working.Ai.Model = CleanField(_aiModel.Text);
            var aiKey = CleanField(_aiApiKey.Text);
            _working.Ai.ApiKeyProtected = string.IsNullOrEmpty(aiKey)
                ? "" : SecretProtector.Protect(aiKey);

            return true;
        }

        private static AppConfig Clone(AppConfig c) =>
            Newtonsoft.Json.JsonConvert.DeserializeObject<AppConfig>(
                Newtonsoft.Json.JsonConvert.SerializeObject(c));

        /// <summary>
        /// Strip every flavour of whitespace and invisible character that can sneak
        /// in via clipboard paste (BOM, NBSP, zero-width space, CR/LF).
        /// </summary>
        private static string CleanField(NStack.ustring text)
        {
            if (text == null) return "";
            var s = text.ToString();
            if (string.IsNullOrEmpty(s)) return "";
            return s.Trim(' ', '\t', '\r', '\n', ' ', '​', '﻿');
        }
    }
}
