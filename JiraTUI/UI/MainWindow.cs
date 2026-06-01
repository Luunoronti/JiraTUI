using System;
using System.Linq;
using JiraTUI.Config;
using JiraTUI.Jira;
using JiraTUI.Themes;
using JiraTUI.UI.Dialogs;
using Terminal.Gui;

namespace JiraTUI.UI
{
    public class MainWindow : Toplevel
    {
        private IJiraClient _jira;
        private AppConfig _config;

        private Button _navToggle;
        private MenuBar _menu;
        private NavigationView _nav;
        private IssueListView _issueList;
        private IssueDetailView _detail;
        private JqlBar _jqlBar;
        private StatusBar _status;
        private StatusItem _pathItem;

        private bool _navVisible = false;
        private bool _jqlVisible = false;
        private bool _detailVisible = true;
        private string _currentPath = "(no selection)";
        private JqlHistory _history;

        private const int NavToggleWidth = 9; // "[ ☰ Nav ]"

        public MainWindow(IJiraClient jira, AppConfig config)
        {
            _jira = jira;
            _config = config;
            _history = JqlHistory.Load();

            X = 0; Y = 0;
            Width = Dim.Fill();
            Height = Dim.Fill();

            BuildTopBar();
            BuildBody();
            BuildStatus();
            WireEvents();
            ApplyPaneLayout();
            ApplyJqlLayout();

            _jqlBar.SetHistory(_history);

            Application.MainLoop.AddIdle(() => { InitialLoad(); return false; });
        }

        private void BuildTopBar()
        {
            _navToggle = new Button("☰ Nav")
            {
                X = 0,
                Y = 0,
                Width = NavToggleWidth,
                Height = 1,
                CanFocus = true,
            };
            _navToggle.Clicked += ToggleNav;

            _menu = new MenuBar(new MenuBarItem[]
            {
                new MenuBarItem("_File", new MenuItem[]
                {
                    // No underscore on "Settings" — leaving it on S would steal Alt-S
                    // globally, which collides with the Set Status action in Issue menu.
                    new MenuItem("Settings...",  "F2",     OpenSettings,                        null, null, Key.F2),
                    new MenuItem("_Refresh",     "F5",     Refresh,                              null, null, Key.F5),
                    null,
                    new MenuItem("_Quit",        "Ctrl-Q", () => Application.RequestStop(),     null, null, Key.CtrlMask | Key.Q),
                }),
                new MenuBarItem("_View", new MenuItem[]
                {
                    new MenuItem("Toggle _Navigation", "Ctrl-B", ToggleNav,                     null, null, Key.CtrlMask | Key.B),
                    new MenuItem("Toggle _Details",    "Ctrl-D", ToggleDetail,                  null, null, Key.CtrlMask | Key.D),
                    new MenuItem("_Columns...",        "",       OpenColumnsDialog),
                    new MenuItem("_Legend...",         "Ctrl-L", OpenLegendDialog,              null, null, Key.CtrlMask | Key.L),
                    new MenuItem("Focus _Issues",      "Alt+2",  () => _issueList.FocusTable(), null, null, Key.AltMask | Key.D2),
                    new MenuItem("Focus De_tails",     "Alt+3",  () => _detail.SetFocus(),      null, null, Key.AltMask | Key.D3),
                    new MenuItem("Focus _JQL",         "Alt+J",  () => _jqlBar.FocusInput(),    null, null, Key.AltMask | Key.J),
                    null,
                    BuildThemeMenu(),
                }),
                new MenuBarItem("_JQL", new MenuItem[]
                {
                    new MenuItem("Toggle JQL _bar",             "Ctrl-J", ToggleJql,           null, null, Key.CtrlMask | Key.J),
                    new MenuItem("Ask AI to _generate...",      "Ctrl-G", AskAiForJql,         null, null, Key.CtrlMask | Key.G),
                    // Accelerator on "L" (fi_lter), not "S" — Alt+S is reserved for
                    // Set Status in the Issue menu, which the user uses more often.
                    new MenuItem("Save current as fi_lter...",  "",       SaveCurrentAsFilter),
                }),
                new MenuBarItem("_Issue", new MenuItem[]
                {
                    new MenuItem("_Open in browser",        "Ctrl-O", OpenInBrowser, null, null, Key.CtrlMask | Key.O),
                    null,
                    new MenuItem("Set _Priority...",        "",       ChangePriority),
                    new MenuItem("Set _Status...",          "",       TransitionIssue),
                    new MenuItem("Set _Assignee...",        "",       ChangeAssignee),
                    new MenuItem("Edit _Description...",    "",       EditDescription),
                    new MenuItem("Add _Comment...",         "",       AddComment),
                    null,
                    new MenuItem("_New issue...",           "",       () => NotImplemented("New issue")),
                }),
                new MenuBarItem("_Help", new MenuItem[]
                {
                    new MenuItem("_About", "", ShowAbout),
                }),
            })
            {
                X = NavToggleWidth,
                Y = 0,
                Width = Dim.Fill(),
            };

            Add(_navToggle, _menu);
        }

