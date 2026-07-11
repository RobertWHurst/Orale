package orale_test

import (
	"testing"

	"github.com/RobertWHurst/orale"
)

func Test_Explain(t *testing.T) {
	loader := &orale.Loader{
		FlagValues: map[string][]any{
			"port": {"3000"},
		},
		EnvironmentValues: map[string][]any{
			"host": {"localhost"},
		},
		ConfigurationFiles: []*orale.File{
			{
				Path: "config.toml",
				Values: map[string][]any{
					"debug": {true},
				},
			},
		},
	}

	entries := loader.Explain()

	if len(entries) != 3 {
		t.Errorf("Explain() returned %d entries, want 3", len(entries))
	}

	expectedPaths := map[string]bool{
		"port":  false,
		"host":  false,
		"debug": false,
	}

	for _, entry := range entries {
		if _, ok := expectedPaths[entry.Path]; !ok {
			t.Errorf("unexpected path: %s", entry.Path)
		}
		expectedPaths[entry.Path] = true
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("missing expected path: %s", path)
		}
	}
}

func Test_Explain_deduplicatesByPrecedence(t *testing.T) {
	loader := &orale.Loader{
		FlagValues: map[string][]any{
			"port": {"3000"},
		},
		EnvironmentValues: map[string][]any{
			"port": {"8080"},
		},
		ConfigurationFiles: []*orale.File{
			{
				Path: "config.toml",
				Values: map[string][]any{
					"port": {"9999"},
				},
			},
		},
	}

	entries := loader.Explain()

	if len(entries) != 1 {
		t.Errorf("Explain() should deduplicate, got %d entries, want 1", len(entries))
	}

	if entries[0].Path != "port" {
		t.Errorf("entry path = %s, want port", entries[0].Path)
	}
	if entries[0].Source != "flag" {
		t.Errorf("entry source = %s, want flag", entries[0].Source)
	}
	if entries[0].Value != "3000" {
		t.Errorf("entry value = %s, want 3000", entries[0].Value)
	}
}

func Test_Explain_masksSecretValues(t *testing.T) {
	loader := &orale.Loader{
		FlagValues: map[string][]any{
			"apiKey": {"secret-api-key"},
			"host":   {"localhost"},
		},
		EnvironmentValues: map[string][]any{},
		ConfigurationFiles: []*orale.File{
			{
				Path: "config.toml",
				Values: map[string][]any{
					"database.password": {"secret-password"},
				},
			},
		},
		SecretPaths: map[string]bool{
			"apiKey":            true,
			"database.password": true,
		},
	}

	entries := loader.Explain()

	values := map[string]orale.ExplainEntry{}
	for _, entry := range entries {
		values[entry.Path] = entry
	}

	if values["apiKey"].Value != "******" {
		t.Errorf("apiKey value = %q, want ******", values["apiKey"].Value)
	}
	if !values["apiKey"].Secret {
		t.Error("apiKey Secret = false, want true")
	}
	if values["database.password"].Value != "******" {
		t.Errorf("database.password value = %q, want ******", values["database.password"].Value)
	}
	if !values["database.password"].Secret {
		t.Error("database.password Secret = false, want true")
	}
	if values["host"].Value != "localhost" {
		t.Errorf("host value = %q, want localhost", values["host"].Value)
	}
	if values["host"].Secret {
		t.Error("host Secret = true, want false")
	}
}

func Test_Explain_masksSecretDescendants(t *testing.T) {
	loader := &orale.Loader{
		FlagValues: map[string][]any{
			"database.password": {"secret-password"},
			"database.user":     {"admin"},
		},
		EnvironmentValues:  map[string][]any{},
		ConfigurationFiles: []*orale.File{},
		SecretPaths: map[string]bool{
			"database": true,
		},
	}

	entries := loader.Explain()

	for _, entry := range entries {
		if entry.Value != "******" {
			t.Errorf("%s value = %q, want ******", entry.Path, entry.Value)
		}
		if !entry.Secret {
			t.Errorf("%s Secret = false, want true", entry.Path)
		}
	}
}

func Test_Explain_sortsByPath(t *testing.T) {
	loader := &orale.Loader{
		FlagValues: map[string][]any{
			"zulu":  {"z"},
			"alpha": {"a"},
			"mike":  {"m"},
		},
		EnvironmentValues:  map[string][]any{},
		ConfigurationFiles: []*orale.File{},
	}

	entries := loader.Explain()

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Path != "alpha" {
		t.Errorf("first entry path = %s, want alpha", entries[0].Path)
	}
	if entries[1].Path != "mike" {
		t.Errorf("second entry path = %s, want mike", entries[1].Path)
	}
	if entries[2].Path != "zulu" {
		t.Errorf("third entry path = %s, want zulu", entries[2].Path)
	}
}

func Test_Explain_formatsValues(t *testing.T) {
	loader := &orale.Loader{
		FlagValues: map[string][]any{
			"empty":    {},
			"multiple": {"a", "b", "c"},
			"single":   {"test"},
		},
		EnvironmentValues:  map[string][]any{},
		ConfigurationFiles: []*orale.File{},
	}

	entries := loader.Explain()

	values := map[string]string{}
	for _, entry := range entries {
		values[entry.Path] = entry.Value
	}

	if values["single"] != "test" {
		t.Errorf("single value = %q, want test", values["single"])
	}
	if values["multiple"] != "[a, b, c]" {
		t.Errorf("multiple value = %q, want [a, b, c]", values["multiple"])
	}
	if values["empty"] != "" {
		t.Errorf("empty value = %q, want empty string", values["empty"])
	}
}
