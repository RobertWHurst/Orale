package orale_test

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	orale "github.com/RobertWHurst/orale"
)

func newTestLoaderSingleValues() *orale.Loader {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"a": {"1"},
		"b": {"2"},
		"c": {"3"},
		"d": {"4"},
	}
	l.EnvironmentValues = map[string][]any{
		"b": {"5"},
		"e": {"6"},
	}
	l.ConfigurationFiles = []*orale.File{
		{
			Path: "path/to/other/file-2.toml",
			Values: map[string][]any{
				"d": {"9"},
				"g": {"10"},
			},
		},
		{
			Path: "path/to/file-1.toml",
			Values: map[string][]any{
				"c": {"7"},
				"f": {"8"},
				"g": {"9"},
				"h": {"10"},
			},
		},
	}
	return l
}

func newTestLoaderMultiValues() *orale.Loader {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"a": {"1", "2"},
		"b": {"3", "4"},
		"c": {"5", "6"},
		"d": {"7", "8"},
	}
	l.EnvironmentValues = map[string][]any{
		"b": {"9", "10"},
		"e": {"11", "12"},
	}
	l.ConfigurationFiles = []*orale.File{}
	return l
}

func Test_Get(t *testing.T) {
	t.Run("should handle type conversions from environment variables", func(t *testing.T) {
		t.Parallel()

		type TestConversionStruct struct {
			// Int conversions
			StringToInt    int `config:"stringToInt"`
			FloatToInt     int `config:"floatToInt"`
			BoolTrueToInt  int `config:"boolTrueToInt"`
			BoolFalseToInt int `config:"boolFalseToInt"`

			// Uint conversions
			StringToUint    uint `config:"stringToUint"`
			FloatToUint     uint `config:"floatToUint"`
			BoolTrueToUint  uint `config:"boolTrueToUint"`
			BoolFalseToUint uint `config:"boolFalseToUint"`

			// Float conversions
			StringToFloat    float64 `config:"stringToFloat"`
			IntToFloat       float64 `config:"intToFloat"`
			BoolTrueToFloat  float64 `config:"boolTrueToFloat"`
			BoolFalseToFloat float64 `config:"boolFalseToFloat"`

			// Bool conversions
			StringTrueToBool  bool `config:"stringTrueToBool"`
			StringYesToBool   bool `config:"stringYesToBool"`
			StringOneToBool   bool `config:"stringOneToBool"`
			StringFalseToBool bool `config:"stringFalseToBool"`
			StringNoToBool    bool `config:"stringNoToBool"`
			StringZeroToBool  bool `config:"stringZeroToBool"`
			IntOneToBool      bool `config:"intOneToBool"`
			IntZeroToBool     bool `config:"intZeroToBool"`

			// String conversions
			IntToString       string `config:"intToString"`
			FloatToString     string `config:"floatToString"`
			BoolTrueToString  string `config:"boolTrueToString"`
			BoolFalseToString string `config:"boolFalseToString"`
		}

		testStruct := TestConversionStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			// Int conversions
			"stringToInt":    {"42"},
			"floatToInt":     {42.7},
			"boolTrueToInt":  {true},
			"boolFalseToInt": {false},

			// Uint conversions
			"stringToUint":    {"84"},
			"floatToUint":     {84.7},
			"boolTrueToUint":  {true},
			"boolFalseToUint": {false},

			// Float conversions
			"stringToFloat":    {"3.14"},
			"intToFloat":       {42},
			"boolTrueToFloat":  {true},
			"boolFalseToFloat": {false},

			// Bool conversions
			"stringTrueToBool":  {"true"},
			"stringYesToBool":   {"yes"},
			"stringOneToBool":   {"1"},
			"stringFalseToBool": {"false"},
			"stringNoToBool":    {"no"},
			"stringZeroToBool":  {"0"},
			"intOneToBool":      {1},
			"intZeroToBool":     {0},

			// String conversions
			"intToString":       {42},
			"floatToString":     {3.14},
			"boolTrueToString":  {true},
			"boolFalseToString": {false},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		// Int conversions
		if testStruct.StringToInt != 42 {
			t.Fatalf("expected StringToInt to be 42, got %d", testStruct.StringToInt)
		}
		if testStruct.FloatToInt != 42 {
			t.Fatalf("expected FloatToInt to be 42, got %d", testStruct.FloatToInt)
		}
		if testStruct.BoolTrueToInt != 1 {
			t.Fatalf("expected BoolTrueToInt to be 1, got %d", testStruct.BoolTrueToInt)
		}
		if testStruct.BoolFalseToInt != 0 {
			t.Fatalf("expected BoolFalseToInt to be 0, got %d", testStruct.BoolFalseToInt)
		}

		// Uint conversions
		if testStruct.StringToUint != 84 {
			t.Fatalf("expected StringToUint to be 84, got %d", testStruct.StringToUint)
		}
		if testStruct.FloatToUint != 84 {
			t.Fatalf("expected FloatToUint to be 84, got %d", testStruct.FloatToUint)
		}
		if testStruct.BoolTrueToUint != 1 {
			t.Fatalf("expected BoolTrueToUint to be 1, got %d", testStruct.BoolTrueToUint)
		}
		if testStruct.BoolFalseToUint != 0 {
			t.Fatalf("expected BoolFalseToUint to be 0, got %d", testStruct.BoolFalseToUint)
		}

		// Float conversions
		if testStruct.StringToFloat != 3.14 {
			t.Fatalf("expected StringToFloat to be 3.14, got %f", testStruct.StringToFloat)
		}
		if testStruct.IntToFloat != 42.0 {
			t.Fatalf("expected IntToFloat to be 42.0, got %f", testStruct.IntToFloat)
		}
		if testStruct.BoolTrueToFloat != 1.0 {
			t.Fatalf("expected BoolTrueToFloat to be 1.0, got %f", testStruct.BoolTrueToFloat)
		}
		if testStruct.BoolFalseToFloat != 0.0 {
			t.Fatalf("expected BoolFalseToFloat to be 0.0, got %f", testStruct.BoolFalseToFloat)
		}

		// Bool conversions
		if testStruct.StringTrueToBool != true {
			t.Fatalf("expected StringTrueToBool to be true, got %v", testStruct.StringTrueToBool)
		}
		if testStruct.StringYesToBool != true {
			t.Fatalf("expected StringYesToBool to be true, got %v", testStruct.StringYesToBool)
		}
		if testStruct.StringOneToBool != true {
			t.Fatalf("expected StringOneToBool to be true, got %v", testStruct.StringOneToBool)
		}
		if testStruct.StringFalseToBool != false {
			t.Fatalf("expected StringFalseToBool to be false, got %v", testStruct.StringFalseToBool)
		}
		if testStruct.StringNoToBool != false {
			t.Fatalf("expected StringNoToBool to be false, got %v", testStruct.StringNoToBool)
		}
		if testStruct.StringZeroToBool != false {
			t.Fatalf("expected StringZeroToBool to be false, got %v", testStruct.StringZeroToBool)
		}
		if testStruct.IntOneToBool != true {
			t.Fatalf("expected IntOneToBool to be true, got %v", testStruct.IntOneToBool)
		}
		if testStruct.IntZeroToBool != false {
			t.Fatalf("expected IntZeroToBool to be false, got %v", testStruct.IntZeroToBool)
		}

		// String conversions
		if testStruct.IntToString != "42" {
			t.Fatalf("expected IntToString to be '42', got %s", testStruct.IntToString)
		}
		if testStruct.FloatToString != "3.14" {
			t.Fatalf("expected FloatToString to be '3.14', got %s", testStruct.FloatToString)
		}
		if testStruct.BoolTrueToString != "true" {
			t.Fatalf("expected BoolTrueToString to be 'true', got %s", testStruct.BoolTrueToString)
		}
		if testStruct.BoolFalseToString != "false" {
			t.Fatalf("expected BoolFalseToString to be 'false', got %s", testStruct.BoolFalseToString)
		}
	})

	t.Run("should correctly resolve values into struct", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			A string `config:"a"`
			B string `config:"b"`
			C string `config:"c"`
			D string `config:"d"`
			F string `config:"f"`
			G string `config:"g"`
			H string `config:"h"`
			I string `config:"i"`
		}

		testStruct := TestStruct{}

		conf := newTestLoaderSingleValues()
		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.A != "1" {
			t.Fatalf("expected A to be 1, got %s", testStruct.A)
		}
		if testStruct.B != "2" {
			t.Fatalf("expected B to be 2, got %s", testStruct.B)
		}
		if testStruct.C != "3" {
			t.Fatalf("expected C to be 3, got %s", testStruct.C)
		}
		if testStruct.D != "4" {
			t.Fatalf("expected D to be 4, got %s", testStruct.D)
		}
		if testStruct.F != "8" {
			t.Fatalf("expected F to be 8, got %s", testStruct.F)
		}
		if testStruct.G != "10" {
			t.Fatalf("expected G to be 10, got %s", testStruct.G)
		}
		if testStruct.H != "10" {
			t.Fatalf("expected H to be 10, got %s", testStruct.H)
		}
		if testStruct.I != "" {
			t.Fatalf("expected I to be empty, got %s", testStruct.I)
		}
	})

	t.Run("should correctly resolve multi values into struct", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			A []string `config:"a"`
			B []string `config:"b"`
			C []string `config:"c"`
			D []string `config:"d"`
			E []string `config:"e"`
		}

		testStruct := TestStruct{}

		conf := newTestLoaderMultiValues()
		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if len(testStruct.A) != 2 {
			t.Fatalf("expected A to have 2 values, got %d", len(testStruct.A))
		}
		if testStruct.A[0] != "1" {
			t.Fatalf("expected A[0] to be 1, got %s", testStruct.A[0])
		}
		if testStruct.A[1] != "2" {
			t.Fatalf("expected A[1] to be 2, got %s", testStruct.A[1])
		}
		if len(testStruct.B) != 2 {
			t.Fatalf("expected B to have 2 values, got %d", len(testStruct.B))
		}
		if testStruct.B[0] != "3" {
			t.Fatalf("expected B[0] to be 3, got %s", testStruct.B[0])
		}
		if testStruct.B[1] != "4" {
			t.Fatalf("expected B[1] to be 4, got %s", testStruct.B[1])
		}
		if len(testStruct.C) != 2 {
			t.Fatalf("expected C to have 2 values, got %d", len(testStruct.C))
		}
		if testStruct.C[0] != "5" {
			t.Fatalf("expected C[0] to be 5, got %s", testStruct.C[0])
		}
		if testStruct.C[1] != "6" {
			t.Fatalf("expected C[1] to be 6, got %s", testStruct.C[1])
		}
		if len(testStruct.D) != 2 {
			t.Fatalf("expected D to have 2 values, got %d", len(testStruct.D))
		}
		if testStruct.D[0] != "7" {
			t.Fatalf("expected D[0] to be 7, got %s", testStruct.D[0])
		}
		if testStruct.D[1] != "8" {
			t.Fatalf("expected D[1] to be 8, got %s", testStruct.D[1])
		}
		if len(testStruct.E) != 2 {
			t.Fatalf("expected E to have 2 values, got %d", len(testStruct.E))
		}
		if testStruct.E[0] != "11" {
			t.Fatalf("expected E[0] to be 11, got %s", testStruct.E[0])
		}
		if testStruct.E[1] != "12" {
			t.Fatalf("expected E[1] to be 12, got %s", testStruct.E[1])
		}
	})

	t.Run("should leave default values when no replacement values are loaded", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			A string `config:"a"`
			B string `config:"b"`
			C string `config:"c"`
			D string `config:"d"`
		}

		testStruct := TestStruct{
			A: "2",
			B: "3",
			C: "4",
			D: "5",
		}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.FlagValues = map[string][]any{
			"a": {"1"},
			"b": {"2"},
			"d": {"4"},
		}
		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.A != "1" {
			t.Fatalf("expected A to be 1, got %s", testStruct.A)
		}
		if testStruct.B != "2" {
			t.Fatalf("expected B to be 2, got %s", testStruct.B)
		}
		if testStruct.C != "4" {
			t.Fatalf("expected C to be 4, got %s", testStruct.C)
		}
		if testStruct.D != "4" {
			t.Fatalf("expected D to be 4, got %s", testStruct.D)
		}
	})

	t.Run("should correctly resolve values when using embedded struct", func(t *testing.T) {
		type EmbeddedStruct struct {
			C string `config:"c"`
			D string `config:"d"`
		}

		type TestStruct struct {
			A string `config:"a"`
			B string `config:"b"`
			EmbeddedStruct
		}

		testStruct := TestStruct{}

		conf := newTestLoaderSingleValues()
		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.A != "1" {
			t.Fatalf("expected A to be 1, got %s", testStruct.A)
		}
		if testStruct.B != "2" {
			t.Fatalf("expected B to be 2, got %s", testStruct.B)
		}
		if testStruct.C != "3" {
			t.Fatalf("expected C to be 3, got %s", testStruct.C)
		}
		if testStruct.D != "4" {
			t.Fatalf("expected D to be 4, got %s", testStruct.D)
		}
	})

	t.Run("should handle time.Duration conversions", func(t *testing.T) {
		t.Parallel()

		type TestDurationStruct struct {
			StringDuration5h    time.Duration `config:"stringDuration5h"`
			StringDuration30m   time.Duration `config:"stringDuration30m"`
			StringDuration1h30m time.Duration `config:"stringDuration1h30m"`
			StringDuration500ms time.Duration `config:"stringDuration500ms"`
			IntNanoseconds      time.Duration `config:"intNanoseconds"`
			Int64Nanoseconds    time.Duration `config:"int64Nanoseconds"`
			Uint64Nanoseconds   time.Duration `config:"uint64Nanoseconds"`
			FloatNanoseconds    time.Duration `config:"floatNanoseconds"`
		}

		testStruct := TestDurationStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"stringDuration5h":    {"5h"},
			"stringDuration30m":   {"30m"},
			"stringDuration1h30m": {"1h30m"},
			"stringDuration500ms": {"500ms"},
			"intNanoseconds":      {int(1000000000)},
			"int64Nanoseconds":    {int64(2000000000)},
			"uint64Nanoseconds":   {uint64(3000000000)},
			"floatNanoseconds":    {4000000000.0},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		if testStruct.StringDuration5h != 5*time.Hour {
			t.Fatalf("expected StringDuration5h to be 5h, got %v", testStruct.StringDuration5h)
		}
		if testStruct.StringDuration30m != 30*time.Minute {
			t.Fatalf("expected StringDuration30m to be 30m, got %v", testStruct.StringDuration30m)
		}
		if testStruct.StringDuration1h30m != 1*time.Hour+30*time.Minute {
			t.Fatalf("expected StringDuration1h30m to be 1h30m, got %v", testStruct.StringDuration1h30m)
		}
		if testStruct.StringDuration500ms != 500*time.Millisecond {
			t.Fatalf("expected StringDuration500ms to be 500ms, got %v", testStruct.StringDuration500ms)
		}
		if testStruct.IntNanoseconds != time.Second {
			t.Fatalf("expected IntNanoseconds to be 1s, got %v", testStruct.IntNanoseconds)
		}
		if testStruct.Int64Nanoseconds != 2*time.Second {
			t.Fatalf("expected Int64Nanoseconds to be 2s, got %v", testStruct.Int64Nanoseconds)
		}
		if testStruct.Uint64Nanoseconds != 3*time.Second {
			t.Fatalf("expected Uint64Nanoseconds to be 3s, got %v", testStruct.Uint64Nanoseconds)
		}
		if testStruct.FloatNanoseconds != 4*time.Second {
			t.Fatalf("expected FloatNanoseconds to be 4s, got %v", testStruct.FloatNanoseconds)
		}
	})

	t.Run("should handle time.Time conversions", func(t *testing.T) {
		t.Parallel()

		type TestTimeStruct struct {
			RFC3339Time     time.Time `config:"rfc3339Time"`
			RFC3339NanoTime time.Time `config:"rfc3339NanoTime"`
			ISODate         time.Time `config:"isoDate"`
			UnixTimestamp   time.Time `config:"unixTimestamp"`
		}

		testStruct := TestTimeStruct{}

		conf, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
		conf.EnvironmentValues = map[string][]any{
			"rfc3339Time":     {"2025-01-15T10:30:00Z"},
			"rfc3339NanoTime": {"2025-01-15T10:30:00.123456789Z"},
			"isoDate":         {"2025-01-15"},
			"unixTimestamp":   {int64(1736938200)},
		}

		if err := conf.Get("", &testStruct); err != nil {
			t.Fatal(err)
		}

		expectedRFC3339, _ := time.Parse(time.RFC3339, "2025-01-15T10:30:00Z")
		if !testStruct.RFC3339Time.Equal(expectedRFC3339) {
			t.Fatalf("expected RFC3339Time to be %v, got %v", expectedRFC3339, testStruct.RFC3339Time)
		}

		expectedRFC3339Nano, _ := time.Parse(time.RFC3339Nano, "2025-01-15T10:30:00.123456789Z")
		if !testStruct.RFC3339NanoTime.Equal(expectedRFC3339Nano) {
			t.Fatalf("expected RFC3339NanoTime to be %v, got %v", expectedRFC3339Nano, testStruct.RFC3339NanoTime)
		}

		expectedISO, _ := time.Parse("2006-01-02", "2025-01-15")
		if !testStruct.ISODate.Equal(expectedISO) {
			t.Fatalf("expected ISODate to be %v, got %v", expectedISO, testStruct.ISODate)
		}

		expectedUnix := time.Unix(1736938200, 0)
		if !testStruct.UnixTimestamp.Equal(expectedUnix) {
			t.Fatalf("expected UnixTimestamp to be %v, got %v", expectedUnix, testStruct.UnixTimestamp)
		}
	})
}
func Test_Get_mapStringString(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"labels.env":     {"development"},
		"labels.version": {"1.0.0"},
		"labels.team":    {"backend"},
	}

	type Config struct {
		Labels map[string]string `config:"labels"`
	}

	var config Config
	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(config.Labels) != 3 {
		t.Errorf("map length = %d, want 3", len(config.Labels))
	}

	if config.Labels["env"] != "development" {
		t.Errorf("Labels[env] = %q, want %q", config.Labels["env"], "development")
	}

	if config.Labels["version"] != "1.0.0" {
		t.Errorf("Labels[version] = %q, want %q", config.Labels["version"], "1.0.0")
	}

	if config.Labels["team"] != "backend" {
		t.Errorf("Labels[team] = %q, want %q", config.Labels["team"], "backend")
	}
}

