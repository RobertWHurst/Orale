package orale_test

import (
	"os"
	"strings"
	"testing"

	orale "github.com/RobertWHurst/orale"
)

func TestExpand(t *testing.T) {
	t.Run("should expand environment variables", func(t *testing.T) {
		type TestStruct struct {
			DatabaseURL string `config:"databaseUrl"`
			APIKey      string `config:"apiKey"`
			Mixed       string `config:"mixed"`
		}

		testStruct := TestStruct{}

		// Set environment variables for the test
		t.Setenv("TEST_DB_HOST", "localhost")
		t.Setenv("TEST_DB_PORT", "5432")
		t.Setenv("TEST_API_SECRET", "secret123")

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"databaseUrl": {"postgres://${TEST_DB_HOST}:${TEST_DB_PORT}/mydb"},
			"apiKey":      {"${TEST_API_SECRET}"},
			"mixed":       {"prefix-${TEST_DB_HOST}-suffix"},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.DatabaseURL != "postgres://localhost:5432/mydb" {
			t.Fatalf("expected DatabaseURL to be 'postgres://localhost:5432/mydb', got %s", testStruct.DatabaseURL)
		}
		if testStruct.APIKey != "secret123" {
			t.Fatalf("expected APIKey to be 'secret123', got %s", testStruct.APIKey)
		}
		if testStruct.Mixed != "prefix-localhost-suffix" {
			t.Fatalf("expected Mixed to be 'prefix-localhost-suffix', got %s", testStruct.Mixed)
		}
	})

	t.Run("should expand config variables", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			BaseURL    string `config:"baseUrl"`
			APIURL     string `config:"apiUrl"`
			WebhookURL string `config:"webhookUrl"`
		}

		testStruct := TestStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"baseUrl":    {"https://example.com"},
			"apiUrl":     {"%{baseUrl}/api/v1"},
			"webhookUrl": {"%{baseUrl}/webhooks/incoming"},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.BaseURL != "https://example.com" {
			t.Fatalf("expected BaseURL to be 'https://example.com', got %s", testStruct.BaseURL)
		}
		if testStruct.APIURL != "https://example.com/api/v1" {
			t.Fatalf("expected APIURL to be 'https://example.com/api/v1', got %s", testStruct.APIURL)
		}
		if testStruct.WebhookURL != "https://example.com/webhooks/incoming" {
			t.Fatalf("expected WebhookURL to be 'https://example.com/webhooks/incoming', got %s", testStruct.WebhookURL)
		}
	})

	t.Run("should expand file variables", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Secret string `config:"secret"`
		}

		testStruct := TestStruct{}

		// Create a temporary file with test content
		tmpfile := t.TempDir() + "/secret.txt"
		content := "my-secret-token"
		if err := os.WriteFile(tmpfile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"secret": {"@{" + tmpfile + "}"},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.Secret != content {
			t.Fatalf("expected Secret to be '%s', got %s", content, testStruct.Secret)
		}
	})

	t.Run("should strip trailing newline from file variables", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Secret string `config:"secret"`
		}

		testStruct := TestStruct{}

		// Create a temporary file with trailing newline
		tmpfile := t.TempDir() + "/secret-with-newline.txt"
		content := "my-secret-token\n"
		if err := os.WriteFile(tmpfile, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"secret": {"@{" + tmpfile + "}"},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.Secret != "my-secret-token" {
			t.Fatalf("expected Secret to be 'my-secret-token', got %s", testStruct.Secret)
		}
	})

	t.Run("should handle mixed variable types in single value", func(t *testing.T) {
		type TestStruct struct {
			ConnectionString string `config:"connectionString"`
		}

		testStruct := TestStruct{}

		t.Setenv("TEST_DB_PASSWORD", "pass123")

		tmpfile := t.TempDir() + "/db-cert.pem"
		certContent := "-----BEGIN CERTIFICATE-----"
		if err := os.WriteFile(tmpfile, []byte(certContent), 0644); err != nil {
			t.Fatal(err)
		}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"dbHost":           {"dbserver.example.com"},
			"connectionString": {"host=%{dbHost};password=${TEST_DB_PASSWORD};cert=@{" + tmpfile + "}"},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		expected := "host=dbserver.example.com;password=pass123;cert=-----BEGIN CERTIFICATE-----"
		if testStruct.ConnectionString != expected {
			t.Fatalf("expected ConnectionString to be '%s', got %s", expected, testStruct.ConnectionString)
		}
	})

	t.Run("should handle empty environment variable", func(t *testing.T) {
		type TestStruct struct {
			Value string `config:"value"`
		}

		testStruct := TestStruct{}

		t.Setenv("TEST_EMPTY_VAR", "")

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"value": {"prefix-${TEST_EMPTY_VAR}-suffix"},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.Value != "prefix--suffix" {
			t.Fatalf("expected Value to be 'prefix--suffix', got %s", testStruct.Value)
		}
	})

	t.Run("should return error for circular variable reference", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			A string `config:"a"`
		}

		testStruct := TestStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"a": {"%{b}"},
			"b": {"%{a}"},
		}

		err := conf.Get("", &testStruct)
		if err == nil {
			t.Fatal("expected error for circular reference, got nil")
		}
		if !strings.Contains(err.Error(), "circular") {
			t.Fatalf("expected error message to contain 'circular', got %s", err.Error())
		}
	})

	t.Run("should return error for missing closing brace in environment variable", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Value string `config:"value"`
		}

		testStruct := TestStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"value": {"${UNCLOSED"},
		}

		err := conf.Get("", &testStruct)
		if err == nil {
			t.Fatal("expected error for missing closing brace, got nil")
		}
		if !strings.Contains(err.Error(), "no closing brace") {
			t.Fatalf("expected error message to contain 'no closing brace', got %s", err.Error())
		}
	})

	t.Run("should return error for missing closing brace in config variable", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Value string `config:"value"`
		}

		testStruct := TestStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"value": {"%{unclosed"},
		}

		err := conf.Get("", &testStruct)
		if err == nil {
			t.Fatal("expected error for missing closing brace, got nil")
		}
		if !strings.Contains(err.Error(), "no closing brace") {
			t.Fatalf("expected error message to contain 'no closing brace', got %s", err.Error())
		}
	})

	t.Run("should return error for missing closing brace in file variable", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Value string `config:"value"`
		}

		testStruct := TestStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"value": {"@{unclosed"},
		}

		err := conf.Get("", &testStruct)
		if err == nil {
			t.Fatal("expected error for missing closing brace, got nil")
		}
		if !strings.Contains(err.Error(), "no closing brace") {
			t.Fatalf("expected error message to contain 'no closing brace', got %s", err.Error())
		}
	})

	t.Run("should return error for non-existent file in file variable", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Value string `config:"value"`
		}

		testStruct := TestStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"value": {"@{/nonexistent/file/path.txt}"},
		}

		err := conf.Get("", &testStruct)
		if err == nil {
			t.Fatal("expected error for non-existent file, got nil")
		}
		if !strings.Contains(err.Error(), "could not be read") {
			t.Fatalf("expected error message to contain 'could not be read', got %s", err.Error())
		}
	})

	t.Run("should not expand variables when value is not a string", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Port int `config:"port"`
		}

		testStruct := TestStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"port": {8080},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.Port != 8080 {
			t.Fatalf("expected Port to be 8080, got %d", testStruct.Port)
		}
	})
}
