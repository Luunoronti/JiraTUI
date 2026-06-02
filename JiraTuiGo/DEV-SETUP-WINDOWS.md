# Windows dev environment — JiraTUI (Go)

Everything needed to build, run, and develop JiraTUI on Windows.
No admin rights required except for the Go installer.

---

## 1. Install Go

Download the Windows installer from **https://go.dev/dl/**
(`go1.22.x.windows-amd64.msi` or newer).

Run the installer — it adds Go to `PATH` automatically.

Verify:
```powershell
go version
# → go version go1.22.x windows/amd64
```

Default install location: `C:\Program Files\Go`
Workspace (modules cache): `%USERPROFILE%\go`

---

## 2. Install Git

Download from **https://git-scm.com/download/win** and install with defaults.

Verify:
```powershell
git --version
# → git version 2.x.x.windows.x
```

---

## 3. Install VS Code + Go extension

1. Download VS Code from **https://code.visualstudio.com/**
2. Open VS Code, press `Ctrl+Shift+X`, search for **Go** (publisher: Go Team at Google)
3. Click Install

On first open of a `.go` file, VS Code will prompt to install Go tools
(`gopls`, `dlv`, `staticcheck`, etc.) — click **Install All**.

Recommended additional extensions:
- **GitLens** — enhanced git integration
- **Error Lens** — inline error display

---

## 4. Clone the repository

```powershell
# Replace with your Gitea URL and username
git clone https://YOUR-GITEA-URL/YOUR-USERNAME/jiratui.git
cd jiratui
```

---

## 5. Download dependencies

```powershell
go mod download
```

This fetches all modules listed in `go.mod` into the local cache.
No internet access needed after this step (until `go.mod` changes).

---

## 6. Run the application

```powershell
go run .
```

On first run with no config: the mock Jira client loads automatically,
showing 60+ sample issues. No Jira credentials needed to develop.

Press `Ctrl-Q` to exit.

---

## 7. Build a local binary

```powershell
# Windows 64-bit (your machine)
go build -o jiratui.exe .

# Cross-compile for Linux 64-bit
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o jiratui-linux-amd64 .; Remove-Item Env:\GOOS, Env:\GOARCH

# Cross-compile for Linux 32-bit
$env:GOOS="linux"; $env:GOARCH="386"; go build -o jiratui-linux-386 .; Remove-Item Env:\GOOS, Env:\GOARCH
```

---

## 8. Terminal recommendation

The default Windows Console (cmd.exe / PowerShell) has limited Unicode and
color support. Use one of these instead:

| Terminal | Notes |
|---|---|
| **Windows Terminal** | Best option. Install from Microsoft Store or winget. Supports true color, Unicode, all glyphs. |
| **Windows Terminal Preview** | Same, with newest features. |
| **VS Code integrated terminal** | Good for development — already open when coding. |

**Windows Terminal** — install via winget:
```powershell
winget install Microsoft.WindowsTerminal
```

**Font recommendation:** Cascadia Code or Cascadia Mono (installed with
Windows Terminal automatically). These include all Unicode glyphs used by
JiraTUI (⊘ ✓ ★ ⬢ ⇈ ▲ ○ ◐ etc.).

Enable true color in Windows Terminal: it is on by default.

Verify true color support in the terminal:
```powershell
$env:COLORTERM
# should output: truecolor
```
If empty, add to your PowerShell profile:
```powershell
$env:COLORTERM = "truecolor"
```

---

## 9. Environment variables (optional)

JiraTUI reads these at startup — none are required for development:

| Variable | Purpose |
|---|---|
| `COLORTERM` | Set to `truecolor` to force true color mode |
| `NO_COLOR` | Set to any value to disable all color output |
| `JIRATUI_CONFIG` | Override config file path (default: `%AppData%\JiraTuiGo\config.json`) |

---

## 10. Useful go commands during development

```powershell
# Run with race detector (catches concurrency bugs)
go run -race .

# Run all tests
go test ./...

# Check for common mistakes
go vet ./...

# Format all code (run before committing)
go fmt ./...

# List all dependencies
go list -m all

# Add a new dependency
go get github.com/some/package

# Remove unused dependencies
go mod tidy

# Build with version info embedded (simulates a release build)
go build -ldflags "-X main.version=dev -X main.commit=local" -o jiratui.exe .
```

---

## 11. Install goreleaser (optional — only for testing releases locally)

goreleaser is only needed if you want to test the release pipeline locally
before pushing to Gitea. It is not needed for normal development.

```powershell
winget install goreleaser.goreleaser
```

Verify:
```powershell
goreleaser --version
```

Dry-run release locally (does not publish anything):
```powershell
goreleaser release --snapshot --clean
# → builds all three binaries in ./dist/
```

---

## 12. Troubleshooting

### `go: command not found`
Go is not in PATH. Re-run the installer or add `C:\Program Files\Go\bin`
to your user PATH manually.

### Glyphs show as `?` or boxes
The terminal font does not include the Unicode codepoints used by JiraTUI.
Switch to Windows Terminal with Cascadia Code font.

### True color not working
Set `$env:COLORTERM = "truecolor"` in your PowerShell profile (`$PROFILE`).
JiraTUI will fall back to 256-color automatically if not set.

### `go run .` opens and immediately closes
The TUI requires an interactive terminal. Run from Windows Terminal or the
VS Code terminal, not from a script or CI environment.

### Keyring errors on first run
Windows Credential Manager may prompt for permission the first time a
credential is stored. Click "Allow". If keyring is unavailable, JiraTUI
falls back to base64 storage with a warning — this is safe for development.

### Port already in use / firewall prompts
JiraTUI makes outbound HTTPS connections only (to Jira, Anthropic/OpenAI,
Gitea). No inbound ports. Dismiss any firewall prompt by allowing outbound
access.