func Test_Get_mapStringInt(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.ConfigurationFiles = []*orale.File{
		{
			Path: "config.toml",
			Values: map[string][]any{
				"ports.http":  {8080},
				"ports.https": {8443},
				"ports.admin": {9000},
			},
		},
	}

	type Config struct {
		Ports map[string]int `config:"ports"`
	}

	var config Config
	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(config.Ports) != 3 {
		t.Errorf("map length = %d, want 3", len(config.Ports))
	}

	if config.Ports["http"] != 8080 {
		t.Errorf("Ports[http] = %d, want 8080", config.Ports["http"])
	}

	if config.Ports["https"] != 8443 {
		t.Errorf("Ports[https] = %d, want 8443", config.Ports["https"])
	}

	if config.Ports["admin"] != 9000 {
		t.Errorf("Ports[admin] = %d, want 9000", config.Ports["admin"])
	}
}

func Test_Get_nestedMapInStruct(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.EnvironmentValues = map[string][]any{
		"database.settings.maxConnections": {"100"},
		"database.settings.timeout":        {"30s"},
		"database.host":                    {"localhost"},
	}

	type Config struct {
		Database struct {
			Host     string            `config:"host"`
			Settings map[string]string `config:"settings"`
		} `config:"database"`
	}

	var config Config
	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if config.Database.Host != "localhost" {
		t.Errorf("Database.Host = %q, want localhost", config.Database.Host)
	}

	if len(config.Database.Settings) != 2 {
		t.Errorf("Database.Settings length = %d, want 2", len(config.Database.Settings))
	}

	if config.Database.Settings["maxConnections"] != "100" {
		t.Errorf("Settings[maxConnections] = %q, want 100", config.Database.Settings["maxConnections"])
	}

	if config.Database.Settings["timeout"] != "30s" {
		t.Errorf("Settings[timeout] = %q, want 30s", config.Database.Settings["timeout"])
	}
}

