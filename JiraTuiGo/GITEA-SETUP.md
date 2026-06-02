# Gitea — setup guide for JiraTUI CI/CD

This document covers everything needed to go from a fresh Gitea instance to
a fully working build + release + auto-update pipeline for JiraTUI.

---

## Prerequisites

| Requirement | Minimum version | Notes |
|---|---|---|
| Gitea | 1.19 | First version with Actions support |
| act_runner | latest | Gitea's own Actions runner |
| Docker | any recent | Used by act_runner to run build jobs |
| Go (on runner) | 1.22 | Only needed if not using Docker images |

---

## 1. Enable Actions in Gitea

Log in as Gitea admin and edit `app.ini`:

```ini
[actions]
ENABLED = true
```

Restart Gitea. The "Actions" tab will appear in repositories.

---

## 2. Install and register act_runner

### 2a. Download act_runner

```bash
# Linux (on the machine that will run builds)
wget https://gitea.com/gitea/act_runner/releases/latest/download/act_runner-linux-amd64
chmod +x act_runner-linux-amd64
mv act_runner-linux-amd64 /usr/local/bin/act_runner
```

### 2b. Get a registration token

In Gitea: **Admin panel → Actions → Runners → Create new runner**

Copy the displayed token.

### 2c. Register the runner

```bash
act_runner register \
  --no-interactive \
  --instance https://YOUR-GITEA-URL \
  --token    YOUR-REGISTRATION-TOKEN \
  --name     build-runner \
  --labels   ubuntu-latest:docker://node:20-bookworm
```

The label `ubuntu-latest:docker://node:20-bookworm` makes workflows that
request `runs-on: ubuntu-latest` work — act_runner will spin up that Docker
image for each job.

### 2d. Run act_runner as a service (Linux/systemd)

```ini
# /etc/systemd/system/act_runner.service
[Unit]
Description=Gitea act_runner
After=network.target

[Service]
ExecStart=/usr/local/bin/act_runner daemon
WorkingDirectory=/var/lib/act_runner
Restart=always
User=act_runner

[Install]
WantedBy=multi-user.target
```

```bash
useradd -r -s /bin/false act_runner
mkdir -p /var/lib/act_runner
chown act_runner /var/lib/act_runner
systemctl enable --now act_runner
```

---

## 3. Create the Gitea repository

