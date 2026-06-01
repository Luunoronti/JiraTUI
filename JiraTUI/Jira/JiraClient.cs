using System;
using System.Collections.Generic;
using System.IO;
using System.Net;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Threading.Tasks;
using JiraTUI.Config;
using JiraTUI.Jira.Models;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace JiraTUI.Jira
{
    /// <summary>
    /// Jira Cloud REST API v3 client.
    /// Auth: HTTP Basic with email + API token (base64 encoded).
    /// All calls are synchronous — Terminal.Gui's main loop is single-threaded and
    /// the volume here is low; this avoids async/sync deadlock pitfalls on .NET 4.8.
    /// </summary>
    public class JiraClient : IJiraClient
    {
        private readonly HttpClient _http;
        private readonly string _baseUrl;
        private readonly string _authHeader;

        private string _userDisplay = "(not connected)";

        public string CurrentUserDisplay => _userDisplay;
        public string ServerLabel => _baseUrl;

        static JiraClient()
        {
            // Jira Cloud requires TLS 1.2+. Don't downgrade existing flags.
            try { ServicePointManager.SecurityProtocol |= SecurityProtocolType.Tls12; }
            catch { /* older runtime — best effort */ }
        }

        // Diagnostic info exposed to error messages (no secrets).
        private readonly string _diagEmail;
        private readonly int _diagTokenLen;

        public JiraClient(AppConfig cfg)
        {
            _baseUrl = NormalizeBaseUrl(cfg.Connection.BaseUrl);

            var email = TrimAll(cfg.Connection.Email);
            var token = TrimAll(SecretProtector.Unprotect(cfg.Connection.TokenProtected));

            _diagEmail = email;
            _diagTokenLen = token.Length;

            var combined = email + ":" + token;
            var b64 = Convert.ToBase64String(Encoding.UTF8.GetBytes(combined));
            _authHeader = b64; // raw base64; AuthenticationHeaderValue prepends "Basic "

            _http = new HttpClient { Timeout = TimeSpan.FromSeconds(30) };
            _http.DefaultRequestHeaders.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));
            _http.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Basic", b64);
            _http.DefaultRequestHeaders.UserAgent.ParseAdd("JiraTUI/0.1");
        }

        private static string NormalizeBaseUrl(string raw)
        {
            var s = TrimAll(raw).TrimEnd('/');
            if (string.IsNullOrEmpty(s)) return "";
            // For Jira Cloud (*.atlassian.net) the REST root is scheme://host.
            // If the user pasted a deep URL (e.g. /jira/your-work), strip the path.
            try
            {
                var uri = new Uri(s);
                if (uri.Host.EndsWith(".atlassian.net", StringComparison.OrdinalIgnoreCase))
                    return uri.Scheme + "://" + uri.Host;
            }
            catch { /* not a parseable URI — return as-is */ }
            return s;
        }

        private static string TrimAll(string s)
        {
            if (string.IsNullOrEmpty(s)) return "";
            // Trim ordinary whitespace + invisible characters that sneak in via clipboard
            // (BOM, NBSP, zero-width space).
            return s.Trim(' ', '\t', '\r', '\n', ' ', '​', '﻿');
        }

        public void Dispose() => _http.Dispose();

        // ============= IJiraClient =============

        public bool TestConnection(out string error)
        {
            try
            {
                var json = GetJson("/rest/api/3/myself");
                _userDisplay = (string)json["displayName"]
                    ?? (string)json["emailAddress"]
                    ?? "(unknown user)";
                error = null;
                return true;
            }
            catch (Exception ex)
            {
                error = Flatten(ex);
                return false;
            }
        }

        public IList<Project> GetProjects()
        {
            var result = new List<Project>();
            string url = "/rest/api/3/project/search?maxResults=100";

            while (!string.IsNullOrEmpty(url))
            {
                var json = GetJson(url);
                var values = json["values"] as JArray;
                if (values != null)
                {
                    foreach (var v in values)
                    {
                        result.Add(new Project
                        {
                            Key = (string)v["key"],
                            Name = (string)v["name"],
                        });
                    }
                }

                // Project search uses startAt/total/isLast — follow nextPage if present.
                bool isLast = json["isLast"]?.Value<bool>() ?? true;
                var nextPage = (string)json["nextPage"];
                if (!isLast && !string.IsNullOrEmpty(nextPage))
                {
                    // nextPage is absolute; convert to relative.
                    int idx = nextPage.IndexOf("/rest/", StringComparison.OrdinalIgnoreCase);
                    url = idx >= 0 ? nextPage.Substring(idx) : null;
                }
                else
                {
                    url = null;
                }
            }

            return result;
        }

        public IList<SavedFilter> GetSavedFilters()
        {
            var result = new List<SavedFilter>();
            try
            {
                var arr = GetJsonArray("/rest/api/3/filter/my?includeFavourites=true");
                foreach (var v in arr)
                {
                    result.Add(new SavedFilter
                    {
                        Name = (string)v["name"],
                        Jql = (string)v["jql"],
                    });
                }
            }
            catch
            {
                // Saved filters are non-critical — fail open, just return empty list.
            }
            return result;
        }

        public IList<Issue> SearchIssues(string jql, int maxResults)
        {
            // Pull description and comments in the same request so the detail panel
            // can render instantly when the user moves through the list. Refresh (F5)
            // becomes the meaningful way to re-sync state.
            //
            // "*navigable" returns every navigable field, including custom ones such
            // as Sprint (whose id differs between Jira sites). description + comment
            // aren't navigable so we add them explicitly.
            var body = new JObject
            {
                ["jql"] = jql ?? "",
                ["maxResults"] = maxResults,
                ["fields"] = new JArray("*navigable", "description", "comment"),
            };

            var json = PostJson("/rest/api/3/search/jql", body);
            var issues = json["issues"] as JArray;
            var list = new List<Issue>();
            if (issues != null)
            {
                foreach (var i in issues)
                    list.Add(MapIssueFull(i));
            }
            return list;
        }

        public Issue GetIssue(string key)
        {
            if (string.IsNullOrEmpty(key)) return null;
            var json = GetJson("/rest/api/3/issue/" + Uri.EscapeDataString(key));
            return MapIssueFull(json);
        }

        // ============= metadata =============

        // Cached after first successful detection on a real issue payload, so we don't
        // re-scan every custom field on every issue.
        private string _sprintFieldId;

        private IList<string> _cachedPriorities;

        public IList<string> GetPriorityNames()
        {
            if (_cachedPriorities != null) return _cachedPriorities;
            var list = new List<string>();
            try
            {
                var arr = GetJsonArray("/rest/api/3/priority");
                foreach (var p in arr)
                {
                    var name = (string)p["name"];
                    if (!string.IsNullOrEmpty(name)) list.Add(name);
                }
            }
            catch { /* fall back to defaults */ }

            if (list.Count == 0)
                list.AddRange(new[] { "Highest", "High", "Medium", "Low", "Lowest" });

            _cachedPriorities = list;
            return list;
        }

        private IList<string> _cachedStatuses;
        private IList<string> _cachedIssueTypes;

        public IList<string> GetStatusNames()
        {
            if (_cachedStatuses != null) return _cachedStatuses;
            var list = new List<string>();
            try
            {
                var arr = GetJsonArray("/rest/api/3/status");
                var seen = new HashSet<string>(StringComparer.OrdinalIgnoreCase);
                foreach (var s in arr)
                {
                    var name = (string)s["name"];
                    if (!string.IsNullOrEmpty(name) && seen.Add(name)) list.Add(name);
                }
            }
            catch { /* fall back to defaults */ }

            if (list.Count == 0)
                list.AddRange(new[] { "Open", "In Progress", "In Review", "Done", "Blocked" });

            _cachedStatuses = list;
            return list;
        }

        public IList<string> GetIssueTypeNames()
        {
            if (_cachedIssueTypes != null) return _cachedIssueTypes;
            var list = new List<string>();
            try
            {
                var arr = GetJsonArray("/rest/api/3/issuetype");
                var seen = new HashSet<string>(StringComparer.OrdinalIgnoreCase);
                foreach (var t in arr)
                {
                    var name = (string)t["name"];
                    if (!string.IsNullOrEmpty(name) && seen.Add(name)) list.Add(name);
                }
            }
            catch { /* fall back to defaults */ }

            if (list.Count == 0)
                list.AddRange(new[] { "Bug", "Task", "Story", "Epic", "Sub-task" });

            _cachedIssueTypes = list;
            return list;
        }

        public IList<Transition> GetTransitions(string issueKey)
        {
            var list = new List<Transition>();
            if (string.IsNullOrEmpty(issueKey)) return list;
            var json = GetJson("/rest/api/3/issue/" + Uri.EscapeDataString(issueKey) + "/transitions");
            var arr = json["transitions"] as JArray;
            if (arr == null) return list;
            foreach (var t in arr)
            {
                list.Add(new Transition
                {
                    Id = (string)t["id"],
                    Name = (string)t["name"],
                    ToStatus = (string)t["to"]?["name"],
                });
            }
            return list;
        }

        public IList<JiraUser> SearchAssignableUsers(string issueKey, string query)
        {
            var list = new List<JiraUser>();
            if (string.IsNullOrEmpty(issueKey)) return list;

            // Empty query → Jira returns the default candidate set for this issue.
            var url = "/rest/api/3/user/assignable/search?issueKey=" + Uri.EscapeDataString(issueKey)
                    + "&maxResults=30";
            if (!string.IsNullOrWhiteSpace(query))
                url += "&query=" + Uri.EscapeDataString(query);

            var arr = GetJsonArray(url);
            foreach (var u in arr)
            {
                list.Add(new JiraUser
                {
                    AccountId = (string)u["accountId"],
                    DisplayName = (string)u["displayName"],
                    EmailAddress = (string)u["emailAddress"],
                });
            }
            return list;
        }

        // ============= mutations =============

        public void SetPriority(string issueKey, string priorityName)
        {
            var body = new JObject
            {
                ["fields"] = new JObject
                {
                    ["priority"] = new JObject { ["name"] = priorityName ?? "" }
                }
            };
            PutJsonNoResponse("/rest/api/3/issue/" + Uri.EscapeDataString(issueKey), body);
        }

        public void SetAssignee(string issueKey, string accountIdOrNull)
        {
            var body = new JObject
            {
                // null accountId → unassign (Jira expects literal JSON null)
                ["accountId"] = accountIdOrNull == null ? JValue.CreateNull() : new JValue(accountIdOrNull)
            };
            PutJsonNoResponse("/rest/api/3/issue/" + Uri.EscapeDataString(issueKey) + "/assignee", body);
        }

        public void TransitionIssue(string issueKey, string transitionId)
        {
            var body = new JObject
            {
                ["transition"] = new JObject { ["id"] = transitionId }
            };
            PostJsonNoResponse("/rest/api/3/issue/" + Uri.EscapeDataString(issueKey) + "/transitions", body);
        }

        public void UpdateDescription(string issueKey, string plainTextDescription)
        {
            var body = new JObject
            {
                ["fields"] = new JObject
                {
                    ["description"] = PlainTextToAdf(plainTextDescription ?? "")
                }
            };
            PutJsonNoResponse("/rest/api/3/issue/" + Uri.EscapeDataString(issueKey), body);
        }

        public void AddComment(string issueKey, string plainTextComment)
        {
            var body = new JObject
            {
                ["body"] = PlainTextToAdf(plainTextComment ?? "")
            };
            PostJsonNoResponse("/rest/api/3/issue/" + Uri.EscapeDataString(issueKey) + "/comment", body);
        }

        public SavedFilter SaveFilter(string name, string description, string jql)
        {
            var body = new JObject
            {
                ["name"] = name ?? "",
                ["description"] = description ?? "",
                ["jql"] = jql ?? "",
                ["favourite"] = true,
            };
            var json = PostJson("/rest/api/3/filter", body);
            return new SavedFilter
            {
                Name = (string)json["name"] ?? name,
                Jql = (string)json["jql"] ?? jql,
            };
        }

        /// <summary>
        /// Wrap plain multiline text as a minimal ADF document — one paragraph per
        /// non-empty line, empty paragraphs for blank lines so spacing survives the
        /// round-trip when re-rendered. Loses any rich formatting from the original,
        /// which is the accepted trade-off for plain-text TUI editing.
        /// </summary>
        private static JObject PlainTextToAdf(string text)
        {
            var paragraphs = new JArray();
            var normalized = (text ?? "").Replace("\r\n", "\n").Replace("\r", "\n");
            var lines = normalized.Split('\n');
            foreach (var line in lines)
            {
                var content = new JArray();
                if (!string.IsNullOrEmpty(line))
                {
                    content.Add(new JObject
                    {
                        ["type"] = "text",
                        ["text"] = line,
                    });
                }
                paragraphs.Add(new JObject
                {
                    ["type"] = "paragraph",
                    ["content"] = content,
                });
            }
            return new JObject
            {
                ["version"] = 1,
                ["type"] = "doc",
                ["content"] = paragraphs,
            };
        }

        // ============= mapping =============

        private Issue MapIssueRow(JToken jt)
        {
            var fields = jt["fields"];
            var key = (string)jt["key"] ?? "";
            return new Issue
            {
                Key = key,
                ProjectKey = key.IndexOf('-') > 0 ? key.Substring(0, key.IndexOf('-')) : key,
                Summary = (string)fields?["summary"],
                Status = (string)fields?["status"]?["name"],
                Priority = (string)fields?["priority"]?["name"],
                IssueType = (string)fields?["issuetype"]?["name"],
                Assignee = (string)fields?["assignee"]?["displayName"],
                Reporter = (string)fields?["reporter"]?["displayName"],
                Labels = ParseLabels(fields?["labels"]),
                Updated = ParseDate((string)fields?["updated"]),
                Sprint = ExtractSprintName(fields),
            };
        }

        /// <summary>
        /// Sprint is a Jira Software custom field (id varies between sites). Its value
        /// is an array of sprint objects with at least name+state, usually also id and
        /// boardId. We scan all customfield_* properties for an array whose first
        /// element looks like a sprint, then prefer active → future → last entry.
        /// </summary>
        private string ExtractSprintName(JToken fields)
        {
            if (fields == null) return null;

            JArray sprints = null;

            // Fast path: re-use the id we cached from a previous mapping.
            if (!string.IsNullOrEmpty(_sprintFieldId))
                sprints = fields[_sprintFieldId] as JArray;

            if (sprints == null || sprints.Count == 0)
            {
                if (fields is JObject obj)
                {
                    foreach (var kv in obj)
                    {
                        if (!kv.Key.StartsWith("customfield_", StringComparison.OrdinalIgnoreCase)) continue;
                        var arr = kv.Value as JArray;
                        if (arr == null || arr.Count == 0) continue;
                        var first = arr[0] as JObject;
                        if (first == null) continue;
                        // A sprint object has a name plus at least one of these
                        // sprint-specific properties — enough to disambiguate from
                        // other array-typed custom fields (versions, components, …).
                        if (first["name"] == null) continue;
                        if (first["state"] == null && first["boardId"] == null
                            && first["startDate"] == null && first["completeDate"] == null)
                            continue;

                        sprints = arr;
                        _sprintFieldId = kv.Key;
                        break;
                    }
                }
            }

            if (sprints == null || sprints.Count == 0) return null;

            JToken active = null, future = null;
            foreach (var s in sprints)
            {
                var state = ((string)s["state"] ?? "").ToLowerInvariant();
                if (state == "active" && active == null) active = s;
                else if (state == "future" && future == null) future = s;
            }
            var pick = active ?? future ?? sprints[sprints.Count - 1];
            return (string)pick?["name"];
        }

        private Issue MapIssueFull(JToken jt)
        {
            var i = MapIssueRow(jt);
            var fields = jt["fields"];

            i.Description = AdfTextRenderer.Render(fields?["description"]);

            var comments = fields?["comment"]?["comments"] as JArray;
            if (comments != null)
            {
                foreach (var c in comments)
                {
                    i.Comments.Add(new Comment
                    {
                        Author = (string)c["author"]?["displayName"]
                                ?? (string)c["author"]?["emailAddress"]
                                ?? "(unknown)",
                        Created = ParseDate((string)c["created"]),
                        Body = AdfTextRenderer.Render(c["body"]),
                    });
                }
            }

            return i;
        }

        private static List<string> ParseLabels(JToken jt)
        {
            var list = new List<string>();
            var arr = jt as JArray;
            if (arr == null) return list;
            foreach (var l in arr)
                if (l.Type == JTokenType.String) list.Add((string)l);
            return list;
        }

        private static DateTime ParseDate(string s)
        {
            if (string.IsNullOrEmpty(s)) return DateTime.MinValue;
            if (DateTime.TryParse(s, System.Globalization.CultureInfo.InvariantCulture,
                System.Globalization.DateTimeStyles.AdjustToUniversal | System.Globalization.DateTimeStyles.AssumeUniversal,
                out var dt))
                return dt;
            return DateTime.MinValue;
        }

        // ============= HTTP =============

        private JObject GetJson(string relativeUrl)
        {
            using (var req = NewRequest(HttpMethod.Get, relativeUrl))
            using (var res = Send(req))
            {
                return JObject.Parse(ReadBody(res));
            }
        }

        private JArray GetJsonArray(string relativeUrl)
        {
            using (var req = NewRequest(HttpMethod.Get, relativeUrl))
            using (var res = Send(req))
            {
                return JArray.Parse(ReadBody(res));
            }
        }

        private JObject PostJson(string relativeUrl, JObject body)
        {
            using (var req = NewRequest(HttpMethod.Post, relativeUrl))
            {
                req.Content = new StringContent(body.ToString(Formatting.None), Encoding.UTF8, "application/json");
                using (var res = Send(req))
                {
                    return JObject.Parse(ReadBody(res));
                }
            }
        }

        private void PostJsonNoResponse(string relativeUrl, JObject body)
        {
            using (var req = NewRequest(HttpMethod.Post, relativeUrl))
            {
                req.Content = new StringContent(body.ToString(Formatting.None), Encoding.UTF8, "application/json");
                using (var res = Send(req)) { /* discard body — 204 typical */ }
            }
        }

        private void PutJsonNoResponse(string relativeUrl, JObject body)
        {
            using (var req = NewRequest(HttpMethod.Put, relativeUrl))
            {
                req.Content = new StringContent(body.ToString(Formatting.None), Encoding.UTF8, "application/json");
                using (var res = Send(req)) { /* discard body — 204 typical */ }
            }
        }

        private HttpRequestMessage NewRequest(HttpMethod method, string relativeUrl)
        {
            if (string.IsNullOrEmpty(_baseUrl))
                throw new InvalidOperationException("Jira Base URL not configured.");

            var fullUrl = _baseUrl + (relativeUrl.StartsWith("/") ? relativeUrl : "/" + relativeUrl);
            // Authorization comes from HttpClient.DefaultRequestHeaders so it survives
            // redirects and works reliably with TryAddWithoutValidation edge cases.
            return new HttpRequestMessage(method, fullUrl);
        }

        private HttpResponseMessage Send(HttpRequestMessage req)
        {
            HttpResponseMessage res;
            try
            {
                res = _http.SendAsync(req, HttpCompletionOption.ResponseContentRead).GetAwaiter().GetResult();
            }
            catch (HttpRequestException ex)
            {
                throw new JiraException("Network error: " + Flatten(ex), ex);
            }
            catch (TaskCanceledException ex)
            {
                throw new JiraException("Request timed out.", ex);
            }

            if (!res.IsSuccessStatusCode)
            {
                string body = SafeRead(res);
                string snippet = TruncateForError(body);
                var code = (int)res.StatusCode;
                var status = res.StatusCode.ToString();
                res.Dispose();

                var msg = new StringBuilder();
                msg.Append("HTTP ").Append(code).Append(' ').Append(status);
                if (!string.IsNullOrEmpty(snippet))
                    msg.Append("\r\n").Append(snippet);

                // Extra hints for the most common Basic-auth screw-ups.
                if (code == 401)
                {
                    msg.AppendLine().AppendLine();
                    msg.AppendLine("Diagnostyka:");
                    msg.AppendLine("  URL    : " + _baseUrl);
                    msg.AppendLine("  Email  : " + (string.IsNullOrEmpty(_diagEmail) ? "(empty)" : _diagEmail));
                    msg.AppendLine("  Token  : długość " + _diagTokenLen + " znaków");
                    msg.AppendLine();
                    msg.AppendLine("Sprawdź:");
                    msg.AppendLine("  • Email to adres logowania Atlassian (nie display name).");
                    msg.AppendLine("  • Token z id.atlassian.com/manage-profile/security/api-tokens,");
                    msg.AppendLine("    skopiowany bez spacji / nowych linii.");
                    msg.AppendLine("  • Base URL = https://TWOJA.atlassian.net (bez ścieżki).");
                }

                throw new JiraException(msg.ToString());
            }

            return res;
        }

        private static string ReadBody(HttpResponseMessage res)
        {
            using (var s = res.Content.ReadAsStreamAsync().GetAwaiter().GetResult())
            using (var r = new StreamReader(s, Encoding.UTF8))
            {
                return r.ReadToEnd();
            }
        }

        private static string SafeRead(HttpResponseMessage res)
        {
            try { return ReadBody(res); } catch { return ""; }
        }

        private static string TruncateForError(string body)
        {
            if (string.IsNullOrWhiteSpace(body)) return "";
            // Try to surface Jira's error message instead of a wall of JSON.
            try
            {
                var jt = JToken.Parse(body);
                var messages = jt["errorMessages"] as JArray;
                if (messages != null && messages.Count > 0)
                    return string.Join("\r\n", messages);
                var errs = jt["errors"] as JObject;
                if (errs != null && errs.Count > 0)
                {
                    var sb = new StringBuilder();
                    foreach (var kv in errs)
                        sb.AppendLine(kv.Key + ": " + kv.Value);
                    return sb.ToString().TrimEnd();
                }
            }
            catch { /* fall through */ }

            return body.Length > 500 ? body.Substring(0, 500) + "…" : body;
        }

        private static string Flatten(Exception ex)
        {
            var sb = new StringBuilder();
            for (var e = ex; e != null; e = e.InnerException)
            {
                if (sb.Length > 0) sb.Append(" → ");
                sb.Append(e.Message);
            }
            return sb.ToString();
        }
    }

    public class JiraException : Exception
    {
        public JiraException(string msg) : base(msg) { }
        public JiraException(string msg, Exception inner) : base(msg, inner) { }
    }
}
