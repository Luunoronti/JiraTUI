using System;
using JiraTUI.Config;
using JiraTUI.Jira;
using JiraTUI.Themes;
using JiraTUI.UI;
using Terminal.Gui;

namespace JiraTUI
{
    public static class Program
    {
        [STAThread]
        public static int Main(string[] args)
        {
            try
            {
                Application.Init();

                ThemeManager.BuildAll();

                var cfg = ConfigStore.Load();
                var themeName = ThemeManager.ResolveName(cfg.Appearance.ThemeName);
                ThemeManager.Apply(themeName);
                cfg.Appearance.ThemeName = themeName;
                ConfigStore.Save(cfg);

                IJiraClient jira = JiraClientFactory.Create(cfg);

                var main = new MainWindow(jira, cfg);
                Application.Run(main);
                Application.Shutdown();
                jira.Dispose();
                return 0;
            }
            catch (Exception ex)
            {
                try { Application.Shutdown(); } catch { /* ignore */ }
                Console.Error.WriteLine("Fatal: " + ex);
                return 1;
            }
        }
    }
}
