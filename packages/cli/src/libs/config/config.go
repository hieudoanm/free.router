package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadAPIKey returns the OpenRouter API key from (in priority order):
//  1. OPENROUTER_API_KEY environment variable
//  2. ~/.fr  (supports KEY=value or bare sk-... lines)
//  3. .env in the current working directory
func LoadAPIKey() string {
	// 1. Env var
	if v := os.Getenv("OPENROUTER_API_KEY"); v != "" {
		return v
	}

	// 2. ~/.fr
	if home, err := os.UserHomeDir(); err == nil {
		if key := readKeyFromFile(filepath.Join(home, ".fr")); key != "" {
			return key
		}
	}

	// 3. .env in cwd
	if key := readKeyFromFile(".env"); key != "" {
		return key
	}

	return ""
}

func readKeyFromFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// KEY=value format
		if strings.HasPrefix(line, "OPENROUTER_API_KEY=") {
			return strings.TrimPrefix(line, "OPENROUTER_API_KEY=")
		}
		// Bare sk-or-... key
		if strings.HasPrefix(line, "sk-") {
			return line
		}
	}
	return ""
}
