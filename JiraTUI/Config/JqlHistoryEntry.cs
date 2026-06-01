using System;
using Newtonsoft.Json;

namespace JiraTUI.Config
{
    public class JqlHistoryEntry
    {
        [JsonProperty("timestamp")]
        public DateTime Timestamp { get; set; }

        /// <summary>What the user typed in the JQL bar.</summary>
        [JsonProperty("originalText")]
        public string OriginalText { get; set; }

        /// <summary>The JQL that was actually sent to Jira. Same as OriginalText
        /// for normal queries; differs for AI-translated ones.</summary>
        [JsonProperty("effectiveJql")]
        public string EffectiveJql { get; set; }

        [JsonProperty("wasAiTranslated")]
        public bool WasAiTranslated { get; set; }
    }
}
