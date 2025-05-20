package orale_test

import (
	"testing"

	orale "github.com/RobertWHurst/orale"
)

func newTestLoaderSingleValues() *orale.Loader {
	return &orale.Loader{
		FlagValues: map[string][]any{
			"a": {"1"},
			"b": {"2"},
			"c": {"3"},
			"d": {"4"},
		},
		EnvironmentValues: map[string][]any{
			"b": {"5"},
			"e": {"6"},
		},
		ConfigurationFiles: []*orale.File{
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
		},
	}
}

func newTestLoaderMultiValues() *orale.Loader {
	return &orale.Loader{
		FlagValues: map[string][]any{
			"a": {"1", "2"},
			"b": {"3", "4"},
			"c": {"5", "6"},
			"d": {"7", "8"},
		},
		EnvironmentValues: map[string][]any{
			"b": {"9", "10"},
			"e": {"11", "12"},
		},
		ConfigurationFiles: []*orale.File{},
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	
	t.Run("should handle type conversions from environment variables", func(t *testing.T) {
		t.Parallel()

		type TestConversionStruct struct {
			// Int conversions
			StringToInt    int  `config:"stringToInt"`
			FloatToInt     int  `config:"floatToInt"`
			BoolTrueToInt  int  `config:"boolTrueToInt"`
			BoolFalseToInt int  `config:"boolFalseToInt"`
			
			// Uint conversions
			StringToUint    uint  `config:"stringToUint"`
			FloatToUint     uint  `config:"floatToUint"`
			BoolTrueToUint  uint  `config:"boolTrueToUint"`
			BoolFalseToUint uint  `config:"boolFalseToUint"`
			
			// Float conversions
			StringToFloat    float64  `config:"stringToFloat"`
			IntToFloat       float64  `config:"intToFloat"`
			BoolTrueToFloat  float64  `config:"boolTrueToFloat"`
			BoolFalseToFloat float64  `config:"boolFalseToFloat"`
			
			// Bool conversions
			StringTrueToBool   bool  `config:"stringTrueToBool"`
			StringYesToBool    bool  `config:"stringYesToBool"`
			StringOneToBool    bool  `config:"stringOneToBool"`
			StringFalseToBool  bool  `config:"stringFalseToBool"`
			StringNoToBool     bool  `config:"stringNoToBool"`
			StringZeroToBool   bool  `config:"stringZeroToBool"`
			IntOneToBool       bool  `config:"intOneToBool"`
			IntZeroToBool      bool  `config:"intZeroToBool"`
			
			// String conversions
			IntToString      string  `config:"intToString"`
			FloatToString    string  `config:"floatToString"`
			BoolTrueToString string  `config:"boolTrueToString"`
			BoolFalseToString string  `config:"boolFalseToString"`
		}

		testStruct := TestConversionStruct{}

		conf := &orale.Loader{
			EnvironmentValues: map[string][]any{
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
			},
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

		conf := &orale.Loader{
			FlagValues: map[string][]any{
				"a": {"1"},
				"b": {"2"},
				"d": {"4"},
			},
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
}
