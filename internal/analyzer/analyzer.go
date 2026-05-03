package analyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

// Analyze clones the repo and analyzes its dependencies
func Analyze(repoURL string) error {
	// Clone to temp dir
	tmpDir, err := os.MkdirTemp("", "go-dep-analyzer-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Cloning %s ...\n\n", repoURL)
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:      repoURL,
		Depth:    1, // shallow clone, faster
		Progress: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Read go.mod
	goModPath := filepath.Join(tmpDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("go.mod not found in repository: %w", err)
	}

	// Parse go.mod
	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return fmt.Errorf("failed to parse go.mod: %w", err)
	}

	// Print module info
	fmt.Printf("Module:     %s\n", modFile.Module.Mod.Path)
	fmt.Printf("Go version: %s\n", modFile.Go.Version)
	fmt.Printf("\n")

	// Check each dependency for updates
	fmt.Println("Dependencies available for update:")
	fmt.Println(strings.Repeat("-", 80))

	hasUpdates := false
	for _, req := range modFile.Require {
		if req.Indirect {
			continue // skip indirect dependencies
		}

		latest, err := getLatestVersion(req.Mod.Path)
		if err != nil {
			continue
		}

		current := req.Mod.Version
		if semver.Compare(latest, current) > 0 {
			hasUpdates = true
			fmt.Printf("  %-60s %s  →  %s\n", req.Mod.Path, current, latest)
		}
	}

	if !hasUpdates {
		fmt.Println("  All direct dependencies are up to date!")
	}

	fmt.Println(strings.Repeat("-", 80))
	return nil
}

// getLatestVersion queries proxy.golang.org for the latest version of a module
func getLatestVersion(modulePath string) (string, error) {
	url := fmt.Sprintf("https://proxy.golang.org/%s/@latest", modulePath)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("proxy returned %d", resp.StatusCode)
	}

	var result struct {
		Version string `json:"Version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Version, nil
}
