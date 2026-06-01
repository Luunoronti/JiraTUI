using Terminal.Gui;

namespace JiraTUI.Themes
{
    /// <summary>
    /// Bundle of all the ColorSchemes that drive the app's appearance.
    /// Each scheme is built explicitly so Focus/Normal contrast is guaranteed
    /// in every context (Base, Menu, Dialog, Error, TopLevel).
    /// </summary>
    public class Theme
    {
        public string Name { get; set; }
        public ColorScheme TopLevel { get; set; }
        public ColorScheme Base { get; set; }
        public ColorScheme Menu { get; set; }
        public ColorScheme Dialog { get; set; }
        public ColorScheme Error { get; set; }
    }
}