func Test_Get_emptyMap(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})

	type Config struct {
		Labels map[string]string `config:"labels"`
	}

	var config Config
	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if config.Labels == nil {
		t.Error("Labels should be initialized, not nil")
	}

	if len(config.Labels) != 0 {
		t.Errorf("Labels length = %d, want 0", len(config.Labels))
	}
}

func Test_Get_mapFromMultipleSources(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"env.nodeEnv": {"production"},
	}
	l.EnvironmentValues = map[string][]any{
		"env.debug": {"false"},
	}
	l.ConfigurationFiles = []*orale.File{
		{
			Path: "config.toml",
			Values: map[string][]any{
				"env.region": {"us-west-2"},
			},
		},
	}

	type Config struct {
		Env map[string]string `config:"env"`
	}

	var config Config
	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(config.Env) != 3 {
		t.Errorf("Env length = %d, want 3", len(config.Env))
	}

	if config.Env["nodeEnv"] != "production" {
		t.Errorf("Env[nodeEnv] = %q, want production", config.Env["nodeEnv"])
	}

	if config.Env["debug"] != "false" {
		t.Errorf("Env[debug] = %q, want false", config.Env["debug"])
	}

	if config.Env["region"] != "us-west-2" {
		t.Errorf("Env[region] = %q, want us-west-2", config.Env["region"])
	}
}

