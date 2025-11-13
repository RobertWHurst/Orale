package orale

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func intoString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), true
	case float32, float64:
		return fmt.Sprintf("%g", v), true
	case bool:
		return fmt.Sprintf("%t", v), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

func intoInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return int64(v), true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case uint64:
		if v > math.MaxInt64 {
			return 0, false
		}
		return int64(v), true
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, true
		}
		if u, err := strconv.ParseFloat(v, 64); err == nil {
			if u > math.MaxInt64 {
				return 0, false
			}
			return int64(u), true
		}
		return 0, false
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func intoUint64(value any) (uint64, bool) {
	switch v := value.(type) {
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return uint64(v), true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int8:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int16:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int32:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case float32:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case float64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case string:
		if u, err := strconv.ParseUint(v, 10, 64); err == nil {
			return u, true
		}
		if i, err := strconv.ParseFloat(v, 64); err == nil {
			if i < 0 {
				return 0, false
			}
			return uint64(i), true
		}
		return 0, false
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func intoFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
		return 0, false
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func intoBool(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case int:
		return v != 0, true
	case int8:
		return v != 0, true
	case int16:
		return v != 0, true
	case int32:
		return v != 0, true
	case int64:
		return v != 0, true
	case uint:
		return v != 0, true
	case uint8:
		return v != 0, true
	case uint16:
		return v != 0, true
	case uint32:
		return v != 0, true
	case uint64:
		return v != 0, true
	case float32:
		return v != 0, true
	case float64:
		return v != 0, true
	case string:
		lower := strings.ToLower(v)
		if lower == "true" || lower == "t" || lower == "yes" || lower == "y" || lower == "1" {
			return true, true
		}
		if lower == "false" || lower == "f" || lower == "no" || lower == "n" || lower == "0" {
			return false, true
		}
		return false, false
	default:
		return false, false
	}
}

func intoTime(value any) (time.Time, bool) {
	switch v := value.(type) {
	case string:
		// Try common time formats (most specific first)
		formats := []string{
			time.RFC3339Nano,      // Most specific ISO 8601 with nanoseconds
			time.RFC3339,          // ISO 8601 without nanoseconds
			time.RFC1123Z,         // RFC1123 with numeric timezone
			time.RFC1123,          // RFC1123 with text timezone
			time.RFC822Z,          // RFC822 with numeric timezone
			time.RFC822,           // RFC822 with text timezone
			time.RFC850,           // Longer date format
			time.RubyDate,         // Ruby's time format
			time.UnixDate,         // Unix date format
			time.ANSIC,            // ANSI C format
			"2006-01-02 15:04:05", // Date-time without timezone (more specific)
			"2006-01-02",          // Date only (less specific)
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, true
			}
		}
		return time.Time{}, false
	case int, int64:
		// Treat as Unix timestamp (seconds)
		var sec int64
		switch val := v.(type) {
		case int:
			sec = int64(val)
		case int64:
			sec = val
		}
		return time.Unix(sec, 0), true
	case float64:
		// Treat as Unix timestamp with fractional seconds
		sec := int64(v)
		nsec := int64((v - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), true
	default:
		return time.Time{}, false
	}
}

func intoDuration(value any) (time.Duration, bool) {
	switch v := value.(type) {
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d, true
		}
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Duration(i), true
		}
		return 0, false
	case int:
		return time.Duration(v), true
	case int8:
		return time.Duration(v), true
	case int16:
		return time.Duration(v), true
	case int32:
		return time.Duration(v), true
	case int64:
		return time.Duration(v), true
	case uint:
		return time.Duration(v), true
	case uint8:
		return time.Duration(v), true
	case uint16:
		return time.Duration(v), true
	case uint32:
		return time.Duration(v), true
	case uint64:
		if v > math.MaxInt64 {
			return 0, false
		}
		return time.Duration(v), true
	case float32:
		return time.Duration(v), true
	case float64:
		return time.Duration(v), true
	default:
		return 0, false
	}
}