        private MenuBarItem BuildThemeMenu()
        {
            var items = new System.Collections.Generic.List<MenuItem>();
            foreach (var name in ThemeManager.AvailableThemes)
            {
                var captured = name;
                items.Add(new MenuItem("Theme: " + captured, "", () => SwitchTheme(captured)));
            }
            return new MenuBarItem("_Theme", items.ToArray());
        }

        private void BuildBody()
        {
            _nav = new NavigationView
            {
                X = 0,
                Y = 1,
                Width = Dim.Percent(22),
                Height = Dim.Fill() - 4,
            };

            _issueList = new IssueListView
            {
                X = 0,
                Y = 1,
                Width = Dim.Percent(55),
                Height = Dim.Fill() - 4,
            };

            _detail = new IssueDetailView
            {
                X = Pos.Right(_issueList),
                Y = 1,
                Width = Dim.Fill(),
                Height = Dim.Fill() - 4,
            };

            _jqlBar = new JqlBar
            {
                X = 0,
                Y = Pos.AnchorEnd(4),
                Width = Dim.Fill(),
                Height = 3,
            };
            _jqlBar.Jql = _config.Behavior.DefaultJql;

            Add(_nav, _issueList, _detail, _jqlBar);
        }

        private void BuildStatus()
        {
            _pathItem = new StatusItem(Key.Null, _currentPath, null);

            _status = new StatusBar(new StatusItem[]
            {
                _pathItem,
                new StatusItem(Key.CtrlMask | Key.B, "~Ctrl-B~ Nav",      ToggleNav),
                new StatusItem(Key.CtrlMask | Key.D, "~Ctrl-D~ Details",  ToggleDetail),
                new StatusItem(Key.CtrlMask | Key.J, "~Ctrl-J~ JQL",      ToggleJql),
                new StatusItem(Key.F2,               "~F2~ Settings",     OpenSettings),
                new StatusItem(Key.F5,               "~F5~ Refresh",      Refresh),
                new StatusItem(Key.CtrlMask | Key.Q, "~Ctrl-Q~ Quit",     () => Application.RequestStop()),
            });
            Add(_status);
        }

        private void WireEvents()
        {
            _nav.NavSelected += (path, jql) =>
            {
                SetPath(path);
                if (!string.IsNullOrEmpty(jql))
                {
                    _jqlBar.Jql = jql;
                    RunSearch(jql);
                }
            };

            // Description + comments are already populated by SearchIssues, so
            // navigating between rows is just a local render — no network call.
            _issueList.SelectionChanged += issue => _detail.ShowIssue(issue);

            _jqlBar.Submitted += jql =>
            {
                var recalled = _jqlBar.RecalledEntry;
                if (recalled != null
                    && recalled.WasAiTranslated
                    && string.Equals(jql, recalled.OriginalText, StringComparison.Ordinal))
                {
                    // User pulled an AI-translated entry from history and pressed Enter
                    // without editing — paste the effective JQL into the bar and run
                    // it directly, no need to round-trip through Claude again.
                    _jqlBar.Jql = recalled.EffectiveJql;
                    SetPath("AI: " + Truncate(recalled.OriginalText, 50));
                    RunSearchDirect(recalled.EffectiveJql, recalled.OriginalText, wasAi: true);
                }
                else
                {
                    SetPath("Custom JQL");
                    RunSearch(jql, addToHistory: true);
                }
            };
        }