func Test_Get_mapStringBool(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"features.darkMode":   {"true"},
		"features.analytics":  {"false"},
		"features.betaAccess": {"1"},
	}

	type Config struct {
		Features map[string]bool `config:"features"`
	}

	var config Config
	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if config.Features["darkMode"] != true {
		t.Errorf("Features[darkMode] = %v, want true", config.Features["darkMode"])
	}

	if config.Features["analytics"] != false {
		t.Errorf("Features[analytics] = %v, want false", config.Features["analytics"])
	}

	if config.Features["betaAccess"] != true {
		t.Errorf("Features[betaAccess] = %v, want true", config.Features["betaAccess"])
	}
}

func Test_Get_mapPrecedence(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"config.key1": {"from-flag"},
	}
	l.EnvironmentValues = map[string][]any{
		"config.key1": {"from-env"},
		"config.key2": {"from-env"},
	}
	l.ConfigurationFiles = []*orale.File{
		{
			Path: "config.toml",
			Values: map[string][]any{
				"config.key1": {"from-file"},
				"config.key2": {"from-file"},
				"config.key3": {"from-file"},
			},
		},
	}

	type Config struct {
		Config map[string]string `config:"config"`
	}

	var config Config
	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if config.Config["key1"] != "from-flag" {
		t.Errorf("Config[key1] = %q, want from-flag (flags have highest precedence)", config.Config["key1"])
	}

	if config.Config["key2"] != "from-env" {
		t.Errorf("Config[key2] = %q, want from-env (environment has second precedence)", config.Config["key2"])
	}

	if config.Config["key3"] != "from-file" {
		t.Errorf("Config[key3] = %q, want from-file", config.Config["key3"])
	}
}
func Test_Get_errorMessageIncludesPath_int(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"server.port": {"not-a-number"},
	}

	type Config struct {
		Server struct {
			Port int `config:"port"`
		} `config:"server"`
	}

	var config Config
	err := l.GetAll(&config)

	if err == nil {
		t.Fatal("GetAll() should return error for invalid int conversion")
	}

	if !strings.Contains(err.Error(), "server.port") {
		t.Errorf("error message should include path 'server.port': %v", err)
	}
}

