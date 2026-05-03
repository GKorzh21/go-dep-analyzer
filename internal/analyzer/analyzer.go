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

const (
	goProxyURL = "https://proxy.golang.org"
)

// Analyze колнирует указанный Git-репозиторий, читает его go.mod
// и выводит список прямых зависимостей, доступных для обновления
func Analyze(repoURL string) error {
	tmpDir, err := cloneRepo(repoURL)
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	modFile, err := parseGoMod(tmpDir)
	if err != nil {
		return err
	}

	printModuleInfo(modFile)
	return printAvailableUpdates(modFile)
}

// cloneRepo выполняет shallow clone репозитория во временную папку
// и возвращает путь к ней, скачивает только последний комит
func cloneRepo(repoURL string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "go-dep-analyzer-*")
	if err != nil {
		return "", fmt.Errorf("не удалось создать временную папку: %w", err)
	}

	fmt.Printf("Клонирование %s ...\n\n", repoURL)

	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:      repoURL,
		Depth:    1,
		Progress: nil,
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("не удалось клонировать репозиторий: %w", err)
	}

	return tmpDir, nil
}

// parseGoMod читает и парсит go.mod из указанной директории
func parseGoMod(dir string) (*modfile.File, error) {
	goModPath := filepath.Join(dir, "go.mod")

	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, fmt.Errorf("go.mod не найден в репозитории: %w", err)
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить go.mod: %w", err)
	}

	return modFile, nil
}

// printModuleInfo выводит базовую информацию о Go-модуле
func printModuleInfo(modFile *modfile.File) {
	fmt.Printf("Модуль:     %s\n", modFile.Module.Mod.Path)
	fmt.Printf("Версия Go:  %s\n\n", modFile.Go.Version)
}

// printAvailableUpdates проверяет каждую прямую зависимость через Go-прокси
// и выводит те, для которых доступна новая версия
func printAvailableUpdates(modFile *modfile.File) error {
	fmt.Println("Зависимости, доступные для обновления:")
	fmt.Println(strings.Repeat("-", 80))

	hasUpdates := false

	for _, req := range modFile.Require {
		if req.Indirect {
			continue
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
		fmt.Println("  Все прямые зависимости актуальны.")
	}

	fmt.Println(strings.Repeat("-", 80))
	return nil
}

// getLatestVersion запрашивает последнюю доступную версию модуля
func getLatestVersion(modulePath string) (string, error) {
	url := fmt.Sprintf("%s/%s/@latest", goProxyURL, modulePath)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("запрос к Go-прокси завершился ошибкой: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Go-прокси вернул статус %d для модуля %s", resp.StatusCode, modulePath)
	}

	var result struct {
		Version string `json:"Version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("не удалось декодировать ответ Go-прокси: %w", err)
	}

	return result.Version, nil
}
