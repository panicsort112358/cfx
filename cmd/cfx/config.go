package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadWorkspaceDir() (string, error) {
	contents, err := os.ReadFile(".env")
	if err != nil {
		return "", fmt.Errorf("read .env: %w", err)
	}

	var workspaceDir string
	for i, rawLine := range strings.Split(string(contents), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return "", fmt.Errorf(".env:%d: expected KEY=VALUE", i+1)
		}
		if strings.TrimSpace(key) != "CFX_DIR" {
			continue
		}
		if workspaceDir != "" {
			return "", fmt.Errorf(".env: CFX_DIR is defined more than once")
		}

		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '\'' && value[len(value)-1] == '\'' || value[0] == '"' && value[len(value)-1] == '"') {
			value = value[1 : len(value)-1]
		} else if comment := strings.Index(value, " #"); comment >= 0 {
			value = strings.TrimSpace(value[:comment])
		}
		if value == "" {
			return "", fmt.Errorf(".env:%d: CFX_DIR must not be empty", i+1)
		}
		workspaceDir = value
	}

	if workspaceDir == "" {
		return "", fmt.Errorf(".env must define CFX_DIR")
	}

	if !filepath.IsAbs(workspaceDir) {
		workspaceDir = filepath.Join(".", workspaceDir)
	}
	return filepath.Clean(workspaceDir), nil
}