func Test_Get_errorMessageIncludesPath_bool(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.EnvironmentValues = map[string][]any{
		"app.debug": {"not-a-bool"},
	}

	type Config struct {
		App struct {
			Debug bool `config:"debug"`
		} `config:"app"`
	}

	var config Config
	err := l.GetAll(&config)

	if err == nil {
		t.Fatal("GetAll() should return error for invalid bool conversion")
	}

	if !strings.Contains(err.Error(), "app.debug") {
		t.Errorf("error message should include path 'app.debug': %v", err)
	}
}

func Test_Get_errorMessageIncludesPath_float(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.ConfigurationFiles = []*orale.File{
		{
			Path: "config.toml",
			Values: map[string][]any{
				"metrics.ratio": {"invalid-float"},
			},
		},
	}

	type Config struct {
		Metrics struct {
			Ratio float64 `config:"ratio"`
		} `config:"metrics"`
	}

	var config Config
	err := l.GetAll(&config)

	if err == nil {
		t.Fatal("GetAll() should return error for invalid float conversion")
	}

	if !strings.Contains(err.Error(), "metrics.ratio") {
		t.Errorf("error message should include path 'metrics.ratio': %v", err)
	}
}

func Test_Get_errorMessageIncludesPath_duration(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"server.timeout": {"not-a-duration"},
	}

	type Config struct {
		Server struct {
			Timeout int64 `config:"timeout"`
		} `config:"server"`
	}

	var config Config
	err := l.GetAll(&config)

	if err == nil {
		t.Fatal("GetAll() should return error for invalid duration conversion")
	}

	if !strings.Contains(err.Error(), "server.timeout") {
		t.Errorf("error message should include path 'server.timeout': %v", err)
	}
}