        private void InitialLoad()
        {
            // Apply persisted column visibility before populating any data.
            _issueList.SetColumns(_config.Columns);

            try
            {
                _nav.Populate(_jira.GetProjects(), _jira.GetSavedFilters());
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Could not load Jira metadata",
                    ex.Message + "\r\n\r\nSprawdź ustawienia (F2) — Base URL, email, token.",
                    "OK");
                _nav.Populate(new System.Collections.Generic.List<Jira.Models.Project>(),
                              new System.Collections.Generic.List<Jira.Models.SavedFilter>());
            }
            RunSearch(_jqlBar.Jql);
            _issueList.FocusTable();
        }

        private void RunSearch(string jql) => RunSearch(jql, addToHistory: false);

        private void RunSearch(string jql, bool addToHistory)
        {
            try
            {
                var results = _jira.SearchIssues(jql, _config.Behavior.PageSize);
                _issueList.SetIssues(results);
                if (addToHistory) _history.Add(jql, jql, wasAiTranslated: false);
                return;
            }
            catch (Exception ex)
            {
                // If the query failed because Jira couldn't parse it (HTTP 400) and we
                // have AI configured, try interpreting it as natural language and rerun.
                if (!string.IsNullOrWhiteSpace(jql) && IsBadJqlError(ex) && IsAiConfigured())
                {
                    if (TryAiInterpretAndRun(jql, addToHistory, out var aiErr)) return;
                    MessageBox.ErrorQuery("Search failed",
                        "Jira: " + ex.Message +
                        "\r\n\r\nAI fallback also failed: " + aiErr,
                        "OK");
                }
                else
                {
                    MessageBox.ErrorQuery("Search failed", ex.Message, "OK");
                }
            }
        }

        /// <summary>
        /// Run an already-resolved JQL straight against Jira, no AI fallback. Used
        /// when the user recalls an AI-translated history entry — we already know
        /// the effective JQL, so no need to ask Claude again.
        /// </summary>
        private void RunSearchDirect(string effectiveJql, string originalText, bool wasAi)
        {
            try
            {
                var results = _jira.SearchIssues(effectiveJql, _config.Behavior.PageSize);
                _issueList.SetIssues(results);
                _history.Add(originalText, effectiveJql, wasAi);
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Search failed", ex.Message, "OK");
            }
        }

        private static bool IsBadJqlError(Exception ex)
        {
            var msg = ex?.Message ?? "";
            return msg.IndexOf("HTTP 400", StringComparison.OrdinalIgnoreCase) >= 0;
        }

        private bool IsAiConfigured()
        {
            return !string.IsNullOrEmpty(_config?.Ai?.ApiKeyProtected);
        }

        /// <summary>
        /// Treat the user's input as a natural-language prompt, ask Claude to convert
        /// it into JQL, write the result into the bar and re-run the search.
        /// Returns false if the AI step itself fails (so the caller can show a combined
        /// error to the user). Never recurses — if the AI-generated JQL is also bad,
        /// we surface the error instead of calling AI a second time.
        /// </summary>
        private bool TryAiInterpretAndRun(string failedQuery, bool addToHistory, out string error)
        {
            error = null;
            try
            {
                SetPath("AI interpreting…");
                _status?.SetNeedsDisplay();
                Application.Refresh();

                var apiKey = Config.SecretProtector.Unprotect(_config.Ai.ApiKeyProtected);
                using (var ai = new Ai.AiClient(apiKey, _config.Ai.Model))
                {
                    var system = Ai.JqlPromptBuilder.BuildSystemPrompt(_jira);
                    var raw = ai.Generate(system, failedQuery);
                    var generated = Ai.AiClient.StripMarkdownFences(raw);

                    if (string.IsNullOrWhiteSpace(generated))
                    {
                        error = "AI returned an empty query.";
                        return false;
                    }

                    _jqlBar.Jql = generated;
                    SetPath("AI: " + Truncate(failedQuery, 50));

                    // Second attempt — do NOT recurse into AI fallback if this fails.
                    var results = _jira.SearchIssues(generated, _config.Behavior.PageSize);
                    _issueList.SetIssues(results);
                    if (addToHistory) _history.Add(failedQuery, generated, wasAiTranslated: true);
                    return true;
                }
            }
            catch (Exception ex)
            {
                error = ex.Message;
                return false;
            }
        }

