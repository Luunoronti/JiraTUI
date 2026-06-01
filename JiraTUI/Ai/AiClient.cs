using System;
using System.IO;
using System.Net;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Text;
using System.Threading.Tasks;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace JiraTUI.Ai
{
    /// <summary>
    /// Anthropic Claude REST API client. POSTs to /v1/messages.
    /// Synchronous — UI thread waits for the response. Acceptable for a one-off
    /// JQL generation that takes a few seconds; can be made async later.
    /// </summary>
    public class AiClient : IDisposable
    {
        private const string Endpoint = "https://api.anthropic.com/v1/messages";
        private const string AnthropicVersion = "2023-06-01";

        private readonly HttpClient _http;
        private readonly string _apiKey;
        private readonly string _model;

        static AiClient()
        {
            try { ServicePointManager.SecurityProtocol |= SecurityProtocolType.Tls12; }
            catch { /* older runtime — best effort */ }
        }

        public AiClient(string apiKey, string model)
        {
            _apiKey = apiKey ?? "";
            _model = string.IsNullOrWhiteSpace(model) ? "claude-sonnet-4-5" : model.Trim();

            _http = new HttpClient { Timeout = TimeSpan.FromSeconds(60) };
            _http.DefaultRequestHeaders.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));
            _http.DefaultRequestHeaders.UserAgent.ParseAdd("JiraTUI/0.1");
        }

        public void Dispose() => _http.Dispose();

        /// <summary>
        /// Defensive cleanup of model output. Anthropic usually obeys "no markdown"
        /// but occasionally wraps in ```jql … ``` fences anyway.
        /// </summary>
        public static string StripMarkdownFences(string s)
        {
            if (string.IsNullOrEmpty(s)) return "";
            s = s.Trim();
            if (s.StartsWith("```"))
            {
                int firstNl = s.IndexOf('\n');
                if (firstNl > 0) s = s.Substring(firstNl + 1);
                if (s.EndsWith("```")) s = s.Substring(0, s.Length - 3);
                s = s.Trim();
            }
            return s;
        }

        public string Generate(string systemPrompt, string userPrompt)
        {
            var body = new JObject
            {
                ["model"] = _model,
                ["max_tokens"] = 1024,
                ["system"] = systemPrompt ?? "",
                ["messages"] = new JArray(new JObject
                {
                    ["role"] = "user",
                    ["content"] = userPrompt ?? "",
                }),
            };

            using (var req = new HttpRequestMessage(HttpMethod.Post, Endpoint))
            {
                req.Headers.TryAddWithoutValidation("x-api-key", _apiKey);
                req.Headers.TryAddWithoutValidation("anthropic-version", AnthropicVersion);
                req.Content = new StringContent(body.ToString(Formatting.None), Encoding.UTF8, "application/json");

                HttpResponseMessage res;
                try
                {
                    res = _http.SendAsync(req, HttpCompletionOption.ResponseContentRead)
                        .GetAwaiter().GetResult();
                }
                catch (TaskCanceledException) { throw new Exception("Request timed out."); }
                catch (HttpRequestException ex) { throw new Exception("Network error: " + ex.Message); }

                using (res)
                {
                    var responseBody = ReadBody(res);
                    if (!res.IsSuccessStatusCode)
                    {
                        throw new Exception("HTTP " + (int)res.StatusCode + " " + res.StatusCode + "\r\n" +
                            ExtractError(responseBody));
                    }
                    return ExtractText(responseBody);
                }
            }
        }

        private static string ReadBody(HttpResponseMessage res)
        {
            using (var s = res.Content.ReadAsStreamAsync().GetAwaiter().GetResult())
            using (var r = new StreamReader(s, Encoding.UTF8))
                return r.ReadToEnd();
        }

        private static string ExtractText(string body)
        {
            try
            {
                var json = JObject.Parse(body);
                var content = json["content"] as JArray;
                if (content == null || content.Count == 0)
                    throw new Exception("Empty response from Anthropic.");
                var first = content[0];
                var text = (string)first["text"];
                if (string.IsNullOrEmpty(text))
                    throw new Exception("Response has no text content.");
                return text.Trim();
            }
            catch (Exception ex)
            {
                throw new Exception("Unparseable response: " + ex.Message + "\r\n" + Truncate(body, 400));
            }
        }

        private static string ExtractError(string body)
        {
            if (string.IsNullOrWhiteSpace(body)) return "";
            try
            {
                var j = JObject.Parse(body);
                var msg = (string)j["error"]?["message"];
                if (!string.IsNullOrEmpty(msg)) return msg;
            }
            catch { /* fall through */ }
            return Truncate(body, 400);
        }

        private static string Truncate(string s, int max)
        {
            if (string.IsNullOrEmpty(s)) return "";
            return s.Length <= max ? s : s.Substring(0, max - 1) + "…";
        }
    }
}
