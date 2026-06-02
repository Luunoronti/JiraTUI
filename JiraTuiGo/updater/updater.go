package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Release holds info about an available GitHub release.
type Release struct {
	Version     string // e.g. "v0.2.1"
	DownloadURL string // URL of the binary asset for current OS/arch
	ChecksumURL string // URL of checksums.txt asset
}

// AssetName returns the platform-specific asset filename for the current OS/arch.
// Returns "" for unsupported platforms.
func AssetName() string {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		return "jiratui-linux-amd64"
	case "linux/386":
		return "jiratui-linux-386"
	case "windows/amd64":
		return "jiratui-windows-amd64.exe"
	default:
		return ""
	}
}

// githubRelease is the JSON shape returned by the GitHub API.
type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// Check fetches the latest GitHub release and returns it if newer than currentVersion.
// Returns (nil, nil) if already up to date or the platform is unsupported.
// All network errors are returned as non-nil error (caller should ignore them).
func Check(owner, repo, currentVersion string) (*Release, error) {
	assetName := AssetName()
	if assetName == "" {
		return nil, nil // unsupported platform
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "jiratui/"+currentVersion)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}

	if !isNewer(rel.TagName, currentVersion) {
		return nil, nil
	}

	// Find download URL and checksum URL.
	var downloadURL, checksumURL string
	for _, a := range rel.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
		}
		if a.Name == "checksums.txt" {
			checksumURL = a.BrowserDownloadURL
		}
	}
	if downloadURL == "" {
		return nil, nil // asset not published for this platform
	}

	return &Release{
		Version:     rel.TagName,
		DownloadURL: downloadURL,
		ChecksumURL: checksumURL,
	}, nil
}

// isNewer returns true when latestTag is semantically newer than currentVersion.
// If currentVersion is "dev" it always returns true (useful for testing).
func isNewer(latestTag, currentVersion string) bool {
	if currentVersion == "dev" {
		return true
	}
	latest := strings.TrimPrefix(latestTag, "v")
	current := strings.TrimPrefix(currentVersion, "v")
	if latest == current {
		return false
	}
	return semverGT(latest, current)
}

// semverGT returns true when a > b using simple numeric component comparison.
func semverGT(a, b string) bool {
	aParts := strings.SplitN(a, ".", 3)
	bParts := strings.SplitN(b, ".", 3)
	// Pad to same length.
	for len(aParts) < 3 {
		aParts = append(aParts, "0")
	}
	for len(bParts) < 3 {
		bParts = append(bParts, "0")
	}
	for i := 0; i < 3; i++ {
		av := numericPrefix(aParts[i])
		bv := numericPrefix(bParts[i])
		if av > bv {
			return true
		}
		if av < bv {
			return false
		}
	}
	return false
}

// numericPrefix parses the leading integer of a semver component (handles pre-release suffixes).
func numericPrefix(s string) int {
	for i, c := range s {
		if c < '0' || c > '9' {
			n, _ := strconv.Atoi(s[:i])
			return n
		}
	}
	n, _ := strconv.Atoi(s)
	return n
}

// Download fetches the binary at url, calling progressFn(downloaded, total) on each chunk.
// total is -1 when Content-Length is unknown.
func Download(url string, progressFn func(downloaded, total int64)) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Minute}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}

	total := resp.ContentLength // -1 if unknown

	var buf []byte
	chunk := make([]byte, 32*1024)
	var downloaded int64
	for {
		n, err := resp.Body.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
			downloaded += int64(n)
			if progressFn != nil {
				progressFn(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

// VerifyChecksum checks the SHA-256 of data against the checksums file at checksumURL.
// Returns (match bool, error). If the asset is not listed in the file, returns (false, nil).
func VerifyChecksum(data []byte, checksumURL, assetName string) (bool, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(checksumURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("checksums fetch returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// Parse lines: "sha256hash  filename"
	var expectedHash string
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		// The filename field may have a leading "*" (binary mode marker) — strip it.
		name := strings.TrimPrefix(fields[1], "*")
		if name == assetName {
			expectedHash = strings.ToLower(fields[0])
			break
		}
	}

	if expectedHash == "" {
		return false, nil // not found in checksums file
	}

	sum := sha256.Sum256(data)
	actual := hex.EncodeToString(sum[:])
	return actual == expectedHash, nil
}

// Apply atomically replaces the running binary with data.
// On Windows it writes alongside as <exe>.new and returns nil;
// the caller should inform the user to restart with the new binary.
func Apply(data []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		newPath := exe + ".new"
		if err := os.WriteFile(newPath, data, 0755); err != nil {
			return err
		}
		return nil
	}

	// Linux/macOS: write to temp in the same dir, then rename atomically.
	dir := filepath.Dir(exe)
	tmp, err := os.CreateTemp(dir, ".jiratui-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { os.Remove(tmpPath) }() // cleanup on failure

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}
	return os.Rename(tmpPath, exe)
}
