using System;
using System.IO;
using Newtonsoft.Json;

namespace JiraTUI.Config
{
    public static class ConfigStore
    {
        private static readonly string Dir =
            Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData), "JiraTUI");

        private static readonly string FilePath = Path.Combine(Dir, "config.json");

        public static string ConfigFilePath => FilePath;

        public static AppConfig Load()
        {
            try
            {
                if (!File.Exists(FilePath))
                    return new AppConfig();

                var json = File.ReadAllText(FilePath);
                var cfg = JsonConvert.DeserializeObject<AppConfig>(json);
                return cfg ?? new AppConfig();
            }
            catch
            {
                return new AppConfig();
            }
        }

        public static void Save(AppConfig cfg)
        {
            if (!Directory.Exists(Dir))
                Directory.CreateDirectory(Dir);

            var json = JsonConvert.SerializeObject(cfg, Formatting.Indented);
            File.WriteAllText(FilePath, json);
        }
    }
}
