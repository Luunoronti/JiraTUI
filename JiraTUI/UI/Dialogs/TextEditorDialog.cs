using Terminal.Gui;

namespace JiraTUI.UI.Dialogs
{
    /// <summary>
    /// Multi-line text editor used for editing description and adding comments.
    /// Big TextView with Save/Cancel.
    /// </summary>
    public class TextEditorDialog : Dialog
    {
        public bool Saved { get; private set; }
        public string Result { get; private set; }

        private readonly TextView _editor;

        public TextEditorDialog(string title, string initialText, string okLabel)
            : base(title, ComputeWidth(), ComputeHeight())
        {
            _editor = new TextView
            {
                X = 0,
                Y = 0,
                Width = Dim.Fill(),
                Height = Dim.Fill() - 2,
                ReadOnly = false,
                WordWrap = true,
                Text = initialText ?? "",
            };
            Add(_editor);

            var ok = new Button(okLabel ?? "Save", true);
            ok.Clicked += () =>
            {
                Saved = true;
                Result = _editor.Text != null ? _editor.Text.ToString() : "";
                Application.RequestStop();
            };
            var cancel = new Button("Cancel");
            cancel.Clicked += () =>
            {
                Saved = false;
                Application.RequestStop();
            };
            AddButton(cancel);
            AddButton(ok);
        }

        private static int ComputeWidth()
        {
            var drv = Application.Driver;
            int cols = drv != null ? drv.Cols : 100;
            return System.Math.Max(60, System.Math.Min(cols - 4, 120));
        }

        private static int ComputeHeight()
        {
            var drv = Application.Driver;
            int rows = drv != null ? drv.Rows : 30;
            return System.Math.Max(15, System.Math.Min(rows - 4, 40));
        }
    }
}