        private static string Truncate(string s, int max)
        {
            if (string.IsNullOrEmpty(s)) return "";
            return s.Length <= max ? s : s.Substring(0, max - 1) + "…";
        }

        private void Refresh() => RunSearch(_jqlBar.Jql);

        private void ToggleNav()
        {
            _navVisible = !_navVisible;
            ApplyPaneLayout();
            if (_navVisible) _nav.FocusTree();
            else _issueList.FocusTable();
        }

        private void ToggleDetail()
        {
            _detailVisible = !_detailVisible;
            ApplyPaneLayout();
            // If focus was on the now-hidden detail panel, move it somewhere sensible.
            if (!_detailVisible) _issueList.FocusTable();
        }

        /// <summary>
        /// Recompute pane sizes from the nav/detail visibility flags. The issue
        /// list's width depends on whether both side panels are visible, only one,
        /// or neither, so it can't be set independently in either toggle.
        /// </summary>
        private void ApplyPaneLayout()
        {
            // === Navigation panel ===
            _nav.Visible = _navVisible;
            if (_navVisible)
            {
                _nav.X = 0;
                _nav.Width = Dim.Percent(22);
                _issueList.X = Pos.Right(_nav);
            }
            else
            {
                _nav.Width = 0;
                _issueList.X = 0;
            }

            // === Detail panel ===
            _detail.Visible = _detailVisible;
            if (_detailVisible)
            {
                _detail.X = Pos.Right(_issueList);
                _detail.Width = Dim.Fill();
            }
            else
            {
                _detail.Width = 0;
            }

            // === Issue list width ===
            // Detail visible: list takes a percentage so detail can fill the rest.
            // Detail hidden:  list fills everything to the right of nav.
            if (_detailVisible)
                _issueList.Width = _navVisible ? Dim.Percent(45) : Dim.Percent(55);
            else
                _issueList.Width = Dim.Fill();

            LayoutSubviews();
            SetNeedsDisplay();
        }

        private void ToggleJql()
        {
            _jqlVisible = !_jqlVisible;
            ApplyJqlLayout();
            if (_jqlVisible)
            {
                _jqlBar.FocusInput();
                // Select the whole query so the user can just start typing to
                // replace it instead of clearing manually first.
                _jqlBar.SelectAll();
            }
            else
            {
                _issueList.FocusTable();
            }
        }

        // Letters used by Polish AltGr layout to produce ąćęłńóśźż. None of
        // these overlap with our top-level menu accelerators (F V J I H), so
        // suppressing them while a text input is focused doesn't cost any menu
        // navigation but unblocks typing.
        private static readonly System.Collections.Generic.HashSet<char> PolishAltGrLetters =
            new System.Collections.Generic.HashSet<char> { 'A', 'C', 'E', 'L', 'N', 'O', 'S', 'X', 'Z' };

        /// <summary>
        /// Intercept AltGr-typed characters before the menu bar sees them.
        /// Different TG 1.x drivers report AltGr differently — sometimes as
        /// Ctrl+Alt, sometimes as Alt + Unicode codepoint, sometimes as bare
        /// Alt + ASCII letter with the produced char delivered separately.
        /// This filter catches all three cases when focus is on a text input.
        /// </summary>
        public override bool ProcessHotKey(KeyEvent keyEvent)
        {
            if (keyEvent.IsAlt)
            {
                // 1. Ctrl+Alt = AltGr on Windows (LCtrl + RAlt internally).
                if (keyEvent.IsCtrl) return false;

                // 2. Non-ASCII codepoint already resolved by the driver — the user
                //    typed a Polish letter, definitely not a menu accelerator.
                if (keyEvent.KeyValue > 127) return false;

                // 3. ASCII Alt+letter that matches a Polish AltGr key while the
                //    user is typing in a text field — bias toward typing.
                if (IsTextInputFocused())
                {
                    var bare = keyEvent.Key & ~(Key.AltMask | Key.CtrlMask | Key.ShiftMask);
                    int v = (int)bare;
                    char c = '\0';
                    if (v >= 'A' && v <= 'Z') c = (char)v;
                    else if (v >= 'a' && v <= 'z') c = (char)(v - 32);
                    if (c != '\0' && PolishAltGrLetters.Contains(c)) return false;
                }
            }
            return base.ProcessHotKey(keyEvent);
        }

