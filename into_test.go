package orale

import (
	"math"
	"testing"
	"time"
)

func Test_IntoInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    int64
		wantOk  bool
	}{
		// Integer types
		{"int", int(42), 42, true},
		{"int8", int8(42), 42, true},
		{"int16", int16(42), 42, true},
		{"int32", int32(42), 42, true},
		{"int64", int64(42), 42, true},
		{"uint", uint(42), 42, true},
		{"uint8", uint8(42), 42, true},
		{"uint16", uint16(42), 42, true},
		{"uint32", uint32(42), 42, true},

		// Float types
		{"float32", float32(42.7), 42, true},
		{"float64", float64(42.7), 42, true},

		// uint64 edge cases
		{"uint64 small", uint64(100), 100, true},
		{"uint64 overflow", uint64(math.MaxUint64), 0, false},

		// String parsing
		{"string int", "42", 42, true},
		{"string negative", "-42", -42, true},
		{"string float", "42.7", 42, true},
		{"string invalid", "not a number", 0, false},
		{"string overflow", "9999999999999999999999", 0, false},

		// Bool
		{"bool true", true, 1, true},
		{"bool false", false, 0, true},

		// Invalid types
		{"invalid type", []int{1, 2}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := intoInt64(tt.input)
			if ok != tt.wantOk {
				t.Errorf("intoInt64(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("intoInt64(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_IntoUint64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    uint64
		wantOk  bool
	}{
		// Unsigned types
		{"uint", uint(42), 42, true},
		{"uint8", uint8(42), 42, true},
		{"uint16", uint16(42), 42, true},
		{"uint32", uint32(42), 42, true},
		{"uint64", uint64(42), 42, true},

		// Signed types (positive)
		{"int positive", int(42), 42, true},
		{"int8 positive", int8(42), 42, true},
		{"int16 positive", int16(42), 42, true},
		{"int32 positive", int32(42), 42, true},
		{"int64 positive", int64(42), 42, true},

		// Signed types (negative - should fail)
		{"int negative", int(-42), 0, false},
		{"int8 negative", int8(-42), 0, false},
		{"int16 negative", int16(-42), 0, false},
		{"int32 negative", int32(-42), 0, false},
		{"int64 negative", int64(-42), 0, false},

		// Float types (positive)
		{"float32 positive", float32(42.7), 42, true},
		{"float64 positive", float64(42.7), 42, true},

		// Float types (negative - should fail)
		{"float32 negative", float32(-42.7), 0, false},
		{"float64 negative", float64(-42.7), 0, false},

		// String parsing
		{"string uint", "42", 42, true},
		{"string float", "42.7", 42, true},
		{"string negative", "-42", 0, false},
		{"string invalid", "not a number", 0, false},

		// Bool
		{"bool true", true, 1, true},
		{"bool false", false, 0, true},

		// Invalid types
		{"invalid type", []int{1, 2}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := intoUint64(tt.input)
			if ok != tt.wantOk {
				t.Errorf("intoUint64(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("intoUint64(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_IntoFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    float64
		wantOk  bool
	}{
		// Float types
		{"float64", float64(42.7), 42.7, true},
		{"float32", float32(42.7), float64(float32(42.7)), true},

		// Integer types
		{"int", int(42), 42.0, true},
		{"int8", int8(42), 42.0, true},
		{"int16", int16(42), 42.0, true},
		{"int32", int32(42), 42.0, true},
		{"int64", int64(42), 42.0, true},
		{"uint", uint(42), 42.0, true},
		{"uint8", uint8(42), 42.0, true},
		{"uint16", uint16(42), 42.0, true},
		{"uint32", uint32(42), 42.0, true},
		{"uint64", uint64(42), 42.0, true},

		// String parsing
		{"string float", "42.7", 42.7, true},
		{"string int", "42", 42.0, true},
		{"string scientific", "1.23e5", 123000.0, true},
		{"string negative", "-42.7", -42.7, true},
		{"string invalid", "not a number", 0, false},

		// Bool
		{"bool true", true, 1.0, true},
		{"bool false", false, 0.0, true},

		// Invalid types
		{"invalid type", []int{1, 2}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := intoFloat64(tt.input)
			if ok != tt.wantOk {
				t.Errorf("intoFloat64(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("intoFloat64(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_IntoBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    bool
		wantOk  bool
	}{
		// Bool
		{"bool true", true, true, true},
		{"bool false", false, false, true},

		// Integer types (zero = false, non-zero = true)
		{"int zero", int(0), false, true},
		{"int non-zero", int(42), true, true},
		{"int8 zero", int8(0), false, true},
		{"int8 non-zero", int8(-1), true, true},
		{"int16", int16(100), true, true},
		{"int32", int32(0), false, true},
		{"int64", int64(-42), true, true},
		{"uint", uint(1), true, true},
		{"uint8", uint8(0), false, true},
		{"uint16", uint16(1), true, true},
		{"uint32", uint32(0), false, true},
		{"uint64", uint64(100), true, true},

		// Float types
		{"float32 zero", float32(0), false, true},
		{"float32 non-zero", float32(0.1), true, true},
		{"float64 zero", float64(0), false, true},
		{"float64 non-zero", float64(-1.5), true, true},

		// String parsing
		{"string true", "true", true, true},
		{"string t", "t", true, true},
		{"string yes", "yes", true, true},
		{"string y", "y", true, true},
		{"string 1", "1", true, true},
		{"string false", "false", false, true},
		{"string f", "f", false, true},
		{"string no", "no", false, true},
		{"string n", "n", false, true},
		{"string 0", "0", false, true},
		{"string TRUE", "TRUE", true, true},
		{"string False", "False", false, true},
		{"string invalid", "maybe", false, false},

		// Invalid types
		{"invalid type", []int{1, 2}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := intoBool(tt.input)
			if ok != tt.wantOk {
				t.Errorf("intoBool(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("intoBool(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_IntoTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantOk  bool
	}{
		// String formats
		{"RFC3339", "2025-01-15T10:30:00Z", true},
		{"RFC3339Nano", "2025-01-15T10:30:00.123456789Z", true},
		{"Date only", "2025-01-15", true},
		{"DateTime", "2025-01-15 10:30:00", true},
		{"RFC1123Z", "Wed, 15 Jan 2025 10:30:00 +0000", true},
		{"Invalid string", "not a date", false},

		// Unix timestamps
		{"int timestamp", int(1736938200), true},
		{"int64 timestamp", int64(1736938200), true},
		{"float64 timestamp", float64(1736938200.5), true},

		// Invalid types
		{"invalid type", []int{1, 2}, false},
		{"bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := intoTime(tt.input)
			if ok != tt.wantOk {
				t.Errorf("intoTime(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
		})
	}
}

func Test_IntoDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    time.Duration
		wantOk  bool
	}{
		// String formats
		{"string 5s", "5s", 5 * time.Second, true},
		{"string 5m", "5m", 5 * time.Minute, true},
		{"string 1h30m", "1h30m", 90 * time.Minute, true},
		{"string nanoseconds", "1000000000", time.Second, true},
		{"string invalid", "not a duration", 0, false},

		// Integer types (nanoseconds)
		{"int", int(1000000000), time.Second, true},
		{"int8", int8(100), 100 * time.Nanosecond, true},
		{"int16", int16(1000), 1000 * time.Nanosecond, true},
		{"int32", int32(1000000), time.Millisecond, true},
		{"int64", int64(1000000000), time.Second, true},
		{"uint", uint(1000), 1000 * time.Nanosecond, true},
		{"uint8", uint8(100), 100 * time.Nanosecond, true},
		{"uint16", uint16(1000), 1000 * time.Nanosecond, true},
		{"uint32", uint32(1000000), time.Millisecond, true},
		{"uint64 small", uint64(1000), 1000 * time.Nanosecond, true},
		{"uint64 overflow", uint64(math.MaxUint64), 0, false},

		// Float types
		{"float32", float32(1000000), time.Millisecond, true},
		{"float64", float64(1000000000), time.Second, true},

		// Invalid types
		{"invalid type", []int{1, 2}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := intoDuration(tt.input)
			if ok != tt.wantOk {
				t.Errorf("intoDuration(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && got != tt.want {
				t.Errorf("intoDuration(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_IntoString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    string
	}{
		{"string", "hello", "hello"},
		{"int", int(42), "42"},
		{"int64", int64(42), "42"},
		{"float64", float64(42.5), "42.5"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"uint", uint(100), "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := intoString(tt.input)
			if !ok {
				t.Errorf("intoString(%v) ok = false, want true", tt.input)
			}
			if got != tt.want {
				t.Errorf("intoString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
