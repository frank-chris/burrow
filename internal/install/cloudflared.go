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

	version, err := latestVersion()
	if err != nil {
		return fmt.Errorf("could not determine latest cloudflared version: %w", err)
	}

	binaryName := platformBinaryName()
	downloadURL := fmt.Sprintf("%s/%s/%s", constants.CloudflaredGitHubBase, version, binaryName)
	checksumURL := downloadURL + ".sha256"

	fmt.Printf("  Downloading cloudflared %s...\n", version)

	data, err := downloadFile(downloadURL)
	if err != nil {
		return fmt.Errorf("could not download cloudflared: %w", err)
	}

	if err := verifyChecksum(data, checksumURL); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
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

func latestVersion() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, constants.CloudflaredReleasesURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "burrow")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach GitHub API: %w", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("could not parse GitHub release response")
	}
	if release.TagName == "" {
		return "", fmt.Errorf("empty version in GitHub release response")
	}
	return release.TagName, nil
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

func verifyChecksum(data []byte, checksumURL string) error {
	checksumData, err := downloadFile(checksumURL)
	if err != nil {
		return fmt.Errorf("could not download checksum file: %w", err)
	}

	// Checksum files are formatted as "hash  filename" or just "hash"
	expected := strings.Fields(string(checksumData))
	if len(expected) == 0 {
		return fmt.Errorf("checksum file is empty")
	}

	actual := sha256.Sum256(data)
	actualHex := hex.EncodeToString(actual[:])

	if !strings.EqualFold(actualHex, expected[0]) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected[0], actualHex)
	}
	return nil
}
