using System;
using System.Collections.Generic;
using System.IO;
using Newtonsoft.Json;

namespace JiraTUI.Config
{
    /// <summary>
    /// Persistent JQL history. Stored in the user's Documents folder so it
    /// rides along with OneDrive / similar sync without us having to do anything.
    /// </summary>
    public class JqlHistory
    {
        private const int MaxEntries = 200;
        private List<JqlHistoryEntry> _entries = new List<JqlHistoryEntry>();

        public static string FilePath
        {
            get
            {
                var docs = Environment.GetFolderPath(Environment.SpecialFolder.MyDocuments);
                return Path.Combine(docs, "JiraTUI", "jql-history.json");
            }
        }

        public int Count => _entries.Count;

        public static JqlHistory Load()
        {
            var h = new JqlHistory();
            try
            {
                if (File.Exists(FilePath))
                {
                    var json = File.ReadAllText(FilePath);
                    var loaded = JsonConvert.DeserializeObject<List<JqlHistoryEntry>>(json);
                    if (loaded != null) h._entries = loaded;
                }
            }
            catch { /* corrupt file → start with empty history */ }
            return h;
        }

        /// <summary>
        /// Add or refresh an entry. Entries are deduplicated by OriginalText —
        /// re-submitting an existing query moves it to the most-recent slot
        /// without leaving a stale copy behind.
        /// </summary>
        public void Add(string originalText, string effectiveJql, bool wasAiTranslated)
        {
            if (string.IsNullOrWhiteSpace(originalText)) return;

            _entries.RemoveAll(e => string.Equals(e.OriginalText, originalText, StringComparison.Ordinal));
            _entries.Add(new JqlHistoryEntry
            {
                Timestamp = DateTime.UtcNow,
                OriginalText = originalText,
                EffectiveJql = string.IsNullOrEmpty(effectiveJql) ? originalText : effectiveJql,
                WasAiTranslated = wasAiTranslated,
            });
            while (_entries.Count > MaxEntries) _entries.RemoveAt(0);
            Save();
        }

        /// <summary>
        /// Get entry by recency index. 0 = most recent, 1 = previous, etc.
        /// Returns null when the index is past the oldest entry.
        /// </summary>
        public JqlHistoryEntry GetByRecentIndex(int recentIndex)
        {
            if (recentIndex < 0 || recentIndex >= _entries.Count) return null;
            return _entries[_entries.Count - 1 - recentIndex];
        }

        private void Save()
        {
            try
            {
                var dir = Path.GetDirectoryName(FilePath);
                if (!string.IsNullOrEmpty(dir) && !Directory.Exists(dir))
                    Directory.CreateDirectory(dir);
                File.WriteAllText(FilePath,
                    JsonConvert.SerializeObject(_entries, Formatting.Indented));
            }
            catch { /* don't surface — history is best-effort persistence */ }
        }
    }
}