func Test_Get_errorMessageIncludesPath_unsupportedType(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"data.values": {"test"},
	}

	type Config struct {
		Data struct {
			Values chan string `config:"values"`
		} `config:"data"`
	}

	var config Config
	err := l.GetAll(&config)

	if err == nil {
		t.Fatal("GetAll() should return error for unsupported type")
	}

	if !strings.Contains(err.Error(), "data.values") {
		t.Errorf("error message should include path 'data.values': %v", err)
	}

	if !strings.Contains(err.Error(), "chan") {
		t.Errorf("error message should mention the unsupported type 'chan': %v", err)
	}
}

func Test_Get_errorMessageForNestedPath(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.ConfigurationFiles = []*orale.File{
		{
			Path: "config.toml",
			Values: map[string][]any{
				"api.endpoints.primary.timeout": {"invalid"},
			},
		},
	}

	type Config struct {
		Api struct {
			Endpoints struct {
				Primary struct {
					Timeout int `config:"timeout"`
				} `config:"primary"`
			} `config:"endpoints"`
		} `config:"api"`
	}

	var config Config
	err := l.GetAll(&config)

	if err == nil {
		t.Fatal("GetAll() should return error for invalid conversion")
	}

	if !strings.Contains(err.Error(), "api.endpoints.primary.timeout") {
		t.Errorf("error message should include full path 'api.endpoints.primary.timeout': %v", err)
	}
}
func Test_Get_missingValuesDoNotError(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"port": {"8080"},
	}

	type Config struct {
		Port    int     `config:"port"`
		Host    string  `config:"host"`
		Debug   bool    `config:"debug"`
		Timeout int64   `config:"timeout"`
		Ratio   float64 `config:"ratio"`
	}

	config := Config{
		Host:    "default-host",
		Debug:   true,
		Timeout: 30,
		Ratio:   1.5,
	}

	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() should not error for missing values, got: %v", err)
	}

	if config.Port != 8080 {
		t.Errorf("Port = %d, want 8080", config.Port)
	}

	if config.Host != "default-host" {
		t.Errorf("Host = %q, want default-host (should preserve default)", config.Host)
	}

	if config.Debug != true {
		t.Errorf("Debug = %v, want true (should preserve default)", config.Debug)
	}

	if config.Timeout != 30 {
		t.Errorf("Timeout = %d, want 30 (should preserve default)", config.Timeout)
	}

	if config.Ratio != 1.5 {
		t.Errorf("Ratio = %f, want 1.5 (should preserve default)", config.Ratio)
	}
}

