using System.Collections.Generic;
using System.Linq;
using Terminal.Gui;

namespace JiraTUI.Themes
{
    public static class ThemeManager
    {
        private static readonly Dictionary<string, Theme> _themes =
            new Dictionary<string, Theme>(System.StringComparer.OrdinalIgnoreCase);

        public static IEnumerable<string> AvailableThemes => _themes.Keys;

        public static string CurrentThemeName { get; private set; } = "Default";

        public static void BuildAll()
        {
            _themes.Clear();
            _themes["Default"]        = MakeDefault();
            _themes["TurboPascal 5"]  = MakeTurboPascal5();
            _themes["Anders"]         = MakeAnders();
            _themes["Dark"]           = MakeDark();
            _themes["Light"]          = MakeLight();
            _themes["Light+"]         = MakeLightPlus();
            _themes["Green Phosphor"] = MakeGreenPhosphor();
            _themes["Amber Phosphor"] = MakeAmberPhosphor();
            _themes["8-Bit"]          = Make8Bit();
            _themes["Solarized"]      = MakeSolarized();
            _themes["Solarized Dark"] = MakeSolarizedDark();
        }

        /// <summary>
        /// Mutate the live Colors.* ColorScheme instances rather than swap their
        /// references. Views in Terminal.Gui 1.x cache the reference they were given
        /// at construction time, so reassignment doesn't propagate — but mutating
        /// in place is picked up on the next redraw.
        /// </summary>
        public static void Apply(string name)
        {
            if (!_themes.ContainsKey(name))
                name = "Default";

            var t = _themes[name];
            Mutate(Colors.Base,     t.Base);
            Mutate(Colors.Menu,     t.Menu);
            Mutate(Colors.Dialog,   t.Dialog);
            Mutate(Colors.Error,    t.Error);
            Mutate(Colors.TopLevel, t.TopLevel);

            CurrentThemeName = t.Name;
        }

        public static string ResolveName(string requested)
        {
            if (!string.IsNullOrEmpty(requested) && _themes.ContainsKey(requested))
                return _themes.Keys.First(k => k.Equals(requested, System.StringComparison.OrdinalIgnoreCase));
            return "Default";
        }

        private static void Mutate(ColorScheme target, ColorScheme src)
        {
            if (target == null || src == null) return;
            target.Normal    = src.Normal;
            target.Focus     = src.Focus;
            target.HotNormal = src.HotNormal;
            target.HotFocus  = src.HotFocus;
            target.Disabled  = src.Disabled;
        }

        private static Attribute Attr(Color fg, Color bg)
        {
            var drv = Application.Driver;
            return drv != null ? drv.MakeAttribute(fg, bg) : Attribute.Make(fg, bg);
        }

