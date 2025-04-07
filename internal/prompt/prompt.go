package prompt

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

// Functions contains various function definitions.
type Functions struct {
	Check string `yaml:"check"`
}

// promptConfig represents the entire YAML configuration.
type promptConfig struct {
	Role      Role      `yaml:"role"`
	Functions Functions `yaml:"functions"`
}

// Default YAML file path.
const yamlFilePath = "internal/prompt/prompts.yaml"

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

// GetFunctionDefinition retrieves the definition of a specific function.
func GetFunctionDefinition(functionName string) (string, error) {
	config, err := readPrompt()
	if err != nil {
		return "", err
	}

	switch functionName {
	case "check":
		return config.Functions.Check, nil
	default:
		return "", fmt.Errorf("function %s not found in configuration", functionName)
	}
}
