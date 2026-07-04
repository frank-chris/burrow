package install

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
	"strings"
	"time"

	"github.com/frank-chris/burrow/internal/constants"
)

const downloadTimeout = 5 * time.Minute

// BinaryPath returns the path to the managed cloudflared binary.
func BinaryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	name := "cloudflared"
	if runtime.GOOS == "windows" {
		name = "cloudflared.exe"
	}
	return filepath.Join(home, constants.ConfigDirName, constants.CloudflaredBinDir, name), nil
}

// IsInstalled reports whether the managed cloudflared binary exists.
func IsInstalled() bool {
	path, err := BinaryPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Install downloads and installs the latest cloudflared binary.
// If already installed, it is a no-op.
func Install() error {
	if IsInstalled() {
		return nil
	}

	release, err := latestRelease()
	if err != nil {
		return fmt.Errorf("could not determine latest cloudflared version: %w", err)
	}

	binaryName := platformBinaryName()
	downloadURL := fmt.Sprintf("%s/%s/%s", constants.CloudflaredGitHubBase, release.version, binaryName)

	fmt.Printf("  Downloading cloudflared %s...\n", release.version)

	data, err := downloadFile(downloadURL)
	if err != nil {
		return fmt.Errorf("could not download cloudflared: %w", err)
	}

	if expected, ok := release.checksums[binaryName]; ok {
		actual := sha256.Sum256(data)
		actualHex := hex.EncodeToString(actual[:])
		if !strings.EqualFold(actualHex, expected) {
			return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", binaryName, expected, actualHex)
		}
	} else {
		fmt.Println("  Warning: could not verify checksum (not found in release notes). Proceeding anyway.")
	}

	binPath, err := BinaryPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(binPath), 0700); err != nil {
		return fmt.Errorf("could not create bin directory: %w", err)
	}

	if err := os.WriteFile(binPath, data, 0700); err != nil {
		return fmt.Errorf("could not write cloudflared binary: %w", err)
	}

	fmt.Println("  cloudflared installed.")
	return nil
}

type releaseInfo struct {
	version   string
	checksums map[string]string // filename -> sha256 hex
}

func latestRelease() (*releaseInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, constants.CloudflaredReleasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "burrow")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach GitHub API: %w", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name   string `json:"name"`
			Digest string `json:"digest"` // "sha256:<hex>"
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("could not parse GitHub release response")
	}
	if release.TagName == "" {
		return nil, fmt.Errorf("empty version in GitHub release response")
	}

	checksums := make(map[string]string)
	for _, asset := range release.Assets {
		hash, ok := strings.CutPrefix(asset.Digest, "sha256:")
		if ok && len(hash) == 64 {
			checksums[asset.Name] = hash
		}
	}

	return &releaseInfo{
		version:   release.TagName,
		checksums: checksums,
	}, nil
}

func platformBinaryName() string {
	name := fmt.Sprintf("cloudflared-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func downloadFile(url string) ([]byte, error) {
	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d downloading from %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