        // =====================================================================
        // Default — matches Terminal.Gui 1.x's hardcoded Colors initialization
        // verbatim. Used as a "known good" reference that's been battle-tested.
        // =====================================================================
        private static Theme MakeDefault() => new Theme
        {
            Name = "Default",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.BrightGreen, Color.Black),
                Focus     = Attr(Color.White,       Color.Cyan),
                HotNormal = Attr(Color.Brown,       Color.Black),
                HotFocus  = Attr(Color.Blue,        Color.Cyan),
                Disabled  = Attr(Color.DarkGray,    Color.Black),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.White,       Color.Black),
                Focus     = Attr(Color.Black,       Color.Gray),
                HotNormal = Attr(Color.BrightCyan,  Color.Black),
                HotFocus  = Attr(Color.BrightBlue,  Color.Gray),
                Disabled  = Attr(Color.White,       Color.Black),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.White,        Color.DarkGray),
                Focus     = Attr(Color.White,        Color.Black),
                HotNormal = Attr(Color.BrightYellow, Color.DarkGray),
                HotFocus  = Attr(Color.BrightYellow, Color.Black),
                Disabled  = Attr(Color.Gray,         Color.DarkGray),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.Black,       Color.Gray),
                Focus     = Attr(Color.White,       Color.DarkGray),
                HotNormal = Attr(Color.Blue,        Color.Gray),
                HotFocus  = Attr(Color.BrightBlue,  Color.DarkGray),
                Disabled  = Attr(Color.DarkGray,    Color.Gray),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.Red,         Color.White),
                Focus     = Attr(Color.Black,       Color.BrightRed),
                HotNormal = Attr(Color.Black,       Color.White),
                HotFocus  = Attr(Color.White,       Color.BrightRed),
                Disabled  = Attr(Color.DarkGray,    Color.White),
            },
        };

        // =====================================================================
        // TurboPascal 5 — classic blue/cyan/yellow palette from the config.
        // Accent → TopLevel, color name mappings for TG v1:
        //   Yellow → BrightYellow, LightGray → Gray, BrightRed stays BrightRed.
        // Focus uses Black/BrightCyan — maximum contrast against the Blue base.
        // =====================================================================
        private static Theme MakeTurboPascal5() => new Theme
        {
            Name = "TurboPascal 5",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.White,      Color.Blue),
                Focus     = Attr(Color.Black,      Color.BrightCyan),
                HotNormal = Attr(Color.White,      Color.Blue),
                HotFocus  = Attr(Color.Black,      Color.BrightCyan),
                Disabled  = Attr(Color.White,      Color.Blue),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.White,        Color.Blue),
                Focus     = Attr(Color.Black,        Color.BrightCyan),
                HotNormal = Attr(Color.BrightCyan,   Color.Blue),
                HotFocus  = Attr(Color.Black,        Color.BrightCyan),
                Disabled  = Attr(Color.DarkGray,     Color.Blue),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.Black,     Color.Cyan),
                Focus     = Attr(Color.Black,     Color.Gray),
                HotNormal = Attr(Color.BrightRed, Color.Cyan),
                HotFocus  = Attr(Color.BrightRed, Color.Gray),
                Disabled  = Attr(Color.DarkGray,  Color.Cyan),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.Black,      Color.Gray),
                Focus     = Attr(Color.White,      Color.DarkGray),
                HotNormal = Attr(Color.Blue,       Color.Gray),
                HotFocus  = Attr(Color.BrightBlue, Color.DarkGray),
                Disabled  = Attr(Color.DarkGray,   Color.Gray),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.BrightRed, Color.Gray),
                Focus     = Attr(Color.White,     Color.BrightRed),
                HotNormal = Attr(Color.Black,     Color.Gray),
                HotFocus  = Attr(Color.White,     Color.BrightRed),
                Disabled  = Attr(Color.DarkGray,  Color.Gray),
            },
        };

        // =====================================================================
        // Anders — dark-blue background with white text, style from the config.
        // WhiteSmoke → White, DimGray → DarkGray, DarkBlue → Blue,
        // BrightBlue stays, LightGray → Gray, Red stays.
        // Focus uses Black/BrightCyan for clear list-selection contrast.
        // =====================================================================
        private static Theme MakeAnders() => new Theme
        {
            Name = "Anders",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.White,      Color.DarkGray),
                Focus     = Attr(Color.Black,      Color.BrightCyan),
                HotNormal = Attr(Color.White,      Color.DarkGray),
                HotFocus  = Attr(Color.Black,      Color.BrightCyan),
                Disabled  = Attr(Color.DarkGray,   Color.DarkGray),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.White,      Color.DarkGray),
                Focus     = Attr(Color.Black,      Color.BrightCyan),
                HotNormal = Attr(Color.BrightCyan, Color.Blue),
                HotFocus  = Attr(Color.Black,      Color.BrightCyan),
                Disabled  = Attr(Color.White,      Color.Blue),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.White,        Color.Blue),
                Focus     = Attr(Color.Black,        Color.Gray),
                HotNormal = Attr(Color.BrightYellow, Color.Blue),
                HotFocus  = Attr(Color.Black,        Color.BrightYellow),
                Disabled  = Attr(Color.DarkGray,     Color.Blue),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.BrightBlue, Color.Gray),
                Focus     = Attr(Color.White,      Color.DarkGray),
                HotNormal = Attr(Color.BrightCyan, Color.Gray),
                HotFocus  = Attr(Color.White,      Color.DarkGray),
                Disabled  = Attr(Color.DarkGray,   Color.Gray),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.Red,      Color.White),
                Focus     = Attr(Color.White,    Color.Red),
                HotNormal = Attr(Color.Black,    Color.White),
                HotFocus  = Attr(Color.White,    Color.BrightRed),
                Disabled  = Attr(Color.DarkGray, Color.White),
            },
        };

        // =====================================================================
        // Dark — from config. TG v2 color mappings for v1:
        //   LightGray/Silver → Gray, Onyx → Black, Charcoal → DarkGray,
        //   OuterSpace → DarkGray, SlateGray → Gray,
        //   IndianRed/LightCoral → BrightRed.
        //   "None" background → Black.
        // =====================================================================
        private static Theme MakeDark() => new Theme
        {
            Name = "Dark",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.Gray,     Color.Black),
                Focus     = Attr(Color.White,    Color.DarkGray),
                HotNormal = Attr(Color.Gray,     Color.Black),
                HotFocus  = Attr(Color.White,    Color.DarkGray),
                Disabled  = Attr(Color.DarkGray, Color.Black),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.Gray,     Color.Black),
                Focus     = Attr(Color.White,    Color.DarkGray),
                HotNormal = Attr(Color.Gray,     Color.Black),
                HotFocus  = Attr(Color.White,    Color.DarkGray),
                Disabled  = Attr(Color.Gray,     Color.Black),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.White,    Color.Black),
                Focus     = Attr(Color.White,    Color.DarkGray),
                HotNormal = Attr(Color.Gray,     Color.Black),
                HotFocus  = Attr(Color.White,    Color.DarkGray),
                Disabled  = Attr(Color.DarkGray, Color.DarkGray),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.Gray,     Color.DarkGray),
                Focus     = Attr(Color.White,    Color.Gray),
                HotNormal = Attr(Color.Gray,     Color.DarkGray),
                HotFocus  = Attr(Color.White,    Color.Gray),
                Disabled  = Attr(Color.DarkGray, Color.DarkGray),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.BrightRed, Color.Black),
                Focus     = Attr(Color.White,     Color.BrightRed),
                HotNormal = Attr(Color.BrightRed, Color.Black),
                HotFocus  = Attr(Color.White,     Color.BrightRed),
                Disabled  = Attr(Color.DarkGray,  Color.Black),
            },
        };

        // =====================================================================
        // Light — from config. TG v2 color mappings for v1:
        //   DimGray → DarkGray, WhiteSmoke → White, Gainsboro → Gray,
        //   LightGray → Gray, Silver → Gray, FireBrick → Red,
        //   LightCoral → BrightRed. "None" background → White.
        // =====================================================================
        private static Theme MakeLight() => new Theme
        {
            Name = "Light",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.DarkGray, Color.White),
                Focus     = Attr(Color.Black,    Color.Gray),
                HotNormal = Attr(Color.Gray,     Color.White),
                HotFocus  = Attr(Color.Black,    Color.Gray),
                Disabled  = Attr(Color.Gray,     Color.White),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.Gray,     Color.White),
                Focus     = Attr(Color.Black,    Color.Gray),
                HotNormal = Attr(Color.Gray,     Color.White),
                HotFocus  = Attr(Color.Black,    Color.Gray),
                Disabled  = Attr(Color.Gray,     Color.White),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.DarkGray, Color.White),
                Focus     = Attr(Color.Black,    Color.Gray),
                HotNormal = Attr(Color.Gray,     Color.White),
                HotFocus  = Attr(Color.Black,    Color.Gray),
                Disabled  = Attr(Color.Gray,     Color.White),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.DarkGray, Color.White),
                Focus     = Attr(Color.Black,    Color.Gray),
                HotNormal = Attr(Color.Gray,     Color.White),
                HotFocus  = Attr(Color.Black,    Color.Gray),
                Disabled  = Attr(Color.Gray,     Color.White),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.Red,       Color.White),
                Focus     = Attr(Color.White,     Color.Red),
                HotNormal = Attr(Color.BrightRed, Color.White),
                HotFocus  = Attr(Color.White,     Color.Red),
                Disabled  = Attr(Color.Gray,      Color.White),
            },
        };

        // =====================================================================
        // Light+ — original "Light" theme from before the config migration.
        // Higher-contrast selection (White/Blue focus) compared to Light.
        // =====================================================================
        private static Theme MakeLightPlus() => new Theme
        {
            Name = "Light+",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.Blue,      Color.White),
                Focus     = Attr(Color.White,     Color.Blue),
                HotNormal = Attr(Color.Red,       Color.White),
                HotFocus  = Attr(Color.BrightRed, Color.Blue),
                Disabled  = Attr(Color.DarkGray,  Color.White),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.Blue,     Color.White),
                Focus     = Attr(Color.White,     Color.Blue),
                HotNormal = Attr(Color.Red,       Color.Gray),
                HotFocus  = Attr(Color.BrightRed, Color.Blue),
                Disabled  = Attr(Color.Black,  Color.Gray),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.Black,     Color.White),
                Focus     = Attr(Color.White,     Color.Blue),
                HotNormal = Attr(Color.Red,       Color.White),
                HotFocus  = Attr(Color.BrightRed, Color.Blue),
                Disabled  = Attr(Color.DarkGray,  Color.White),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.Black,     Color.White),
                Focus     = Attr(Color.White,     Color.Blue),
                HotNormal = Attr(Color.Red,       Color.White),
                HotFocus  = Attr(Color.BrightRed, Color.Blue),
                Disabled  = Attr(Color.DarkGray,  Color.White),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.Red,      Color.White),
                Focus     = Attr(Color.White,    Color.Red),
                HotNormal = Attr(Color.Black,    Color.White),
                HotFocus  = Attr(Color.White,    Color.BrightRed),
                Disabled  = Attr(Color.DarkGray, Color.White),
            },
        };

        // =====================================================================
        // Green Phosphor — from config. GreenPhosphor → BrightGreen,
        //   Charcoal → DarkGray, Onyx → Black, None → Black.
        //   Focus is inverted (Black/BrightGreen) so list selection is visible.
        // =====================================================================
        private static Theme MakeGreenPhosphor() => new Theme
        {
            Name = "Green Phosphor",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.BrightGreen, Color.DarkGray),
                Focus     = Attr(Color.Black,       Color.BrightGreen),
                HotNormal = Attr(Color.BrightGreen, Color.DarkGray),
                HotFocus  = Attr(Color.Black,       Color.BrightGreen),
                Disabled  = Attr(Color.DarkGray,    Color.Black),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.BrightGreen, Color.DarkGray),
                Focus     = Attr(Color.Black,       Color.BrightGreen),
                HotNormal = Attr(Color.BrightGreen, Color.Black),
                HotFocus  = Attr(Color.Black,       Color.BrightGreen),
                Disabled  = Attr(Color.BrightGreen,    Color.Black),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.Black,       Color.BrightGreen),
                Focus     = Attr(Color.BrightGreen, Color.Black),
                HotNormal = Attr(Color.Black,       Color.BrightGreen),
                HotFocus  = Attr(Color.BrightGreen, Color.Black),
                Disabled  = Attr(Color.DarkGray,    Color.BrightGreen),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.Black,       Color.BrightGreen),
                Focus     = Attr(Color.BrightGreen, Color.Black),
                HotNormal = Attr(Color.Black,       Color.BrightGreen),
                HotFocus  = Attr(Color.BrightGreen, Color.Black),
                Disabled  = Attr(Color.DarkGray,    Color.BrightGreen),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.BrightGreen, Color.Black),
                Focus     = Attr(Color.Black,       Color.BrightGreen),
                HotNormal = Attr(Color.BrightGreen, Color.Black),
                HotFocus  = Attr(Color.Black,       Color.BrightGreen),
                Disabled  = Attr(Color.DarkGray,    Color.DarkGray),
            },
        };

        // =====================================================================
        // Amber Phosphor — from config. AmberPhosphor → BrightYellow,
        //   same structure as Green Phosphor.
        //   Focus is inverted (Black/BrightYellow) so list selection is visible.
        // =====================================================================
        private static Theme MakeAmberPhosphor() => new Theme
        {
            Name = "Amber Phosphor",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.BrightYellow, Color.Black),
                Focus     = Attr(Color.Black,        Color.BrightYellow),
                HotNormal = Attr(Color.BrightYellow, Color.Black),
                HotFocus  = Attr(Color.Black,        Color.BrightYellow),
                Disabled  = Attr(Color.DarkGray,     Color.Black),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.BrightYellow, Color.Black),
                Focus     = Attr(Color.Black,        Color.BrightYellow),
                HotNormal = Attr(Color.BrightYellow, Color.Black),
                HotFocus  = Attr(Color.Black,        Color.BrightYellow),
                Disabled  = Attr(Color.DarkGray,     Color.DarkGray),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.Black,        Color.BrightYellow),
                Focus     = Attr(Color.BrightYellow, Color.Black),
                HotNormal = Attr(Color.Black,        Color.BrightYellow),
                HotFocus  = Attr(Color.BrightYellow, Color.Black),
                Disabled  = Attr(Color.DarkGray,     Color.BrightYellow),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.Black,        Color.BrightYellow),
                Focus     = Attr(Color.BrightYellow, Color.Black),
                HotNormal = Attr(Color.Black,        Color.BrightYellow),
                HotFocus  = Attr(Color.BrightYellow, Color.Black),
                Disabled  = Attr(Color.DarkGray,     Color.BrightYellow),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.BrightYellow, Color.Black),
                Focus     = Attr(Color.Black,        Color.BrightYellow),
                HotNormal = Attr(Color.BrightYellow, Color.Black),
                HotFocus  = Attr(Color.Black,        Color.BrightYellow),
                Disabled  = Attr(Color.DarkGray,     Color.DarkGray),
            },
        };

        // =====================================================================
        // 8-Bit — from config. Strictly black-and-white. Config Disabled uses
        // Black/Black (invisible) — mapped to DarkGray/Black for usability.
        // =====================================================================
        private static Theme Make8Bit() => new Theme
        {
            Name = "8-Bit",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.White,    Color.Black),
                Focus     = Attr(Color.Black,    Color.White),
                HotNormal = Attr(Color.Black,    Color.White),
                HotFocus  = Attr(Color.White,    Color.Black),
                Disabled  = Attr(Color.DarkGray, Color.Black),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.White,    Color.Black),
                Focus     = Attr(Color.Black,    Color.White),
                HotNormal = Attr(Color.Black,    Color.White),
                HotFocus  = Attr(Color.White,    Color.Black),
                Disabled  = Attr(Color.White, Color.Black),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.Black,    Color.White),
                Focus     = Attr(Color.White,    Color.Black),
                HotNormal = Attr(Color.White,    Color.Black),
                HotFocus  = Attr(Color.Black,    Color.White),
                Disabled  = Attr(Color.DarkGray, Color.White),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.Black,    Color.White),
                Focus     = Attr(Color.White,    Color.Black),
                HotNormal = Attr(Color.White,    Color.Black),
                HotFocus  = Attr(Color.Black,    Color.White),
                Disabled  = Attr(Color.DarkGray, Color.White),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.White,    Color.Black),
                Focus     = Attr(Color.Black,    Color.White),
                HotNormal = Attr(Color.Black,    Color.White),
                HotFocus  = Attr(Color.White,    Color.Black),
                Disabled  = Attr(Color.DarkGray, Color.Black),
            },
        };

        // =====================================================================
        // Solarized (Light) — light variant. Białe tło z cyjanowymi akcentami
        // oddającymi charakterystyczny "tealowy" klimat palety.
        //   Base3 (#fdf6e3, main bg)       → White
        //   Base2 (#eee8d5, bg highlights) → Gray
        //   Base0 (#657b83, body text)     → DarkGray
        //   Yellow (#b58900)               → Brown
        //   Cyan (#2aa198, selection)      → Cyan
        //   Red (#dc322f)                  → Red
        // =====================================================================
        private static Theme MakeSolarized() => new Theme
        {
            Name = "Solarized",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.DarkGray, Color.White),
                Focus     = Attr(Color.Black,    Color.Cyan),
                HotNormal = Attr(Color.Brown,    Color.White),
                HotFocus  = Attr(Color.Black,    Color.Cyan),
                Disabled  = Attr(Color.DarkGray, Color.Gray),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.DarkGray, Color.White),
                Focus     = Attr(Color.Black,    Color.Cyan),
                HotNormal = Attr(Color.Brown,    Color.White),
                HotFocus  = Attr(Color.Black,    Color.Cyan),
                Disabled  = Attr(Color.DarkGray, Color.White),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.DarkGray, Color.Gray),
                Focus     = Attr(Color.Black,    Color.Cyan),
                HotNormal = Attr(Color.Brown,    Color.Gray),
                HotFocus  = Attr(Color.Black,    Color.Cyan),
                Disabled  = Attr(Color.DarkGray, Color.Gray),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.DarkGray, Color.White),
                Focus     = Attr(Color.Black,    Color.Cyan),
                HotNormal = Attr(Color.Brown,    Color.White),
                HotFocus  = Attr(Color.Black,    Color.Cyan),
                Disabled  = Attr(Color.DarkGray, Color.White),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.Red,       Color.White),
                Focus     = Attr(Color.White,     Color.Red),
                HotNormal = Attr(Color.BrightRed, Color.White),
                HotFocus  = Attr(Color.White,     Color.BrightRed),
                Disabled  = Attr(Color.DarkGray,  Color.White),
            },
        };

        // =====================================================================
        // Solarized Dark — charakterystyczne tealowe tło (Base03: #002b36).
        // W 16-kolorowej palecie TG v1 Color.Cyan jest najbliższe ciemnemu
        // tealowi Solarized (terminale z Solarized renderują "Black" jako
        // #002b36, ale w standardowej palecie Cyan daje najbardziej zbliżony
        // wizualny efekt).
        //   Tło główne      → Cyan  (teal)
        //   Tło wyróżnień   → Blue  (ciemniejszy teal)
        //   Tekst           → White / Gray
        //   Hotkey/akcent   → BrightYellow (#b58900 yellow)
        //   Selekcja Focus  → Black/BrightYellow (charakterystyczny Solarized)
        //   Błędy           → BrightRed
        // =====================================================================
        private static Theme MakeSolarizedDark() => new Theme
        {
            Name = "Solarized Dark",
            TopLevel = new ColorScheme
            {
                Normal    = Attr(Color.White,        Color.Cyan),
                Focus     = Attr(Color.Black,        Color.BrightYellow),
                HotNormal = Attr(Color.BrightYellow, Color.Cyan),
                HotFocus  = Attr(Color.Black,        Color.BrightYellow),
                Disabled  = Attr(Color.DarkGray,     Color.Cyan),
            },
            Base = new ColorScheme
            {
                Normal    = Attr(Color.White,        Color.Cyan),
                Focus     = Attr(Color.Black,        Color.BrightYellow),
                HotNormal = Attr(Color.BrightYellow, Color.Cyan),
                HotFocus  = Attr(Color.Black,        Color.BrightYellow),
                Disabled  = Attr(Color.White,     Color.Cyan),
            },
            Menu = new ColorScheme
            {
                Normal    = Attr(Color.White,        Color.Blue),
                Focus     = Attr(Color.Black,        Color.BrightYellow),
                HotNormal = Attr(Color.BrightYellow, Color.Blue),
                HotFocus  = Attr(Color.Black,        Color.BrightYellow),
                Disabled  = Attr(Color.DarkGray,     Color.Blue),
            },
            Dialog = new ColorScheme
            {
                Normal    = Attr(Color.White,        Color.Cyan),
                Focus     = Attr(Color.Black,        Color.BrightYellow),
                HotNormal = Attr(Color.BrightYellow, Color.Cyan),
                HotFocus  = Attr(Color.Black,        Color.BrightYellow),
                Disabled  = Attr(Color.DarkGray,     Color.Cyan),
            },
            Error = new ColorScheme
            {
                Normal    = Attr(Color.BrightRed,    Color.Cyan),
                Focus     = Attr(Color.White,        Color.BrightRed),
                HotNormal = Attr(Color.BrightRed,    Color.Cyan),
                HotFocus  = Attr(Color.White,        Color.BrightRed),
                Disabled  = Attr(Color.DarkGray,     Color.Cyan),
            },
        };
    }
}
