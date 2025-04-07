package config

import (
	"log"
	"os"
	"strings"
)

// envFlag is a custom flag type to collect environment variables.
type EnvFlag []string

// String returns the string representation of the environment variable names.
func (e *EnvFlag) String() string {
	return strings.Join(*e, ", ")
}

// Set appends a new environment variable name to the slice.
func (e *EnvFlag) Set(value string) error {
	*e = append(*e, value)
	return nil
}

// LoadEnvVars retrieves specified environment variables.
func LoadExtraHeadersFromEnvVars(envs EnvFlag) map[string]string {
	// Retrieve environment variable values
	headers := make(map[string]string)
	for _, envVar := range envs {
		value := os.Getenv(envVar)
		if value == "" {
			log.Fatalf("Environment variable %s is not set.", envVar)
		}
		headers[strings.ReplaceAll(envVar, "_", "-")] = value
	}

	return headers
}