        private bool IsTextInputFocused()
        {
            var v = MostFocused;
            return v is TextField || v is TextView;
        }

        private void ApplyJqlLayout()
        {
            // Reserve 4 rows at the bottom for JQL bar (3) + status bar (1) when
            // the bar is visible; only the status row when it's hidden.
            int reserved = _jqlVisible ? 4 : 1;

            _jqlBar.Visible = _jqlVisible;
            _jqlBar.Height = _jqlVisible ? 3 : 0;
            _nav.Height = Dim.Fill() - reserved;
            _issueList.Height = Dim.Fill() - reserved;
            _detail.Height = Dim.Fill() - reserved;

            LayoutSubviews();
            SetNeedsDisplay();
        }

        private void SetPath(string path)
        {
            _currentPath = string.IsNullOrEmpty(path) ? "(no selection)" : path;
            _pathItem.Title = _currentPath;
            _status.SetNeedsDisplay();
        }

        private void OpenSettings()
        {
            var dlg = new SettingsDialog(_config);
            Application.Run(dlg);
            if (dlg.Saved)
            {
                _config = dlg.Result;
                ConfigStore.Save(_config);
                ThemeManager.Apply(_config.Appearance.ThemeName);
                _detail?.RefreshColors();
                ForceFullRedraw();

                // Connection details might have changed — rebuild the Jira client
                // and reload everything (projects, filters, default search).
                RebuildJiraClient();
                _jqlBar.Jql = _config.Behavior.DefaultJql;
                InitialLoad();
            }
        }

        private void RebuildJiraClient()
        {
            try { _jira?.Dispose(); } catch { /* ignore */ }
            _jira = JiraClientFactory.Create(_config);
        }

        private void SwitchTheme(string name)
        {
            ThemeManager.Apply(name);
            _config.Appearance.ThemeName = ThemeManager.CurrentThemeName;
            ConfigStore.Save(_config);
            _detail?.RefreshColors();
            ForceFullRedraw();
        }

        private void ForceFullRedraw()
        {
            // Mark every view in the tree dirty so it repaints with the new attributes,
            // then ask Terminal.Gui to re-render the whole screen.
            InvalidateRecursive(this);
            Application.Refresh();
        }

        private static void InvalidateRecursive(View v)
        {
            if (v == null) return;
            v.SetNeedsDisplay();
            if (v.Subviews != null)
            {
                foreach (var sub in v.Subviews)
                    InvalidateRecursive(sub);
            }
        }

        private void OpenLegendDialog()
        {
            var dlg = new LegendDialog();
            Application.Run(dlg);
        }

        private void OpenColumnsDialog()
        {
            var dlg = new ColumnsDialog(_config.Columns ?? new Config.ColumnVisibilityConfig());
            Application.Run(dlg);
            if (dlg.Saved)
            {
                _config.Columns = dlg.Result;
                ConfigStore.Save(_config);
                _issueList.SetColumns(_config.Columns);
                // Repopulate with current results so cells reflect the new column layout.
                Refresh();
            }
        }

        private void ShowAbout()
        {
            MessageBox.Query("About JiraTUI",
                "JiraTUI — terminal UI for Jira\r\n" +
                ".NET Framework 4.8 + Terminal.Gui\r\n\r\n" +
                "Config: " + ConfigStore.ConfigFilePath,
                "OK");
        }

        private void NotImplemented(string what)
        {
            MessageBox.Query(what, "Funkcja jeszcze nie zaimplementowana.", "OK");
        }

        // =========================================================================
        // Issue actions
        // =========================================================================

