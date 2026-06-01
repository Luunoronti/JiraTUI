using JiraTUI.Config;

namespace JiraTUI.Jira
{
    public static class JiraClientFactory
    {
        public static IJiraClient Create(AppConfig cfg)
        {
            return HasCreds(cfg) ? (IJiraClient)new JiraClient(cfg) : new MockJiraClient();
        }

        public static bool HasCreds(AppConfig cfg)
        {
            if (cfg?.Connection == null) return false;
            return !string.IsNullOrWhiteSpace(cfg.Connection.BaseUrl)
                && !string.IsNullOrWhiteSpace(cfg.Connection.Email)
                && !string.IsNullOrWhiteSpace(cfg.Connection.TokenProtected);
        }
    }
}
