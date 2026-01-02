package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RobertWHurst/orale"
)

func Test_runExplain_noConfig(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change dir: %v", err)
	}

	err := runExplain("testApp", []string{})
	if err != nil {
		t.Fatalf("runExplain() error = %v", err)
	}
}

func Test_runExplain_fromConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change dir: %v", err)
	}

	configContent := `
port = 8080
host = "localhost"

[database]
connection_string = "postgres://localhost/db"
`

	configPath := filepath.Join(tempDir, "test-app.config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	err := runExplain("testApp", []string{})
	if err != nil {
		t.Fatalf("runExplain() error = %v", err)
	}
}

func Test_runExplain_fromFlags(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change dir: %v", err)
	}

	args := []string{"--port=3000", "--debug=true"}

	err := runExplain("testApp", args)
	if err != nil {
		t.Fatalf("runExplain() error = %v", err)
	}
}

func Test_runExplain_fromEnvironment(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change dir: %v", err)
	}

	originalEnv := os.Environ()
	os.Clearenv()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			parts := splitEnvVar(env)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	os.Setenv("TEST_APP__PORT", "5000")
	os.Setenv("TEST_APP__DEBUG", "true")

	err := runExplain("testApp", []string{})
	if err != nil {
		t.Fatalf("runExplain() error = %v", err)
	}
}

func Test_runExplain_precedence(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change dir: %v", err)
	}

	configContent := `
port = 8080
host = "config-host"
`

	configPath := filepath.Join(tempDir, "test-app.config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalEnv := os.Environ()
	os.Clearenv()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			parts := splitEnvVar(env)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}()

	os.Setenv("TEST_APP__HOST", "env-host")

	args := []string{"--port=9999"}

	err := runExplain("testApp", args)
	if err != nil {
		t.Fatalf("runExplain() error = %v", err)
	}
}

func Test_collectConfigEntries(t *testing.T) {
	loader := &orale.Loader{
		FlagValues: map[string][]any{
			"port": {"3000"},
		},
		EnvironmentValues: map[string][]any{
			"host": {"localhost"},
		},
		ConfigurationFiles: []*orale.File{
			{
				Path: "/path/to/config.toml",
				Values: map[string][]any{
					"debug": {true},
				},
			},
		},
	}

	entries := collectConfigEntries(loader)

	if len(entries) != 3 {
		t.Errorf("collectConfigEntries() returned %d entries, want 3", len(entries))
	}

	expectedPaths := map[string]bool{
		"port":  false,
		"host":  false,
		"debug": false,
	}

	for _, entry := range entries {
		if _, ok := expectedPaths[entry.path]; !ok {
			t.Errorf("unexpected path: %s", entry.path)
		}
		expectedPaths[entry.path] = true
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("missing expected path: %s", path)
		}
	}
}

func Test_collectConfigEntries_deduplicates(t *testing.T) {
	loader := &orale.Loader{
		FlagValues: map[string][]any{
			"port": {"3000"},
		},
		EnvironmentValues: map[string][]any{
			"port": {"8080"},
		},
		ConfigurationFiles: []*orale.File{
			{
				Path: "/path/to/config.toml",
				Values: map[string][]any{
					"port": {"9999"},
				},
			},
		},
	}

	entries := collectConfigEntries(loader)

	if len(entries) != 1 {
		t.Errorf("collectConfigEntries() should deduplicate, got %d entries, want 1", len(entries))
	}

	if entries[0].path != "port" {
		t.Errorf("entry path = %s, want port", entries[0].path)
	}

	if entries[0].source != "flag" {
		t.Errorf("entry source = %s, want flag", entries[0].source)
	}

	if entries[0].value != "3000" {
		t.Errorf("entry value = %s, want 3000", entries[0].value)
	}
}

func Test_collectConfigEntries_sortsByPath(t *testing.T) {
	loader := &orale.Loader{
		FlagValues: map[string][]any{
			"zulu":  {"z"},
			"alpha": {"a"},
			"mike":  {"m"},
		},
		EnvironmentValues:  map[string][]any{},
		ConfigurationFiles: []*orale.File{},
	}

	entries := collectConfigEntries(loader)

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	if entries[0].path != "alpha" {
		t.Errorf("first entry path = %s, want alpha", entries[0].path)
	}

	if entries[1].path != "mike" {
		t.Errorf("second entry path = %s, want mike", entries[1].path)
	}

	if entries[2].path != "zulu" {
		t.Errorf("third entry path = %s, want zulu", entries[2].path)
	}
}

func Test_formatValue_single(t *testing.T) {
	values := []any{"test"}
	result := formatValue(values)

	if result != "test" {
		t.Errorf("formatValue() = %q, want %q", result, "test")
	}
}

func Test_formatValue_multiple(t *testing.T) {
	values := []any{"a", "b", "c"}
	result := formatValue(values)

	expected := "[a, b, c]"
	if result != expected {
		t.Errorf("formatValue() = %q, want %q", result, expected)
	}
}

func Test_formatValue_empty(t *testing.T) {
	values := []any{}
	result := formatValue(values)

	if result != "" {
		t.Errorf("formatValue() = %q, want empty string", result)
	}
}

func Test_wrapText_noWrap(t *testing.T) {
	text := "short"
	width := 20

	lines := wrapText(text, width)

	if len(lines) != 1 {
		t.Errorf("wrapText() returned %d lines, want 1", len(lines))
	}

	if lines[0] != text {
		t.Errorf("wrapText() = %q, want %q", lines[0], text)
	}
}

func Test_wrapText_wrapWords(t *testing.T) {
	text := "this is a long text that should wrap"
	width := 15

	lines := wrapText(text, width)

	if len(lines) < 2 {
		t.Errorf("wrapText() returned %d lines, want at least 2", len(lines))
	}

	for i, line := range lines {
		if len(line) > width {
			t.Errorf("line %d exceeds width: got %d, want <= %d", i, len(line), width)
		}
	}
}

func Test_wrapText_longWord(t *testing.T) {
	text := "verylongwordthatexceedswidth"
	width := 10

	lines := wrapText(text, width)

	if len(lines) < 2 {
		t.Errorf("wrapText() should split long words, got %d lines", len(lines))
	}

	for i, line := range lines {
		if len(line) > width {
			t.Errorf("line %d exceeds width: got %d, want <= %d", i, len(line), width)
		}
	}
}

func Test_convertAppNameToEnvPrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"myApp", "MY_APP"},
		{"testApp", "TEST_APP"},
		{"my-app", "MYAPP"},
		{"MyApp", "MY_APP"},
		{"api", "API"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertAppNameToEnvPrefix(tt.input)
			if result != tt.want {
				t.Errorf("convertAppNameToEnvPrefix(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

func Test_convertAppNameToConfigName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"myApp", "my-app"},
		{"testApp", "test-app"},
		{"my_app", "myapp"},
		{"MyApp", "my-app"},
		{"api", "api"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertAppNameToConfigName(tt.input)
			if result != tt.want {
				t.Errorf("convertAppNameToConfigName(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

func splitEnvVar(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env}
}