        private void OpenInBrowser()
        {
            var issue = RequireSelectedIssue();
            if (issue == null) return;

            var baseUrl = (_config.Connection.BaseUrl ?? "").TrimEnd('/');
            if (string.IsNullOrEmpty(baseUrl) ||
                !(baseUrl.StartsWith("http://", StringComparison.OrdinalIgnoreCase) ||
                  baseUrl.StartsWith("https://", StringComparison.OrdinalIgnoreCase)))
            {
                MessageBox.Query("Open in browser",
                    "Brak skonfigurowanego Jira URL.\r\nUstaw Base URL w F2 → Connection.",
                    "OK");
                return;
            }

            var url = baseUrl + "/browse/" + Uri.EscapeDataString(issue.Key);
            try
            {
                System.Diagnostics.Process.Start(new System.Diagnostics.ProcessStartInfo
                {
                    FileName = url,
                    UseShellExecute = true,
                });
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Open in browser", ex.Message, "OK");
            }
        }

        private void ChangePriority()
        {
            var issue = RequireSelectedIssue();
            if (issue == null) return;

            try
            {
                var priorities = _jira.GetPriorityNames();
                if (priorities == null || priorities.Count == 0)
                {
                    MessageBox.ErrorQuery("Priority", "Brak dostępnych priorytetów.", "OK");
                    return;
                }

                int currentIdx = priorities.ToList().FindIndex(
                    p => string.Equals(p, issue.Priority, StringComparison.OrdinalIgnoreCase));
                if (currentIdx < 0) currentIdx = 0;

                var dlg = new ChoiceDialog("Set priority — " + issue.Key, priorities.ToList(), currentIdx);
                Application.Run(dlg);
                if (dlg.SelectedIndex < 0) return;

                var picked = priorities[dlg.SelectedIndex];
                if (string.Equals(picked, issue.Priority, StringComparison.OrdinalIgnoreCase)) return;

                _jira.SetPriority(issue.Key, picked);
                RefreshIssueAfterChange(issue.Key);
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Priority", ex.Message, "OK");
            }
        }

        private void TransitionIssue()
        {
            var issue = RequireSelectedIssue();
            if (issue == null) return;

            try
            {
                var transitions = _jira.GetTransitions(issue.Key);

                // Filter out self-loop transitions (destination == current status).
                // These are typically global "Any Status → X" transitions that Jira
                // returns even when the issue is already in state X.
                var available = transitions
                    .Where(t => !string.Equals(t.ToStatus, issue.Status,
                                               System.StringComparison.OrdinalIgnoreCase))
                    .ToList();

                if (available.Count == 0)
                {
                    MessageBox.ErrorQuery("Status", "Brak dostępnych przejść dla tego issue.", "OK");
                    return;
                }

                var labels = available.Select(t =>
                    (issue.Status ?? "?") + "  →  " + (t.ToStatus ?? t.Name)).ToList();

                var dlg = new ChoiceDialog(
                    "Transition — " + issue.Key + " (now: " + (issue.Status ?? "?") + ")",
                    labels, 0);
                Application.Run(dlg);
                if (dlg.SelectedIndex < 0) return;

                _jira.TransitionIssue(issue.Key, available[dlg.SelectedIndex].Id);
                RefreshIssueAfterChange(issue.Key);
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Status", ex.Message, "OK");
            }
        }

        private void ChangeAssignee()
        {
            var issue = RequireSelectedIssue();
            if (issue == null) return;

            try
            {
                var dlg = new AssigneeDialog(_jira, issue.Key, issue.Assignee);
                Application.Run(dlg);
                if (!dlg.Saved) return;

                // Empty string is our "unassign" sentinel; map to null for the API.
                var accountId = dlg.SelectedAccountId == AssigneeDialog.UnassignSentinel
                    ? null
                    : dlg.SelectedAccountId;

                _jira.SetAssignee(issue.Key, accountId);
                RefreshIssueAfterChange(issue.Key);
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Assignee", ex.Message, "OK");
            }
        }

