package utils

import (
	"fmt"
	"os"
	"strconv"
)

// GetStringOrDefault retrieves the value of the environment variable named by the key.
// If the environment variable is not set, it returns the provided default value.
func GetStringOrDefault(envName string, defaultValue string) string {
	if value, exists := os.LookupEnv(envName); exists {
		return value
	}
	return defaultValue
}

// GetRequiredString retrieves the value of the environment variable named by the key.
// If the environment variable is not set, it returns an error.
func GetRequiredString(envName string) (string, error) {
	if value, exists := os.LookupEnv(envName); exists {
		return value, nil
	}
	return "", fmt.Errorf("missing required env: %s", envName)
}

// GetIntOrDefault retrieves the value of the environment variable named by the key as an integer.
// If the environment variable is not set or cannot be parsed as an integer, it returns the provided default value.
func GetIntOrDefault(envName string, defaultValue int) int {
	if value, exists := os.LookupEnv(envName); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
