using Newtonsoft.Json;

namespace JiraTUI.Config
{
    public class AppConfig
    {
        [JsonProperty("connection")]
        public ConnectionConfig Connection { get; set; } = new ConnectionConfig();

        [JsonProperty("appearance")]
        public AppearanceConfig Appearance { get; set; } = new AppearanceConfig();

        [JsonProperty("behavior")]
        public BehaviorConfig Behavior { get; set; } = new BehaviorConfig();

        [JsonProperty("ai")]
        public AiConfig Ai { get; set; } = new AiConfig();

        [JsonProperty("columns")]
        public ColumnVisibilityConfig Columns { get; set; } = new ColumnVisibilityConfig();
    }

    public class ColumnVisibilityConfig
    {
        [JsonProperty("key")]      public bool Key      { get; set; } = true;
        [JsonProperty("type")]     public bool Type     { get; set; } = true;
        [JsonProperty("priority")] public bool Priority { get; set; } = true;
        [JsonProperty("status")]   public bool Status   { get; set; } = true;
        [JsonProperty("assignee")] public bool Assignee { get; set; } = true;
        [JsonProperty("summary")]  public bool Summary  { get; set; } = true;
    }

    public class AiConfig
    {
        [JsonProperty("provider")]
        public string Provider { get; set; } = "Anthropic";

        [JsonProperty("model")]
        public string Model { get; set; } = "claude-sonnet-4-5";

        [JsonProperty("apiKeyProtected")]
        public string ApiKeyProtected { get; set; } = "";
    }

    public class ConnectionConfig
    {
        [JsonProperty("baseUrl")]
        public string BaseUrl { get; set; } = "https://example.atlassian.net";

        [JsonProperty("email")]
        public string Email { get; set; } = "";

        // DPAPI-protected, base64-encoded. Never plaintext on disk.
        [JsonProperty("tokenProtected")]
        public string TokenProtected { get; set; } = "";

        [JsonProperty("authType")]
        public AuthType AuthType { get; set; } = AuthType.ApiToken;
    }

    public enum AuthType
    {
        Basic = 0,
        ApiToken = 1,
        PersonalAccessToken = 2
    }

    public class AppearanceConfig
    {
        [JsonProperty("themeName")]
        public string ThemeName { get; set; } = "Default";
    }

    public class BehaviorConfig
    {
        [JsonProperty("defaultJql")]
        public string DefaultJql { get; set; } = "assignee = currentUser() ORDER BY updated DESC";

        [JsonProperty("pageSize")]
        public int PageSize { get; set; } = 50;

        [JsonProperty("autoRefreshSeconds")]
        public int AutoRefreshSeconds { get; set; } = 0;
    }
}
