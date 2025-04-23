package prompts

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Role contains different roles with their respective prompts.
type Role struct {
	Security string `yaml:"security"`
	Review   string `yaml:"review"`
}

// promptConfig represents the entire YAML configuration.
type promptConfig struct {
	Role Role `yaml:"role"`
}

// Default YAML file path.
const yamlFilePath = "internal/prompts/prompts.yaml"

// readPrompt reads the prompt configuration from the YAML file and returns a promptConfig struct.
func readPrompt() (*promptConfig, error) {
	data, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file: %w", err)
	}

	var config promptConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML data: %w", err)
	}

	return &config, nil
}

// GetRolePrompt retrieves the prompt for a specific role.
func GetSystemRole(role string) (string, error) {
	config, err := readPrompt()
	if err != nil {
		return "", err
	}

	switch role {
	case "security":
		return config.Role.Security, nil
	case "review":
		return config.Role.Review, nil
	default:
		return "", fmt.Errorf("role %s not found in configuration", role)
	}
}