func Test_Get_missingMapKeysDoNotError(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})

	type Config struct {
		Labels map[string]string `config:"labels"`
	}

	var config Config

	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() should not error for missing map, got: %v", err)
	}

	if config.Labels == nil {
		t.Error("Labels should be initialized to empty map, not nil")
	}

	if len(config.Labels) != 0 {
		t.Errorf("Labels should be empty, got %d entries", len(config.Labels))
	}
}

func Test_Get_partialMapDoesNotError(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"config.key1": {"value1"},
	}

	type Config struct {
		Config map[string]string `config:"config"`
		Other  string            `config:"other"`
	}

	config := Config{
		Other: "default-value",
	}

	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() should not error for partially missing values, got: %v", err)
	}

	if config.Config["key1"] != "value1" {
		t.Errorf("Config[key1] = %q, want value1", config.Config["key1"])
	}

	if config.Other != "default-value" {
		t.Errorf("Other = %q, want default-value (should preserve default)", config.Other)
	}
}

func Test_Get_encryptedDefaultValue(t *testing.T) {
	tempDir := t.TempDir()
	home := os.Getenv("HOME")
	defer os.Setenv("HOME", home)
	os.Setenv("HOME", tempDir)

	configDir := filepath.Join(tempDir, ".config", "orale")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}

	salt := make([]byte, 32)
	encryptionKey := make([]byte, 32)
	hmacKey := make([]byte, 32)

	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(encryptionKey); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(hmacKey); err != nil {
		t.Fatal(err)
	}

	keyData := make([]byte, 0, 96+len("test-key"))
	keyData = append(keyData, salt...)
	keyData = append(keyData, encryptionKey...)
	keyData = append(keyData, hmacKey...)
	keyData = append(keyData, []byte("test-key")...)

	keyPath := filepath.Join(configDir, "test-key.ok")
	if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
		t.Fatal(err)
	}

	plaintext := "secret-password"
	encrypted, err := orale.Encrypt([]byte(plaintext), encryptionKey, hmacKey)
	if err != nil {
		t.Fatal(err)
	}
	encryptedValue := "ENC[" + encrypted + "]"

	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})

	type Config struct {
		Password string `config:"password"`
	}

	config := Config{
		Password: encryptedValue,
	}

	err = l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if config.Password != plaintext {
		t.Errorf("Password = %q, want %q (encrypted default should be decrypted)", config.Password, plaintext)
	}
}