        private void EditDescription()
        {
            var issue = RequireSelectedIssue();
            if (issue == null) return;

            try
            {
                var dlg = new TextEditorDialog(
                    "Edit description — " + issue.Key,
                    issue.Description ?? "",
                    "Save");
                Application.Run(dlg);
                if (!dlg.Saved) return;

                _jira.UpdateDescription(issue.Key, dlg.Result ?? "");
                RefreshIssueAfterChange(issue.Key);
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Description", ex.Message, "OK");
            }
        }

        private void AddComment()
        {
            var issue = RequireSelectedIssue();
            if (issue == null) return;

            try
            {
                var dlg = new TextEditorDialog(
                    "Add comment — " + issue.Key,
                    "",
                    "Send");
                Application.Run(dlg);
                if (!dlg.Saved) return;
                var body = (dlg.Result ?? "").Trim();
                if (body.Length == 0) return;

                _jira.AddComment(issue.Key, body);
                RefreshIssueAfterChange(issue.Key);
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Comment", ex.Message, "OK");
            }
        }

        private void AskAiForJql()
        {
            if (string.IsNullOrWhiteSpace(_config.Ai?.ApiKeyProtected))
            {
                MessageBox.Query("AI not configured",
                    "Skonfiguruj klucz Anthropic w F2 → tab AI.",
                    "OK");
                return;
            }

            var apiKey = Config.SecretProtector.Unprotect(_config.Ai.ApiKeyProtected);
            if (string.IsNullOrWhiteSpace(apiKey))
            {
                MessageBox.ErrorQuery("AI", "Nie udało się odszyfrować klucza AI.", "OK");
                return;
            }

            // Make the JQL bar visible so the user sees the result land where they
            // expect, even if they triggered Ctrl-G from somewhere else.
            if (!_jqlVisible)
            {
                _jqlVisible = true;
                ApplyJqlLayout();
            }

            using (var ai = new Ai.AiClient(apiKey, _config.Ai.Model))
            {
                var dlg = new AiJqlDialog(ai, _jira, _jqlBar.Jql);
                Application.Run(dlg);
                if (dlg.Accepted && !string.IsNullOrEmpty(dlg.GeneratedJql))
                {
                    _jqlBar.Jql = dlg.GeneratedJql;
                    SetPath("AI-generated JQL");
                    RunSearch(dlg.GeneratedJql);
                    _jqlBar.FocusInput();
                }
            }
        }

        private void SaveCurrentAsFilter()
        {
            var jql = (_jqlBar.Jql ?? "").Trim();
            if (string.IsNullOrEmpty(jql))
            {
                MessageBox.Query("Save filter", "JQL bar jest pusty — nie ma czego zapisać.", "OK");
                return;
            }

            var dlg = new SaveFilterDialog(jql);
            Application.Run(dlg);
            if (!dlg.Saved) return;

            try
            {
                var saved = _jira.SaveFilter(dlg.FilterName, dlg.FilterDescription, jql);
                // Repopulate nav so the new filter shows up under "Saved filters".
                try
                {
                    _nav.Populate(_jira.GetProjects(), _jira.GetSavedFilters());
                }
                catch { /* nav refresh is best-effort */ }

                MessageBox.Query("Save filter",
                    "Zapisano w Jirze: " + (saved?.Name ?? dlg.FilterName), "OK");
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Save filter", ex.Message, "OK");
            }
        }

        private Jira.Models.Issue RequireSelectedIssue()
        {
            var issue = _issueList.Selected;
            if (issue == null)
            {
                MessageBox.Query("No selection", "Najpierw zaznacz issue na liście.", "OK");
                return null;
            }
            return issue;
        }

        /// <summary>
        /// After a successful mutation, refetch the issue from Jira and update both
        /// the list row and the detail panel so the UI reflects authoritative state.
        /// </summary>
        private void RefreshIssueAfterChange(string key)
        {
            try
            {
                var fresh = _jira.GetIssue(key);
                if (fresh == null) return;
                _issueList.UpdateIssue(fresh);
                _detail.ShowIssue(fresh);
            }
            catch (Exception ex)
            {
                MessageBox.ErrorQuery("Refresh failed",
                    "Zmiana zapisana, ale nie udało się odświeżyć:\r\n" + ex.Message, "OK");
            }
        }
    }
}
