package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetFiles retrieves all files from the given path recursively.
func GetFiles(rootPath string) ([]string, error) {
	if _, err := os.Stat(rootPath); err != nil {
		return nil, fmt.Errorf("failed to access root path: %v", err)
	}

	var files []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasPrefix(path, ".git") && len(files) < 1 {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %v", err)
	}

	return files, nil
}

// ReadFile reads the content of the file at the specified path.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", path, err)
	}
	return string(data), nil
}
