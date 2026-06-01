using JiraTUI.Config;
using Terminal.Gui;

namespace JiraTUI.UI.Dialogs
{
    /// <summary>
    /// Lets the user pick which columns are visible in the issue list.
    /// </summary>
    public class ColumnsDialog : Dialog
    {
        public bool Saved { get; private set; }
        public ColumnVisibilityConfig Result { get; private set; }

        private readonly CheckBox _key, _type, _priority, _status, _assignee, _summary;

        public ColumnsDialog(ColumnVisibilityConfig current) : base("Visible columns", 44, 16)
        {
            var c = current ?? new ColumnVisibilityConfig();

            Add(new Label("Zaznacz kolumny do wyświetlenia:") { X = 1, Y = 0 });

            _key      = new CheckBox("Key")             { X = 2, Y = 2, Checked = c.Key };
            _type     = new CheckBox("Type (icon)")     { X = 2, Y = 3, Checked = c.Type };
            _priority = new CheckBox("Priority (icon)") { X = 2, Y = 4, Checked = c.Priority };
            _status   = new CheckBox("Status")          { X = 2, Y = 5, Checked = c.Status };
            _assignee = new CheckBox("Assignee")        { X = 2, Y = 6, Checked = c.Assignee };
            _summary  = new CheckBox("Summary")         { X = 2, Y = 7, Checked = c.Summary };
            Add(_key, _type, _priority, _status, _assignee, _summary);

            var save = new Button("Save");
            save.Clicked += OnSave;
            var cancel = new Button("Cancel");
            cancel.Clicked += () => { Saved = false; Application.RequestStop(); };
            AddButton(cancel);
            AddButton(save);

            Loaded += () => _key.SetFocus();
        }

        private void OnSave()
        {
            if (!_key.Checked && !_type.Checked && !_priority.Checked &&
                !_status.Checked && !_assignee.Checked && !_summary.Checked)
            {
                MessageBox.ErrorQuery("Invalid", "Wybierz przynajmniej jedną kolumnę.", "OK");
                return;
            }
            Result = new ColumnVisibilityConfig
            {
                Key = _key.Checked,
                Type = _type.Checked,
                Priority = _priority.Checked,
                Status = _status.Checked,
                Assignee = _assignee.Checked,
                Summary = _summary.Checked,
            };
            Saved = true;
            Application.RequestStop();
        }
    }
}
