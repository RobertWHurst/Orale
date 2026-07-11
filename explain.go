package orale

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const secretValueMask = "******"

// ExplainEntry describes one loaded configuration value and where it came from.
type ExplainEntry struct {
	Path   string
	Value  string
	Source string
	Secret bool
}

// Explain returns loaded configuration values in precedence order with one
// entry per path. Values from higher-precedence sources hide lower-precedence
// values. Values tagged `secret:"true"` are masked.
func (l *Loader) Explain() []ExplainEntry {
	seen := make(map[string]bool)
	var entries []ExplainEntry

	for path, values := range l.FlagValues {
		if seen[path] {
			continue
		}
		seen[path] = true

		entries = append(entries, l.explainEntry(path, values, "flag"))
	}

	for path, values := range l.EnvironmentValues {
		if seen[path] {
			continue
		}
		seen[path] = true

		entries = append(entries, l.explainEntry(path, values, "environment"))
	}

	for _, file := range l.ConfigurationFiles {
		for path, values := range file.Values {
			if seen[path] {
				continue
			}
			seen[path] = true

			source := file.Path
			if workingDir, err := os.Getwd(); err == nil {
				if rel, err := filepath.Rel(workingDir, file.Path); err == nil {
					source = rel
				}
			}

			entries = append(entries, l.explainEntry(path, values, source))
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	return entries
}

func (l *Loader) explainEntry(path string, values []any, source string) ExplainEntry {
	secret := l.isSecretPath(path)
	value := formatExplainValue(values)
	if secret {
		value = secretValueMask
	}

	return ExplainEntry{
		Path:   path,
		Value:  value,
		Source: source,
		Secret: secret,
	}
}

func (l *Loader) isSecretPath(path string) bool {
	if l == nil {
		return false
	}

	for secretPath := range l.SecretPaths {
		if path == secretPath || strings.HasPrefix(path, secretPath+".") {
			return true
		}
	}
	return false
}

func formatExplainValue(values []any) string {
	if len(values) == 0 {
		return ""
	}

	if len(values) == 1 {
		return fmt.Sprintf("%v", values[0])
	}

	strValues := make([]string, len(values))
	for i, v := range values {
		strValues[i] = fmt.Sprintf("%v", v)
	}
	return "[" + strings.Join(strValues, ", ") + "]"
}