func Test_Get_encryptedDefaultValue_fallbackToEmptyWhenDecryptionFails(t *testing.T) {
	tempDir := t.TempDir()
	home := os.Getenv("HOME")
	defer os.Setenv("HOME", home)
	os.Setenv("HOME", tempDir)

	// Don't create any key files - decryption will fail

	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.DisableDecryptionWarnings = true

	type Config struct {
		Password string `config:"password"`
	}

	config := Config{
		Password: "ENC[someinvalidencryptedvalue]",
	}

	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() should not error when decryption fails, got: %v", err)
	}

	if config.Password != "" {
		t.Errorf("Password = %q, want empty string when decryption fails", config.Password)
	}
}

func Test_Get_registersSecretPaths(t *testing.T) {
	l, _ := orale.LoadFromValues([]string{}, "", []string{}, "", []string{})
	l.FlagValues = map[string][]any{
		"database.password": {"secret-password"},
		"apiKey":            {"secret-api-key"},
	}

	type Config struct {
		Database struct {
			Password string `config:"password" secret:"true"`
		} `config:"database"`
		APIKey string `config:"apiKey" secret:"true"`
	}

	var config Config
	err := l.GetAll(&config)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if config.Database.Password != "secret-password" {
		t.Errorf("Database.Password = %q, want secret-password", config.Database.Password)
	}
	if config.APIKey != "secret-api-key" {
		t.Errorf("APIKey = %q, want secret-api-key", config.APIKey)
	}
	if !l.SecretPaths["database.password"] {
		t.Error("SecretPaths[database.password] = false, want true")
	}
	if !l.SecretPaths["apiKey"] {
		t.Error("SecretPaths[apiKey] = false, want true")
	}
}
