package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func BootstrapEnv() {
	loadDotEnvFromWorkingDir()
}

func loadDotEnvFromWorkingDir() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if trimmed, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(trimmed)
		}

		key, rawValue, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		if existing, exists := os.LookupEnv(key); exists && strings.TrimSpace(existing) != "" {
			continue
		}

		value := parseDotEnvValue(rawValue)
		_ = os.Setenv(key, value)
	}
}

func parseDotEnvValue(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}

	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		if parsed, err := strconv.Unquote(value); err == nil {
			return parsed
		}

		return strings.Trim(value, `"`)
	}

	if strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`) {
		return strings.Trim(value, `'`)
	}

	// For unquoted values, support trailing inline comments: KEY=value # comment
	if idx := strings.Index(value, " #"); idx >= 0 {
		value = value[:idx]
	}

	return strings.TrimSpace(value)
}
