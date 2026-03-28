package run

import (
	"fmt"
	"regexp"
	"strings"
)

var envPattern = regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

func ReplaceEnvs(data []byte, envs map[string]string) []byte {
	if envs == nil {
		return data
	}

	result := string(data)

	for key, value := range envs {
		placeholder := fmt.Sprintf("${%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return []byte(result)
}

func FindMissingEnvs(data []byte, envs map[string]string) []string {
	if envs == nil {
		return nil
	}

	matches := envPattern.FindAllStringSubmatch(string(data), -1)
	missing := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			key := match[1]
			if _, exists := envs[key]; !exists {
				missing[key] = true
			}
		}
	}

	result := make([]string, 0, len(missing))
	for key := range missing {
		result = append(result, key)
	}

	return result
}