1. Create a new repository in Gitea (can be private — releases API still
   works without auth if Gitea's `REQUIRE_SIGNIN_VIEW` is false)
2. Push the JiraTUI source code
3. In repository **Settings → Actions**: confirm Actions are enabled

---

## 4. Create a Gitea token for goreleaser

goreleaser needs permission to create releases and upload assets.

In Gitea: **User settings → Applications → Generate new token**

- Token name: `goreleaser`
- Permissions: `write:repository` (includes releases)

Copy the token. **You will not see it again.**

Add it as a repository secret:

**Repository → Settings → Actions → Secrets → Add secret**

| Name | Value |
|---|---|
| `GORELEASER_GITEA_TOKEN` | *(paste token here)* |

---

## 5. Add the goreleaser config

Create `.goreleaser.yml` in the repository root:

```yaml
# .goreleaser.yml
version: 2

project_name: jiratui

gitea_urls:
  api:      https://YOUR-GITEA-URL/api/v1
  download: https://YOUR-GITEA-URL
  skip_tls_verify: false          # set true only if using self-signed cert

before:
  hooks:
    - go mod tidy

builds:
  - id: jiratui
    env:
      - CGO_ENABLED=0
    goos:
      - windows
      - linux
    goarch:
      - amd64
      - "386"
    ignore:
      - goos: windows
        goarch: "386"             # no 32-bit Windows target
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.ShortCommit}}
      - -X main.date={{.Date}}
      - -X main.giteaURL=https://YOUR-GITEA-URL
      - -X main.giteaOwner=YOUR-USERNAME
      - -X main.giteaRepo=jiratui

archives:
  - id: binaries
    format: binary              # plain binaries, no .tar.gz
    name_template: >-
      {{ .ProjectName }}-
      {{- .Os }}-
      {{- .Arch }}

checksum:
  name_template: checksums.txt
  algorithm: sha256

release:
  gitea:
    owner: YOUR-USERNAME
    name:  jiratui
  name_template: "v{{ .Version }}"
  draft: false
  prerelease: auto              # tags like v1.0.0-beta become pre-releases

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

Replace `YOUR-GITEA-URL`, `YOUR-USERNAME` with real values.

---

## 6. Add the Gitea Actions workflow

Create `.gitea/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0          # goreleaser needs full git history

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run goreleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GORELEASER_GITEA_TOKEN: ${{ secrets.GORELEASER_GITEA_TOKEN }}
          GITEA_TOKEN: ${{ secrets.GORELEASER_GITEA_TOKEN }}
```

Also create `.gitea/workflows/build.yml` for CI on every push (no release,
just verifies the code compiles):

```yaml
name: Build

on:
  push:
    branches: ['**']
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build (Linux amd64)
        run: GOOS=linux GOARCH=amd64 go build -o /dev/null ./...

      - name: Build (Linux 386)
        run: GOOS=linux GOARCH=386 go build -o /dev/null ./...

      - name: Build (Windows amd64)
        run: GOOS=windows GOARCH=amd64 go build -o /dev/null ./...
```

---

## 7. First release — step by step

```bash
# 1. Make sure everything compiles locally
go build ./...

# 2. Commit and push all changes (including .goreleaser.yml and workflows)
git add .
git commit -m "chore: add goreleaser and Gitea Actions"
git push

# 3. Tag the first release
git tag v1.0.0
git push origin v1.0.0
```

Gitea Actions picks up the tag, runs goreleaser, and creates a release at:
```
https://YOUR-GITEA-URL/YOUR-USERNAME/jiratui/releases/tag/v1.0.0
```

Assets attached to the release:
```
jiratui-windows-amd64.exe
jiratui-linux-amd64
jiratui-linux-386
checksums.txt
```

---

## 8. Anonymous release access (no login required for downloads)

For auto-update to work without any credentials, Gitea must allow anonymous
access to the API and release assets.

In Gitea `app.ini`:

```ini
[service]
REQUIRE_SIGNIN_VIEW = false     # allow browsing without login

[api]
ENABLE_SWAGGER  = true
```

If the repository itself must stay private but releases should be public,
the simplest solution is a **separate public Gitea repository** used only
for releases — goreleaser can target any repo, not necessarily the source repo.

Alternatively, embed a read-only API token in the binary (scoped to
`read:repository` only) for the update check call. This token is not a
secret — it only allows reading public release metadata.

---

## 9. Self-signed TLS certificate (if applicable)

If your Gitea uses a self-signed certificate:

1. Set `skip_tls_verify: true` in `.goreleaser.yml` (build machine only)
2. In the app's updater code, configure the HTTP client to trust the cert:

```go
// embed the CA cert in the binary at build time
//go:embed gitea-ca.crt
var giteaCACert []byte
```

Or distribute the CA cert alongside the binary and load it at runtime.

---

## 10. Verify everything works

```bash
# Check that the release appeared
curl https://YOUR-GITEA-URL/api/v1/repos/YOUR-USERNAME/jiratui/releases/latest

# Expected: JSON with tag_name, assets array containing the three binaries
```

Run the app and confirm the status bar shows version info (`Help → About`).
Wait until next day (or temporarily lower the check interval in dev builds)
to verify the update notification appears when a newer tag is pushed.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| Actions tab missing | Actions not enabled in `app.ini` | Add `[actions] ENABLED=true`, restart Gitea |
| Runner shows "offline" | act_runner not running | `systemctl status act_runner` |
| goreleaser: "401 Unauthorized" | Wrong or missing token | Re-check `GORELEASER_GITEA_TOKEN` secret |
| goreleaser: "no such host" | Wrong Gitea URL in `.goreleaser.yml` | Check `gitea_urls.api` value |
| Update check silently fails | Gitea requires login | Set `REQUIRE_SIGNIN_VIEW=false` |
| TLS errors on update check | Self-signed cert | Embed CA cert or set `InsecureSkipVerify` (dev only) |
| Windows binary won't replace itself | Binary is running | goreleaser's trampoline handles this — restart app first |
