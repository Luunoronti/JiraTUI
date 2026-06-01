using System.Collections.Generic;
using System.Linq;
using Terminal.Gui;

namespace JiraTUI.UI.Dialogs
{
    /// <summary>
    /// Modal "pick one from list" dialog. Backed by a ListView so arrow keys
    /// navigate naturally and Enter commits the selection. Returns the chosen
    /// index via SelectedIndex (-1 on cancel).
    /// </summary>
    public class ChoiceDialog : Dialog
    {
        public int SelectedIndex { get; private set; } = -1;

        private readonly ListView _list;
        private readonly IList<string> _options;

        public ChoiceDialog(string title, IList<string> options, int initialIndex)
            : base(title, ComputeWidth(options), ComputeHeight(options))
        {
            _options = options;

            _list = new ListView(_options.ToList())
            {
                X = 1,
                Y = 0,
                Width = Dim.Fill() - 2,
                Height = Dim.Fill() - 2,
            };
            if (initialIndex >= 0 && initialIndex < _options.Count)
                _list.SelectedItem = initialIndex;
            _list.OpenSelectedItem += _ => Commit();
            Add(_list);

            var ok = new Button("OK", true);
            ok.Clicked += Commit;
            var cancel = new Button("Cancel");
            cancel.Clicked += () =>
            {
                SelectedIndex = -1;
                Application.RequestStop();
            };
            AddButton(cancel);
            AddButton(ok);

            Loaded += () => _list.SetFocus();
        }

        private void Commit()
        {
            SelectedIndex = _list.SelectedItem;
            Application.RequestStop();
        }

        private static int ComputeWidth(IList<string> opts)
        {
            int w = 30;
            foreach (var s in opts)
                if (s != null && s.Length + 6 > w) w = s.Length + 6;
            return System.Math.Min(w, 80);
        }

        private static int ComputeHeight(IList<string> opts) =>
            System.Math.Min(opts.Count + 6, 20);
    }
}
