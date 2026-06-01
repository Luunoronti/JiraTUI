using Terminal.Gui;

namespace JiraTUI.UI.Dialogs
{
    /// <summary>
    /// Asks for filter name + description before POSTing /rest/api/3/filter.
    /// </summary>
    public class SaveFilterDialog : Dialog
    {
        public bool Saved { get; private set; }
        public string FilterName { get; private set; }
        public string FilterDescription { get; private set; }

        public SaveFilterDialog(string jqlPreview) : base("Save current JQL as filter", 78, 14)
        {
            Add(new Label("JQL:") { X = 1, Y = 0 });
            var jqlLabel = new Label(Truncate(jqlPreview, 70))
            {
                X = 14, Y = 0, Width = Dim.Fill() - 2,
            };
            Add(jqlLabel);

            Add(new Label("Name:") { X = 1, Y = 2 });
            var nameField = new TextField("") { X = 14, Y = 2, Width = Dim.Fill() - 2 };
            Add(nameField);

            Add(new Label("Description:") { X = 1, Y = 4 });
            var descField = new TextField("") { X = 14, Y = 4, Width = Dim.Fill() - 2 };
            Add(descField);

            Add(new Label("Filter zostanie zapisany w Jirze i oznaczony jako Favourite.")
                { X = 1, Y = 6 });

            var save = new Button("Save");
            save.Clicked += () =>
            {
                var name = nameField.Text != null ? nameField.Text.ToString().Trim() : "";
                if (string.IsNullOrEmpty(name))
                {
                    MessageBox.ErrorQuery("Name required", "Filter musi mieć nazwę.", "OK");
                    return;
                }
                FilterName = name;
                FilterDescription = descField.Text != null ? descField.Text.ToString() : "";
                Saved = true;
                Application.RequestStop();
            };
            var cancel = new Button("Cancel");
            cancel.Clicked += () => { Saved = false; Application.RequestStop(); };
            AddButton(cancel);
            AddButton(save);

            Loaded += () => nameField.SetFocus();
        }

        private static string Truncate(string s, int max)
        {
            if (string.IsNullOrEmpty(s)) return "";
            return s.Length <= max ? s : s.Substring(0, max - 1) + "…";
        }
    }
}
