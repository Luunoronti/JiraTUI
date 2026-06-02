# JiraTUI for Go — Application Specification

> Cross-platform Jira Cloud TUI client written in Go.
> Targets: Windows 64-bit, Linux 64-bit, Linux 32-bit.
> Based on the C# JiraTUI reference implementation.

---

## Tech stack

| Layer | Library |
|---|---|
| TUI framework | [tview](https://github.com/rivo/tview) + tcell v2 |
| HTTP client | stdlib `net/http` |
| JSON | stdlib `encoding/json` |
| Config storage | stdlib `os` / `encoding/json` |
| Secret storage | [99designs/keyring](https://github.com/99designs/keyring) (OS keychain / fallback to encrypted file) |
| Browser open | [pkg/browser](https://github.com/pkg/browser) |
| AI (Claude) | Anthropic REST API v1 (direct HTTP) |

---

## Project layout

```
JiraTuiGo/
├── main.go
├── go.mod
├── go.sum
├── config/
│   ├── config.go          # AppConfig struct + Load/Save
│   ├── secrets.go         # Keyring wrapper (protect/unprotect tokens)
│   └── jql_history.go     # JQL search history (max 200, deduplicated)
├── jira/
│   ├── client.go          # IJiraClient interface
│   ├── real_client.go     # Real Jira Cloud REST v3 implementation
│   ├── mock_client.go     # Offline demo client (60+ fake issues)
│   ├── adf_renderer.go    # Atlassian Document Format → plain text
│   └── models.go          # Issue, Comment, Project, JiraUser, Transition, etc.
├── ai/
│   ├── client.go          # Anthropic /v1/messages client
│   └── jql_prompt.go      # System prompt builder with live Jira context
├── ui/
│   ├── app.go             # Application bootstrap, tview.Application
│   ├── main_window.go     # Root layout + keyboard routing
│   ├── nav_view.go        # Left navigation tree
│   ├── issue_list.go      # Center issue table
│   ├── issue_detail.go    # Right detail panel
│   ├── jql_bar.go         # Bottom JQL input bar
│   ├── glyphs.go          # Unicode glyphs + color definitions
│   └── dialogs/
│       ├── settings.go        # Settings dialog (Connection/Appearance/Behavior/AI)
│       ├── choice.go          # Generic list picker (priority, status, transitions)
│       ├── assignee.go        # Assignee search dialog
│       ├── text_editor.go     # Multi-line editor (description, comments)
│       ├── columns.go         # Column visibility picker
│       ├── legend.go          # Glyph legend (read-only)
│       ├── ai_jql.go          # AI JQL generation dialog
│       └── save_filter.go     # Save current JQL as Jira filter
└── themes/
    ├── theme.go           # Theme struct (name + 5 color schemes)
    └── manager.go         # BuildAll, Apply, AvailableThemes, CurrentThemeName
```

---

## Responsive layout

The app must be fully usable at any realistic terminal size, from a small
4:3 window to an ultrawide monitor. Layout adapts continuously as the terminal
is resized — no restart needed.

### Size tiers

| Tier | Width | Typical context |
|---|---|---|
| **Compact** | < 100 cols | Small window, 4:3 screen, SSH in small pane |
| **Normal** | 100 – 159 cols | Standard terminal, 16:9 at moderate size |
| **Wide** | ≥ 160 cols | Large monitor, maximised on widescreen |

### Issue list column behaviour by tier

| Column | Compact | Normal | Wide |
|---|---|---|---|
| Key | 10 | 10 | 10 |
| Type glyph | 3 | 3 | 3 |
| Priority glyph | 3 | 3 | 3 |
| Status | hidden if needed¹ | max 15 | max 15 |
| Assignee | hidden if needed¹ | max 12 | max 15 |
| Summary | fills rest (min 20) | fills rest | fills rest |

¹ Status and Assignee are hidden automatically (in that order) when the
remaining width after Key + Type + Priority + Summary-minimum would be
negative. User-configured column visibility is respected first; auto-hiding
only triggers for columns the user has not explicitly hidden.

Summary always has a minimum of 20 characters. If even that cannot fit,
the app shows a warning row instead of crashing.

### Navigation panel — width cap by tier

| Tier | Max nav panel width |
|---|---|
| Compact | 60% of terminal width |
| Normal | 40 cols |
| Wide | 48 cols |

The panel never exceeds these caps regardless of content length — long node
labels are truncated with `…`.

### Detail side panel — width by tier

| Tier | Detail panel width |
|---|---|
| Compact | not available (Ctrl-D does nothing; use Enter for fullscreen) |
| Normal | 40% of terminal width |
| Wide | 38% of terminal width (more space for the list) |

In Compact tier, pressing Ctrl-D opens the fullscreen detail view instead
of the side panel, and a brief status bar hint explains why.

### Height adaptations

| Available rows | Behaviour |
|---|---|
| < 12 | Show a "terminal too small" message and nothing else |
| 12 – 20 | Status bar shows one line only (key shortcuts abbreviated) |
| 21 – 30 | Normal layout |
| > 30 | No change — extra rows go to the issue list and detail views |

### Dialog sizing

All dialogs are sized relative to the current terminal dimensions:

| Dimension | Rule |
|---|---|
| Max dialog width | `min(requested, termWidth - 4)` |
| Max dialog height | `min(requested, termHeight - 4)` |
| Min dialog width | 36 cols (below this a dialog cannot render usably) |
| Choice dialog height | `min(optionCount + 6, termHeight - 6)` |
| Text editor width | `max(50, min(120, termWidth - 8))` |
| Text editor height | `max(10, min(40, termHeight - 8))` |

If the terminal is too small to display a dialog at minimum size, the dialog
does not open and the status bar shows: `Terminal too small for this dialog`.

### Status bar behaviour

| Available width | Behaviour |
|---|---|
| ≥ 80 cols | Full hints: `F2:Settings  Ctrl-R:Refresh  Ctrl-Q:Quit  …` |
| 60 – 79 cols | Abbreviated: `F2  Ctrl-R  Ctrl-Q  …` |
| < 60 cols | Only the most critical: `Ctrl-Q:Quit` |

The update indicator (`↑ v1.x.x — Ctrl-U`) is always shown at the far right
as long as there is at least 20 cols of space after other status bar content.

### JQL bar

Full width at all times. On very narrow terminals the hint text inside the
bar is hidden, leaving only the input field.

### Resize handling

- tview calls `Draw()` on every resize event automatically
- Column widths, panel widths, and dialog sizes are all recomputed on each draw
- No cached fixed sizes — everything is a function of current terminal dimensions
- The selected row in the issue list is preserved across resizes

### Minimum supported size

| Dimension | Minimum |
|---|---|
| Width | 60 cols |
| Height | 12 rows |

Below these values the app shows only: `Terminal too small (WxH). Please resize.`

---

## UI layout

The issue list is the **only permanent main view** — it always occupies the full
terminal width. There are three ways to see issue details and two kinds of overlays:

| View | Trigger | How it opens | Dismissal |
|---|---|---|---|
| Detail side panel | Ctrl-D | Floating overlay, list stays active | Ctrl-D or Escape |
| **Detail fullscreen** | **Enter** | Replaces the list entirely | **Escape** |
| Navigation panel | Ctrl-B | Floating overlay, captures input | Ctrl-B or Escape |

### Default state (no overlays open)

```
┌─[Menu: File | View | JQL | Issue | Help]────────────────────────────────────┐
│                                                                               │
│  Key        T P  Status        Assignee     Summary                         │
│  ─────────────────────────────────────────────────────────────────────────  │
│  KEY-1      ✓ ▲  In Progress   John Doe     Fix login timeout on mobile     │
│  KEY-2      ⊘ ⇈  To Do         Jane Smith   Critical payment gateway error  │
│  KEY-3      ★ ─  In Review     John Doe     Add dark mode to settings       │
│  ...                                                                          │
│                                                                               │
│                                                                               │
├───────────────────────────────────────────────────────────────────────────────┤
│  JQL: assignee = currentUser() AND statusCategory != Done  (Ctrl-J · ↑↓)    │
├───────────────────────────────────────────────────────────────────────────────┤
│  F2:Settings  F5:Refresh  Ctrl-Q:Quit  Ctrl-G:AI  Ctrl-L:Legend             │
└───────────────────────────────────────────────────────────────────────────────┘
```

### Navigation panel overlay (Ctrl-B to open/close)

Appears as a **floating window on top of the issue list** — the list beneath
does not move or resize. Takes full keyboard input while open — the issue list
is non-interactive until the panel is closed. Width is whatever is needed to
fit the content comfortably (not constrained to a fixed percentage).
Pressing Ctrl-B or Escape closes it and returns focus to the issue list.

**Opening Nav automatically closes the Detail side panel** (if it was open).
The reverse does not apply — opening Detail does not affect Nav, but since Nav
captures all input, Detail cannot be opened while Nav is active anyway.

```
┌─[Menu]──────────────────────────────────────────────────────────────────────┐
│                                                                               │
│ ┌─Navigation──────────────────┐                                              │
│ │ Quick Views                 │ T P  Status      Assignee   Summary          │
│ │   My Issues                 │─────────────────────────────────────────────│
│ │   Reported by me            │ ✓ ▲  In Progress  John Doe   Fix login...   │
│ │   Recently updated          │ ⊘ ⇈  To Do        Jane S.    Critical pay...│
│ │   All issues                │ ★ ─  In Review    John Doe   Add dark mod...│
│ │                             │                                              │
│ │ Projects                    │                                              │
│ │ ▼ PROJ                      │                                              │
│ │     Backlog                 │                                              │
│ │     In Progress             │                                              │
│ │     Done                    │                                              │
│ │ ▶ INFRA                     │                                              │
│ │ ▶ DEV                       │                                              │
│ │                             │                                              │
│ │ Filters                     │                                              │
│ │   My open issues            │                                              │
│ │   Critical bugs             │                                              │
│ └─────────────────────────────┘                                              │
├───────────────────────────────────────────────────────────────────────────────┤
│  JQL: ...                                                                    │
└───────────────────────────────────────────────────────────────────────────────┘
```

### Detail panel overlay (Ctrl-D to open/close)

Slides in from the right and covers ~40% of the screen. Does **not** capture
keyboard input — the issue list remains focused and interactive. Moving the
cursor in the list immediately updates the detail panel content.
Pressing Ctrl-D or Escape closes it.

```
┌─[Menu]──────────────────────────────────────────────────────────────────────┐
│                                                                               │
│  Key     T P  Status      Assignee  ┌─Detail──────────────────────────────┐ │
│  ──────────────────────────────────  │ KEY-1                               │ │
│  KEY-1   ✓ ▲  In Progress John Doe  │ Task · High · In Progress           │ │
│  KEY-2   ⊘ ⇈  To Do       Jane S.  │ Assignee: John Doe                  │ │
│  KEY-3   ★ ─  In Review   John Doe  │ Reporter: Jane Smith                │ │
│  ...                                │ Updated: 2025-06-01 14:32           │ │
│                                     │ Labels: mobile, auth                │ │
│                                     │                                     │ │
│                                     │ Description:                        │ │
│                                     │ Login sessions expire too quickly   │ │
│                                     │ on mobile devices. Token refresh... │ │
│                                     │                                     │ │
│                                     │ Comments:                           │ │
│                                     │ [John D.] 2025-05-30               │ │
│                                     │ Reproduced on iOS 17.               │ │
│                                     └─────────────────────────────────────┘ │
├───────────────────────────────────────────────────────────────────────────────┤
│  JQL: ...                                                Ctrl-D: close det  │
└───────────────────────────────────────────────────────────────────────────────┘
```

### Detail fullscreen view (Enter to open, Escape to close)

Pressing Enter on a selected issue **replaces the issue list** with a full-screen
detail view. The list is completely hidden while this view is active.
All issue mutation actions are available here with the same shortcuts as
everywhere else (Ctrl-P, Ctrl-T, Ctrl-A, Ctrl-E, Ctrl-K, Ctrl-O).
Pressing Escape closes the fullscreen view and returns to the issue list
with the same selection.

```
┌─[Menu]──────────────────────────────────────────────────────────────────────┐
│                                                                               │
│  KEY-1  ·  Task  ·  High  ·  In Progress                                    │
│  ─────────────────────────────────────────────────────────────────────────  │
│  Assignee:  John Doe                                                          │
│  Reporter:  Jane Smith                                                        │
│  Updated:   2025-06-01 14:32                                                  │
│  Labels:    mobile, auth                                                      │
│  Sprint:    Sprint 42                                                         │
│                                                                               │
│  Description:                                                                 │
│  Login sessions expire too quickly on mobile devices. The token refresh       │
│  mechanism does not account for background app state on iOS. Users are        │
│  logged out unexpectedly after ~15 minutes of inactivity.                     │
│                                                                               │
│  Comments:                                                                    │
│  ── John Doe  ·  2025-05-30 09:12 ─────────────────────────────────────────  │
│  Reproduced on iOS 17.4 and 17.5. Android not affected.                      │
│                                                                               │
│  ── Jane Smith  ·  2025-05-31 11:45 ───────────────────────────────────────  │
│  Fix in PR #892, please review.                                               │
│                                                                               │
├───────────────────────────────────────────────────────────────────────────────┤
│  Esc:Back  Ctrl-P:Priority  Ctrl-T:Status  Ctrl-A:Assignee  Ctrl-E:Desc      │
│            Ctrl-K:Comment   Ctrl-O:Browser                                   │
└───────────────────────────────────────────────────────────────────────────────┘
```

### Overlay / view interaction rules

| Action | Result |
|---|---|
| Ctrl-B (open Nav) | Closes Detail side panel if open; Nav appears on top of list |
| Ctrl-D (open Detail side panel) | Nav must be closed first (Nav captures all input) |
| Enter (open fullscreen detail) | List is replaced; all overlays (Nav, Detail, JQL) close |
| Escape in fullscreen detail | Returns to list; restores previous overlay state |

### Overlay behaviour summary

The issue list never moves or resizes. Nav and Detail side panel float on top.
Fullscreen detail replaces the list entirely.

| Panel / View | Position | Sizing | Input | Dismiss |
|---|---|---|---|---|
| Navigation | top-left floating | fits content | **Captures all input** | Ctrl-B or Escape |
| Detail side panel | top-right floating | ~40% width | Passive — list keeps focus | Ctrl-D or Escape |
| Detail fullscreen | full screen | replaces list | **Full input** | Escape |
| JQL bar | bottom strip | full width | Captures input while open | Enter, Ctrl-J, or Escape (see below) |

**JQL bar Escape behaviour:**
- Escape on non-empty bar → clears the text, bar stays open
- Escape on empty bar → closes the bar (same as Ctrl-J)

**Defaults on startup:**
- Navigation: hidden
- Detail side panel: hidden
- JQL bar: hidden (Ctrl-J shows and focuses it)

---

## Keyboard shortcuts

No Alt shortcuts — everything uses Ctrl or function keys.

### Global

| Shortcut | Action |
|---|---|
| Ctrl-Q | Quit |
| Ctrl-B | Toggle Navigation panel |
| Ctrl-D | Toggle Detail panel |
| Ctrl-J | Toggle JQL bar / focus it |
| Ctrl-G | Open AI JQL dialog |
| Ctrl-O | Open selected issue in browser |
| Ctrl-L | Show Legend (glyph meanings) |
| Ctrl-R | Refresh current search |
| F2 | Open Settings |

### JQL bar (when open)

| Shortcut | Action |
|---|---|
| Enter | Run query and close bar |
| Ctrl-J | Close bar |
| Escape | Clear text (if non-empty) / close bar (if empty) |
| ↑ / ↓ | Navigate JQL history |

### Issue list (issue list focused, issue selected)

| Shortcut | Action |
|---|---|
| Enter | Open fullscreen detail view |
| Ctrl-P | Change Priority |
| Ctrl-T | Change Status (Transition) |
| Ctrl-A | Change Assignee |
| Ctrl-E | Edit Description |
| Ctrl-K | Add Comment |
| Ctrl-O | Open in Browser |
| Ctrl-F | Save current JQL as Filter |

### Fullscreen detail view

| Shortcut | Action |
|---|---|
| Escape | Close, return to issue list |
| Ctrl-P | Change Priority |
| Ctrl-T | Change Status (Transition) |
| Ctrl-A | Change Assignee |
| Ctrl-E | Edit Description |
| Ctrl-K | Add Comment |
| Ctrl-O | Open in Browser |

### Dialogs

| Shortcut | Action |
|---|---|
| Enter | Confirm / OK |
| Escape | Cancel / close |
| Down (in search field) | Jump to result list (AssigneeDialog) |

---

## Issue list columns

| Column | Header | Width | Notes |
|---|---|---|---|
| Key | Key | 10 | e.g. PROJ-123 |
| Type | (blank) | 3 | Unicode glyph + color |
| Priority | (blank) | 3 | Unicode glyph + color |
| Status | Status | max 15 | truncated |
| Assignee | Assignee | max 12 | truncated |
| Summary | Summary | fills rest | truncated to fit |

All columns individually toggleable via View → Columns (or settings dialog).
Summary always auto-expands to fill remaining terminal width.

---

## Glyphs

### Issue types

| Glyph | Color | Types |
|---|---|---|
| ⊘ | Red | Bug |
| ✓ | Cyan | Task |
| ★ | Green | Story |
| ⬢ | Magenta | Epic |
| ↳ | Gray | Sub-task |
| ⚒ | Blue | Improvement |
| ✦ | Yellow | New Feature |
| ⌬ | Brown | Test |
| ‼ | BrightRed | Incident |
| ✉ | Cyan | Service Request |
| ? | Gray | Unknown |

### Priorities

| Glyph | Color | Priorities |
|---|---|---|
| ⇈ | BrightRed | Highest, Critical, Blocker |
| ▲ | Red | High, Major |
| ─ | (none) | Medium |
| ▼ | Cyan | Low, Minor |
| ⇊ | DarkGray | Lowest, Trivial |

### Statuses

| Glyph | Statuses |
|---|---|
| ○ | To Do, Open, Backlog |
| ◐ | In Progress |
| ◑ | In Review, Testing, QA |
| ✕ | Blocked, On Hold |
| ✓ | Done, Closed, Resolved |
| ⊘ | Cancelled, Won't Do |
| ? | Unknown |

---

## Jira API

**Target:** Jira Cloud REST API v3
**Auth:** HTTP Basic — `base64(email:apiToken)`
**Base URL:** `https://<workspace>.atlassian.net`

### Endpoints used

| Method | Path | Purpose |
|---|---|---|
| GET | /rest/api/3/myself | Test connection + get current user |
| POST | /rest/api/3/issue/search | Search issues by JQL |
| GET | /rest/api/3/issue/{key} | Get single issue (full fields) |
| GET | /rest/api/3/project/search | List projects (paginated) |
| GET | /rest/api/3/priority | List priorities |
| GET | /rest/api/3/status | List statuses |
| GET | /rest/api/3/issuetype | List issue types |
| GET | /rest/api/3/filter/my | List user's saved filters |
| GET | /rest/api/3/issue/{key}/transitions | List available transitions |
| POST | /rest/api/3/issue/{key}/transitions | Execute a transition |
| GET | /rest/api/3/user/assignable/search | Search assignable users |
| PUT | /rest/api/3/issue/{key} | Update priority or assignee |
| PUT | /rest/api/3/issue/{key} | Update description (ADF) |
| POST | /rest/api/3/issue/{key}/comment | Add comment |
| POST | /rest/api/3/filter | Save new filter |

### Issue search fields requested
```
*navigable,description,comment,priority,status,assignee,reporter,
issuetype,project,summary,labels,sprint
```

### ADF (Atlassian Document Format)
Jira descriptions and comments arrive as ADF JSON.
Must render to readable plain text.

Supported node types:
- **Block**: paragraph, heading (prefix `#`/`##`/`###`), bulletList, orderedList,
  codeBlock (fenced with ` ``` `), blockquote (prefix `>`), rule (`─────`)
- **Inline**: text, hardBreak, mention (`@Name`), emoji, link/inlineCard

---

## Configuration

### Storage locations

| File | Path |
|---|---|
| Main config | `%AppData%\JiraTuiGo\config.json` (Windows) / `~/.config/jiratui/config.json` (Linux) |
| JQL history | `%Documents%\JiraTuiGo\jql-history.json` (Windows) / `~/Documents/jiratui/jql-history.json` (Linux) |

### Config structure

```json
{
  "connection": {
    "baseUrl": "https://workspace.atlassian.net",
    "email": "user@example.com",
    "tokenProtected": "<encrypted>",
    "authType": "ApiToken"
  },
  "appearance": {
    "themeName": "Default"
  },
  "behavior": {
    "defaultJql": "assignee = currentUser() AND statusCategory != Done",
    "pageSize": 50,
    "autoRefreshSeconds": 0
  },
  "ai": {
    "adapter": "anthropic",
    "baseUrl": "",
    "model": "claude-sonnet-4-5",
    "apiKeyProtected": "<encrypted>"
  },
  "columns": {
    "key": true,
    "type": true,
    "priority": true,
    "status": true,
    "assignee": true,
    "summary": true
  }
}
```

### Secret storage (cross-platform)

Tokens are never stored in plaintext.

| Platform | Storage |
|---|---|
| Windows | Windows Credential Manager via keyring |
| Linux | Secret Service (libsecret) via keyring; fallback: AES-256 encrypted file with key derived from machine ID |
| macOS | Keychain via keyring |

---

## JQL history

- Max 200 entries
- Deduplication by `originalText` (re-submitting moves to top)
- Tracks: `timestamp`, `originalText`, `effectiveJql`, `wasAiTranslated`
- Navigation: ↑ recalls older, ↓ recalls newer; going past newest clears bar

---

## AI — JQL generation

**Max tokens:** 1024
**Timeout:** 60s

### Adapters

Two adapters cover all supported providers:

#### `anthropic` adapter
Anthropic has its own API format (`/v1/messages`).

| Setting | Value |
|---|---|
| Base URL | `https://api.anthropic.com` |
| Endpoint | `/v1/messages` |
| Auth header | `x-api-key: <key>` |
| Default model | `claude-sonnet-4-5` |

#### `openai-compatible` adapter
OpenAI format (`/v1/chat/completions`). Used by all of the following —
only the base URL and model name differ:

| Provider | Base URL | API key |
|---|---|---|
| OpenAI | `https://api.openai.com` | required |
| Groq | `https://api.groq.com/openai` | required |
| Google Gemini | `https://generativelanguage.googleapis.com/v1beta/openai` | required |
| Azure OpenAI | `https://<resource>.openai.azure.com/openai/deployments/<model>` | required |
| Ollama (local) | `http://localhost:11434` | *(empty)* |
| LM Studio (local) | `http://localhost:1234` | *(empty)* |

### Config

```json
"ai": {
  "adapter":         "anthropic",
  "baseUrl":         "",
  "model":           "claude-sonnet-4-5",
  "apiKeyProtected": "<encrypted>"
}
```

- `adapter` — `"anthropic"` or `"openai-compatible"`
- `baseUrl` — leave empty to use the provider default; set for Azure, Ollama, LM Studio, etc.
- `model` — any model name supported by the chosen provider
- `apiKeyProtected` — encrypted; leave empty for local providers (Ollama, LM Studio)

### System prompt context (dynamic)
- Current user display name
- Available projects (Key + Name)
- Available statuses
- Available priorities
- Available issue types
- JQL syntax reference (operators, functions, date literals)

### Auto-fallback
If a user types something that is not valid JQL, the app automatically
sends it to the configured AI provider for translation before showing an error.

---

## Themes

Themes are fully custom — not tied to tview's internal color scheme structure.
Each theme defines **semantic color roles** for every UI element in the app.
tview/tcell supports true color (24-bit RGB), so themes are not limited to
the terminal's 16 basic colors.

### Color roles

```
Theme
│
├── Chrome (app frame)
│     background        main app background
│     border            panel/box borders (unfocused)
│     borderFocused     panel/box border when focused
│
├── Text
│     textNormal        default readable text
│     textMuted         secondary info (timestamps, reporters, hints)
│     textEmphasis      headings, issue keys, section labels
│
├── Issue list
│     listBg            list background
│     listFg            normal row text
│     listSelectedBg    selected row background   ← must contrast with listBg
│     listSelectedFg    selected row text
│     listHeaderBg      column header background
│     listHeaderFg      column header text
│
├── Navigation panel
│     navBg             panel background
│     navFg             item text
│     navSelectedBg     selected item background  ← must contrast with navBg
│     navSelectedFg     selected item text
│     navSectionFg      section headings ("Projects", "Filters")
│
├── Detail panel (side + fullscreen)
│     detailBg          background
│     detailFg          body text
│     detailLabelFg     field labels ("Assignee:", "Status:", …)
│     detailValueFg     field values
│
├── JQL bar
│     jqlBg             bar background
│     jqlFg             input text
│     jqlHintFg         hint text (history count, "Ctrl-J to close")
│
├── Status bar
│     statusBg          bar background
│     statusFg          normal text
│     statusKeyFg       keyboard shortcut highlights ("F2", "Ctrl-Q")
│     statusUpdateFg    update available indicator ("↑ v1.3.0 available")
│
├── Dialogs
│     dialogBg          dialog background
│     dialogFg          dialog text
│     dialogBorderFg    dialog border
│     buttonBg          normal button background
│     buttonFg          normal button text
│     buttonFocusBg     focused button background ← must contrast with buttonBg
│     buttonFocusFg     focused button text
│     inputBg           text field background
│     inputFg           text field text
│     inputFocusBg      focused text field background
│
├── Issue type glyphs
│     typeBug           ⊘
│     typeTask          ✓
│     typeStory         ★
│     typeEpic          ⬢
│     typeSubtask       ↳
│     typeOther         ? (default for unknown types)
│
├── Priority glyphs
│     priHighest        ⇈  (Highest / Critical / Blocker)
│     priHigh           ▲  (High / Major)
│     priMedium         ─  (Medium)
│     priLow            ▼  (Low / Minor)
│     priLowest         ⇊  (Lowest / Trivial)
│
└── Status glyphs
      statusTodo        ○  (To Do / Open / Backlog)
      statusInProgress  ◐  (In Progress)
      statusInReview    ◑  (In Review / Testing / QA)
      statusBlocked     ✕  (Blocked / On Hold)
      statusDone        ✓  (Done / Closed / Resolved)
      statusCancelled   ⊘  (Cancelled / Won't Do)
```

### Terminal color capability detection

At startup the app detects the terminal's color depth and selects the best
available color representation for every theme color role.

**Detection order:**

1. `COLORTERM=truecolor` or `COLORTERM=24bit` → **true color (24-bit RGB)**
2. `TERM` contains `256color`, or tcell reports 256-color support → **256-color palette**
3. Fallback → **16 ANSI colors**

**How each theme handles the tiers:**

Each color role in a theme is defined as a struct with three values:

```go
type ThemeColor struct {
    TrueColor  tcell.Color   // 24-bit RGB, e.g. tcell.NewRGBColor(0x1e, 0x1e, 0x2e)
    Color256   tcell.Color   // nearest xterm-256 color, e.g. tcell.PaletteColor(235)
    Color16    tcell.Color   // nearest ANSI-16 color, e.g. tcell.ColorBlack
}
```

The theme manager reads the detected capability once at startup and returns the
appropriate field for every color lookup. No runtime branching per draw call.

**Fallback mapping strategy:**

- True color → 256: author-specified (pick the closest xterm-256 index by eye /
  euclidean distance in RGB space)
- 256 → 16: author-specified (pick the closest named ANSI color)
- 16-color themes (TurboPascal, High Contrast) look intentional at all tiers
  since they are designed around ANSI colors from the start

**The `High Contrast` theme is tier-independent** — it uses only the 8 standard
ANSI colors (black, white, red, green, yellow, blue, magenta, cyan) and looks
identical at all three tiers.

### Built-in themes

| Name | Style |
|---|---|
| **Dark** | Dark background, neutral grays, colored accents — default |
| **Light** | Light background, dark text |
| **TurboPascal** | Classic blue/cyan, Turbo Pascal 5 IDE look |
| **Green Phosphor** | Black background, green text, retro CRT monitor |
| **Amber Phosphor** | Black background, amber text, retro CRT monitor |
| **Solarized Dark** | Solarized dark palette |
| **Solarized Light** | Solarized light palette |
| **High Contrast** | Pure black/white + 8 ANSI colors, looks the same at all tiers |

### Rules enforced for all themes

- `listSelectedBg` ≠ `listBg` — selected row must be visible
- `navSelectedBg` ≠ `navBg` — selected nav item must be visible
- `buttonFocusBg` ≠ `buttonBg` — focused button must be visible
- `inputFocusBg` ≠ `inputBg` — focused input field must be visible
- Glyph colors must be readable on both `listBg` and `detailBg`
- All three tiers (true color, 256, 16) must satisfy the above rules

---

## Settings dialog

Four tabs:

### Connection
- Base URL (Jira workspace URL)
- Email
- API Token (masked input, paste button, character counter)
- [Test Connection] button — validates without saving
- Auth type: API Token / Basic / PAT

### Appearance
- Theme selector (live preview, Cancel reverts)

### Behavior
- Default JQL
- Page size (default 50)
- Auto-refresh interval in seconds (0 = disabled)

### AI
- Adapter: `anthropic` or `openai-compatible` (dropdown)
- Base URL — empty = provider default; set for Azure, Ollama, LM Studio
- Model name
- API Key (masked input, paste button, character counter; leave empty for local providers)

---

## Issue mutation actions

Accessible from Issue menu or direct shortcuts while an issue is selected.
No Alt shortcuts — all direct Ctrl bindings.

| Action | Shortcut | Dialog |
|---|---|---|
| Change Priority | Ctrl-P | ChoiceDialog with available priorities |
| Change Status (Transition) | Ctrl-T | ChoiceDialog with available transitions |
| Change Assignee | Ctrl-A | AssigneeDialog (search + list) |
| Edit Description | Ctrl-E | TextEditorDialog (multi-line) |
| Add Comment | Ctrl-K | TextEditorDialog (multi-line) |
| Open in Browser | Ctrl-O | System browser, constructs URL from base + key |
| Save as Filter | Ctrl-F | SaveFilterDialog (name + description) |

---

## Demo / offline mode

When no credentials are configured, a `MockJiraClient` is used automatically.

Mock data:
- 4 projects: PROJ, INFRA, DEV, QA
- 60+ randomly seeded issues
- 3 saved filters
- Basic JQL parsing (project =, assignee = currentUser(), type =, status =, statusCategory !=)
- Issues sorted by Updated descending

---

## Updates

### Version embedding

Every binary has version info embedded at build time via `goreleaser`:

```
version  — git tag, e.g. "v1.2.3"
commit   — short git SHA
date     — build date (UTC)
```

Displayed in `Help → About` and in the update notification.

### Update check cadence

- Checked **once per day** at most — last check timestamp stored in config
- Triggered on startup (in a background goroutine, never blocks the UI)
- No way to disable — the check is always active
- If the check fails (no network, Gitea unreachable) it is silently ignored;
  the next attempt happens the following day

### Update notification

When a newer version is found, a persistent indicator appears in the
**status bar**, styled with a distinct accent color (theme's `statusUpdateFg`
role, e.g. bright yellow on dark themes):

```
F2:Settings  Ctrl-R:Refresh  Ctrl-Q:Quit  …   ↑ v1.3.0 available — Ctrl-U
```

Rules:
- **Cannot be turned off** — no setting to suppress it
- Stays visible for the entire session once detected
- Does not pop up a dialog, does not steal focus, does not interrupt anything
- The `↑ v1.3.0 available — Ctrl-U` segment is always at the far right of the
  status bar so it does not push other hints around

### Update install (Ctrl-U)

1. A confirmation dialog appears:
   ```
   ┌─ Update available ──────────────────────────────┐
   │                                                  │
   │  Current:   v1.2.0  (2025-04-10)                │
   │  Available: v1.3.0  (2025-06-01)                │
   │                                                  │
   │  The application will be replaced and must be   │
   │  restarted to apply the update.                 │
   │                                                  │
   │            [ Update ]  [ Cancel ]               │
   └──────────────────────────────────────────────────┘
   ```
2. Download progress shown inline in the dialog (bytes / total, %)
3. SHA256 checksum verified against `checksums.txt` from the release
4. Binary replaced atomically; on Windows a trampoline handles the
   locked-file case
5. Dialog closes with message: **"Update applied. Please restart."**
6. Status bar indicator changes to: `↑ Restart to apply v1.3.0`

### Libraries

| Purpose | Library |
|---|---|
| Self-update + Gitea Releases | `github.com/creativeprojects/go-selfupdate` (has Gitea provider) |
| Release building + checksums | `goreleaser` ≥ v1.5 (CI only, not a runtime dep) |

### Update check — Gitea API

The app calls the Gitea releases API to find the latest version:

```
GET https://<gitea>/api/v1/repos/<owner>/<repo>/releases/latest
```

No authentication required if the Gitea repository is public or if Gitea
is configured to allow anonymous API access to releases.

The Gitea base URL, owner, and repo name are **embedded at build time**
(same `-ldflags` mechanism as version). Users do not configure this.

### Build & release pipeline

`goreleaser` runs in Gitea Actions on every version tag (`v*`) and produces:

```
jiratui-windows-amd64.exe
jiratui-linux-amd64
jiratui-linux-386
checksums.txt          ← SHA256 of all three binaries
```

All uploaded automatically as assets to the Gitea Release.
Full setup instructions: see `GITEA-SETUP.md` in the repository root.

---

## Out of scope (v1)

- Jira Server / Data Center (only Cloud REST v3)
- OAuth 2.0 authentication
- Issue creation (only editing existing issues)
- Attachment handling
- Board / sprint views
- Subtask management
- Bulk operations
- Notifications / webhooks

---

## Implementation plan

Each step produces a **compiling, runnable application**. Do not proceed to
the next step until the current one compiles cleanly and behaves as described.
All steps assume the dev environment described in `DEV-SETUP-WINDOWS.md` is
already in place.

---

### Step 1 — Project skeleton + config layer

**Goal:** runnable binary that immediately exits cleanly; config persists to disk.

Create:
- `go.mod` — module `jiratui`, Go 1.22
- `main.go` — parse flags (`--version`), load config, print version and exit
- `config/config.go` — `AppConfig` struct with all sub-configs
  (ConnectionConfig, AppearanceConfig, BehaviorConfig, AiConfig,
  ColumnVisibilityConfig); `Load()` / `Save()` using platform config dir
- `config/secrets.go` — `Protect(plain string) string` /
  `Unprotect(enc string) string` using `99designs/keyring`; graceful fallback
  to base64 if keyring unavailable (log warning, never crash)
- `config/jql_history.go` — `JqlHistory` with `Add()`, `GetByRecentIndex()`,
  `Load()`, `Save()`; max 200 entries, dedup by `OriginalText`

**Verify:**
```
go run . --version          → prints "jiratui dev"
go run .                    → exits 0, config file created in AppData
```

---

### Step 2 — Jira models + ADF renderer + MockClient

**Goal:** all Jira data structures exist; mock client returns realistic data
without any network call.

Create:
- `jira/models.go` — `Issue`, `Comment`, `Project`, `SavedFilter`,
  `Transition`, `JiraUser`
- `jira/client.go` — `Client` interface with all methods
- `jira/adf_renderer.go` — `RenderADF(rawJson string) string`; handles
  paragraph, heading, bulletList, orderedList, codeBlock, blockquote, rule,
  text, hardBreak, mention, link/inlineCard
- `jira/mock_client.go` — implements `Client`; 60+ seeded issues across
  4 projects (PROJ, INFRA, DEV, QA); 3 saved filters; simple JQL parsing
  (project=, assignee=currentUser(), type=, status=, statusCategory!=)

**Verify:**
```go
// quick smoke test in main.go temporarily:
c := jira.NewMockClient()
issues, _ := c.SearchIssues("project = PROJ", 20)
fmt.Println(len(issues)) // → 20
```

---

### Step 3 — Basic TUI shell

**Goal:** window opens, status bar visible, menu bar visible, empty issue list
area, Ctrl-Q quits. Resize is handled correctly from the very first frame.

Create:
- `ui/layout.go` — **foundation for all resize logic**:
  - `TermSize() (width, height int)` — reads live screen dimensions from
    `tview.Application` (never cached; called fresh on every draw)
  - `SizeTier(width int) Tier` — Compact / Normal / Wide
  - `TooSmall(width, height int) bool` — true if below 60×12
  - `StatusBarHints(width int, hints []Hint) string` — returns full / abbreviated /
    minimal hint string based on available width
  - All overlay primitives call `TermSize()` inside their `Draw()` method to
    reposition themselves; no overlay stores its own position or size

- `ui/app.go` — initialise `tview.Application`; wire Ctrl-Q global;
  register a resize callback via `app.SetBeforeDrawFunc`:
  - if `TooSmall` → switch to "too small" page, show message
  - if previously too small and now big enough → switch back to main page

- `ui/main_window.go` — root `tview.Pages` (not `tview.Flex`):
  - page `"main"` — `tview.Flex` (vertical): menu bar (1 row) + centre area +
    status bar (1 row)
  - page `"toosmall"` — centred message: `Terminal too small (WxH). Please resize.`
  - Status bar text recomputed on every draw via `StatusBarHints(termWidth, ...)`

- `ui/issue_list.go` — `tview.Table` with header row only, no data yet

- `themes/theme.go` — `ThemeColor` struct (TrueColor / Color256 / Color16);
  `Theme` struct with all color roles listed in SPEC
- `themes/manager.go` — `Detect()` sets color tier at startup; `Apply(theme)`
  sets tview's global styles; `Dark` theme only (others come in Step 4)

Wire: `main.go` calls `themes.Detect()`, `themes.Apply(Dark)`, then
`ui.Run(cfg, jiraClient)`.

**Verify:**
```
go run .   → TUI window opens, status bar shows hints, Ctrl-Q closes cleanly
resize terminal to very wide → status bar shows full hint text
resize terminal narrow → status bar abbreviates hints
shrink below 60×12 → "Terminal too small (WxH)" message appears instantly
grow back above 60×12 → normal UI reappears instantly, no restart
change font size in Windows Terminal → same resize events fire, same behaviour
```

---

### Step 4 — Theme system (all themes)

**Goal:** all 8 themes implemented; live switching works.

Extend `themes/manager.go`:
- Implement all 8 themes: Dark, Light, TurboPascal, Green Phosphor,
  Amber Phosphor, Solarized Dark, Solarized Light, High Contrast
- Each color role has all three tiers (TrueColor / Color256 / Color16)
- `Switch(name string)` re-applies theme to all live tview primitives
- `AvailableThemes() []string` returns names in display order

Add temporary theme switcher to status bar (cycle with a key) to verify visually.

**Verify:**
```
go run .   → cycle through all 8 themes; each must have clearly visible
             selected-row highlight in the (empty) table
```

---

### Step 5 — Issue list with mock data + glyphs

**Goal:** 60 mock issues visible in the list with correct columns, glyphs,
and colors; column widths recomputed on every draw; Summary fills remaining
width; columns auto-hide/show as terminal is resized.

Extend `ui/issue_list.go`:
- Connect `MockClient.SearchIssues(defaultJql, pageSize)` on startup
- `ui/glyphs.go` — all glyph + color definitions from SPEC
- Column visibility driven by `ColumnVisibilityConfig`; Summary always shown

**Column width strategy — computed on every draw, never stored:**

  Implement `issueList` as a custom primitive (embed `*tview.Box`):
  - `Draw(screen tcell.Screen)` calls `layout.TermSize()` to get current width
  - Computes active columns from config + auto-hide rules (see Responsive
    layout section in SPEC) each time `Draw()` is called
  - Builds a `[]ColumnDef{id, fixedWidth}` slice; Summary gets the remainder
  - Renders each row by calling `screen.SetContent()` directly, or rebuilds
    the underlying `tview.Table` cell content before delegating to its `Draw()`

  This means a mid-session window resize (or font size change in Windows
  Terminal) causes the very next frame to show the correct column layout —
  no user action needed.

- Cell text is truncated to its computed column width with `…` if overflowing
- Header row redrawn with same widths on every frame

**Verify:**
```
go run .   → 60 issues visible; ↑↓ navigate; type/priority glyphs colored
grab terminal window corner and drag to resize continuously →
  Summary column expands/shrinks smoothly on every frame
narrow below 100 cols → Status column disappears automatically
narrow further → Assignee column disappears automatically
widen again → both columns reappear at correct widths
change font size in Windows Terminal → same instant adaptation
selected row stays selected and visible throughout all resizes
```

---

### Step 6 — Navigation panel overlay

**Goal:** Ctrl-B opens a floating nav panel on top of the list; the list
beneath never moves; panel captures all input; resize while panel is open
repositions it correctly.

Create `ui/nav_view.go`:
- Implement `navPanel` as a custom primitive (embed `*tview.Box`):
  - `Draw(screen tcell.Screen)`:
    1. Calls `layout.TermSize()` to get current terminal dimensions
    2. Computes panel width: `min(longestLabel+4, layout.NavMaxWidth(tier))`
    3. Sets its own rect to `(x=0, y=1, w=panelWidth, h=termHeight-3)`
       (y=1 skips menu bar; h leaves room for status bar and JQL bar if open)
    4. Draws border, title, tree content within that rect
  - If terminal width < panelWidth + 10 (not enough room to see any list
    behind it): panel uses full width and shows a hint at the bottom:
    `[terminal too narrow — press Escape to close]`
- Nodes: Quick Views (4), Projects (from `Client.GetProjects()`), Filters
  (from `Client.GetSavedFilters()`); projects with Backlog / In Progress /
  Done sub-nodes; each node carries an embedded JQL string
- Selecting a node fires `OnSelect(jql)` → closes panel → updates issue list
- Managed via `tview.Pages` — page `"nav"` sits above page `"main"`
- Opening Nav closes the Detail side panel if it is open
- Ctrl-B and Escape close the panel and return focus to issue list

**Verify:**
```
go run .   → Ctrl-B opens nav; list visible behind it, unchanged
             resize terminal while nav is open → panel reflows to new height,
             stays anchored top-left, never overlaps menu or status bar
             narrow terminal while nav open → panel shrinks width to cap,
             shows "too narrow" hint at very small sizes
             Escape closes nav; focus returns to list at same selection
```

---

### Step 7 — Detail side panel + fullscreen detail

**Goal:** Ctrl-D shows side panel (passive, right-anchored); Enter shows
fullscreen; both reflow correctly on any resize.

Create `ui/issue_detail.go`:
- Shared `renderIssue(issue, width int) string` — formats all fields,
  description, comments; `width` is the available render width so text can
  be word-wrapped correctly for the current panel size

- **Side panel** — custom primitive (embed `*tview.Box`):
  - `Draw(screen tcell.Screen)`:
    1. Calls `layout.TermSize()` → computes `panelWidth = layout.DetailWidth(tier)`
    2. If tier is Compact: does not draw; the page is hidden (see below)
    3. Sets own rect to `(x=termWidth-panelWidth, y=1, w=panelWidth, h=termHeight-3)`
    4. Calls `renderIssue(current, panelWidth-2)` and draws wrapped text
  - No input capture; list keeps focus at all times
  - Updates content whenever list selection changes (callback from issue list)
  - Ctrl-D and Escape close it

- **Compact tier auto-switch:** when Ctrl-D is pressed and `SizeTier` is
  Compact, skip the side panel entirely and open the fullscreen view instead;
  show a one-time status bar hint: `Detail panel unavailable at this width — showing fullscreen`

- **Mid-session tier change:** if the side panel is open and the user resizes
  the terminal so the tier drops to Compact:
  - Side panel closes automatically
  - Status bar shows: `Terminal too narrow for side panel (Ctrl-D for fullscreen)`
  - If the tier rises back to Normal/Wide while in that state, the panel does
    NOT reopen automatically (user must press Ctrl-D again)

- **Fullscreen view** — `tview.Pages` page `"detail-full"`:
  - Replaces the list entirely; `renderIssue(issue, termWidth-2)` recomputed
    on every draw so text reflows as terminal is resized
  - All mutation shortcuts active (Ctrl-P/T/A/E/K/O)
  - Escape returns to list with same selection

**Verify:**
```
go run .   → Ctrl-D shows side panel anchored right; list still responds to ↑↓
             resize terminal while panel open → panel stays right-anchored,
             reflows text to new width, height fills available rows
             shrink to Compact tier → side panel closes, status bar hint shown
             grow back → hint goes away (panel stays closed until Ctrl-D again)
             Enter on issue → fullscreen; resize → text reflows in fullscreen too
             Escape → back to list, same row selected
```

---

### Step 8 — JQL bar + history

**Goal:** Ctrl-J shows/focuses JQL bar; Enter runs query; Escape clears/closes;
↑↓ recalls history.

Create `ui/jql_bar.go`:
- Single-line `tview.InputField` in a frame at the bottom
- Shown via `tview.Pages` on top of everything else
- Ctrl-J: if hidden → show + focus; if shown → close
- Enter: save to history, call `OnSubmit(jql)`, close bar
- Escape: if text non-empty → clear text, stay open;
          if text empty → close bar
- ↑: recall older history entry (index+1); ↓: recall newer (index-1);
  navigating past newest → clear field
- Hint text in bar: `Enter:run  ↑↓:history  Ctrl-J:close`

**Verify:**
```
go run .   → Ctrl-J opens bar; type JQL, Enter → list filters;
             ↑↓ navigate history; Escape clears then closes
```

---

### Step 9 — Issue mutation dialogs

**Goal:** all mutation shortcuts open the correct dialog; mock client
reflects changes in the list immediately.

Create `ui/dialogs/`:
- `choice.go` — `ChoiceDialog(title, options, initialIndex)` → selected index;
  `tview.List` inside a modal frame; Enter or double-click commits; Escape cancels
- `assignee.go` — `AssigneeDialog(client, issueKey)`;
  search field + `tview.List`; Enter in search calls
  `Client.SearchAssignableUsers()`; Down arrow moves focus to list;
  "(Unassign)" always first row; returns `accountId` or `""` for unassign
- `text_editor.go` — `TextEditorDialog(title, initial, buttonLabel)`;
  multi-line `tview.TextView` (editable); sized to 70% of terminal
- `columns.go` — `ColumnsDialog(current ColumnVisibilityConfig)`;
  checkboxes for each column; at least one must remain checked
- `legend.go` — read-only glyph reference; two-column layout
- `save_filter.go` — `SaveFilterDialog(jql)` → name + description

Wire shortcuts in `ui/main_window.go` and `ui/issue_detail.go`:
- Ctrl-P → priorities from `Client.GetPriorityNames()` → `ChoiceDialog` →
  `Client.SetPriority()` → refresh row
- Ctrl-T → transitions from `Client.GetTransitions()` → `ChoiceDialog` →
  `Client.TransitionIssue()` → refresh row
- Ctrl-A → `AssigneeDialog` → `Client.SetAssignee()` → refresh row
- Ctrl-E → `TextEditorDialog` → `Client.UpdateDescription()` → refresh detail
- Ctrl-K → `TextEditorDialog` → `Client.AddComment()` → refresh detail
- Ctrl-O → `pkg/browser` → open issue URL in system browser
- Ctrl-F → `SaveFilterDialog` → `Client.SaveFilter()` → reload nav filters

**Verify:**
```
go run .   → Ctrl-P opens priority picker; select one → row updates;
             Ctrl-A opens assignee search; type partial name → list filters;
             Ctrl-O → browser opens correct Jira URL (mock URL)
```

---

### Step 10 — Settings dialog + real Jira client

**Goal:** F2 opens settings; saving valid credentials connects to real Jira.

Create:
- `ui/dialogs/settings.go` — four tabs: Connection, Appearance, Behavior, AI;
  masked token fields with paste button + live char counter;
  [Test Connection] validates without saving; saving restarts Jira client
- `jira/real_client.go` — implements `Client` against Jira Cloud REST v3;
  Basic auth (`base64(email:apiToken)`); 30s timeout; all endpoints from SPEC;
  sprint field auto-detection; ADF rendering; caching for priorities/statuses/types
- `jira/factory.go` — `NewClient(cfg) Client`: returns `RealClient` if
  credentials present, `MockClient` otherwise

Replace mock client wiring in `main.go` with `jira.NewClient(cfg)`.

**Verify:**
```
go run .   → F2 → Connection tab → enter real Jira URL + email + token
             → Test Connection → shows your display name
             → Save → issue list loads real issues from Jira
```

---

### Step 11 — AI JQL generation

**Goal:** Ctrl-G opens AI dialog; natural language → JQL; auto-fallback on
invalid JQL.

Create:
- `ai/client.go` — `AiClient` interface; `Generate(system, user string) string`
- `ai/anthropic.go` — Anthropic adapter (`/v1/messages`)
- `ai/openai.go` — OpenAI-compatible adapter (`/v1/chat/completions`);
  used for OpenAI, Groq, Gemini, Ollama, LM Studio
- `ai/factory.go` — `NewAiClient(cfg AiConfig) AiClient`
- `ai/jql_prompt.go` — `BuildSystemPrompt(client JiraClient) string`;
  injects current user, projects, statuses, priorities, types, JQL syntax guide
- `ui/dialogs/ai_jql.go` — prompt textarea + Generate button + result textarea;
  Ctrl-G or Ctrl-Enter triggers generation; status label shows progress;
  "Use" button copies JQL to JQL bar

Auto-fallback: if `Client.SearchIssues()` returns a JQL parse error, silently
call AI to translate the original input, then retry.

**Verify:**
```
go run .   → Ctrl-G → type "bugs assigned to me updated this week"
             → Generate → JQL appears → Use → list filters correctly
```

---

### Step 12 — Update system

**Goal:** version embedded at build; daily update check; status bar indicator;
Ctrl-U downloads and installs.

- Embed version/commit/date/giteaURL/giteaOwner/giteaRepo via `-ldflags` in
  `main.go` package-level vars
- `updater/updater.go` — wraps `go-selfupdate` with Gitea provider;
  `CheckForUpdate() (latestVersion string, available bool, err error)`;
  `InstallUpdate(progressFn func(pct int)) error`
- On startup: launch goroutine; check if 24h have passed since last check
  (stored in config); call `CheckForUpdate()`; if available, set a flag that
  the status bar reads
- Status bar: if update available, append `↑ v1.x.x available — Ctrl-U`
  in `statusUpdateFg` color at far right; cannot be hidden
- Ctrl-U: open update confirmation dialog with current vs available version;
  on confirm, show download progress, verify SHA256, apply, prompt restart
- After apply: status bar changes to `↑ Restart to apply v1.x.x`

**Verify (local, without real Gitea):**
- Temporarily hard-code `available = true` and `latestVersion = "v99.0.0"` to
  verify the status bar indicator and Ctrl-U dialog work correctly before
  wiring to a real Gitea instance.

---

### Step 13 — Gitea CI/CD

**Goal:** pushing a `v*` tag triggers Actions, builds all three binaries,
creates a Gitea Release with assets.

Create in repository root:
- `.goreleaser.yml` — as described in `GITEA-SETUP.md`; fill in real
  Gitea URL / owner / repo
- `.gitea/workflows/release.yml` — trigger on `v*` tags, run goreleaser
- `.gitea/workflows/build.yml` — trigger on every push, verify compilation
  for all three targets

Update `main.go` `-ldflags` to match `.goreleaser.yml` variable names.

**Verify:**
```bash
git tag v0.1.0
git push origin v0.1.0
# → Gitea Actions runs → release appears with 3 binaries + checksums.txt
# → download binary, run it → Help → About shows v0.1.0
# → push v0.2.0 → run v0.1.0 binary → status bar shows update available
```
