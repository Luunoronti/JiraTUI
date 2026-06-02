## v0.4.4-alpha

- Kolumna ⊙ dla schowanych issues w trybie "pokaż wszystkie" (Ctrl-Y)
- Poprawka: deadlock przy naciśnięciu Update w dialogu aktualizacji (Ctrl-U)
- Nowy glyf ▷ dla statusów Ready/Queued (np. "Ready For Build")
- Ulepszony matching statusów — więcej niestandardowych nazw rozpoznanych poprawnie
- F1: What's New (ten ekran)

## v0.4.0-alpha

- Ctrl-H: lokalne ukrywanie issues (plik issue-meta.json obok configu)
- Ctrl-Y: przełączanie widoku normalny / pokaż wszystkie (w tym schowane)
- Status bar: liczba schowanych issues widoczna w nagłówku

## v0.3.1-alpha

- Pasek JQL: wieloliniowy (4 wiersze), bez tekstu pomocniczego — samo query
- Status bar: kolorowe litery skrótów klawiszowych bez nawiasów

## v0.3.0-alpha

- Nowe opcje kolumny statusu: glyph zamiast tekstu (○◐◑✕✓⊘?)
- Nagłówki kolumn usunięte — więcej miejsca na dane
- JiraTUI przeniesiony do status bara (prawy koniec)
- Krok 11: generowanie JQL przez AI (Ctrl-G); adaptery Anthropic i OpenAI-compatible
- Krok 12: automatyczne sprawdzanie aktualizacji co 15 minut (Ctrl-U)

## v0.2.0-alpha

- Krok 10: dialog ustawień (F2) — Connection, Appearance, Behavior, AI
- Krok 9: mutacje issues: zmiana priorytetu, statusu, assignee, opisu, komentarzy
- Dialogi: Choice, TextEditor, Assignee, SaveFilter, Columns, Legend
- Prawdziwy klient Jira Cloud (REST API v3 + fallback v2)
- Poprawka parsowania dat Jira (+0000 zamiast +00:00)
- ADF renderer: opisy i komentarze renderowane jako tekst

## v0.1.0-alpha

- Pasek JQL (Ctrl-J) z historią, Enter submituje, ↑↓ nawigacja po historii
- Panel nawigacji (Ctrl-B): Quick Views, Projekty (expandable), Filtry
- Panel szczegółów: side panel (Ctrl-D) i fullscreen (Enter)
- 8 motywów kolorystycznych (TrueColor / 256 / 16-kolor)
- Lista issues z glifami typów, priorytetu, statusu
- Mock client (demo bez Jiry) i real Jira Cloud client
