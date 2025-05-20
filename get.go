package orale

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// Get populates loaded configuration values into the target. The target must be
// a pointer to a variable. The target's value will be replaced with the
// loaded configuration values. Note that if the target contains paths which
// are not present in the loaded configuration values, those paths will be
// ignored allowing you to set defaults. Nil pointers will be initialized.
//
// Example:

// ```go
//
//	type TestConfig struct {
//		Database struct {
//			ConnectionUri string `config:"connection_uri"`
//		} `config:"database"`
//		Server struct {
//			Port int `config:"port"`
//		} `config:"server"`
//		Channels []struct {
//			Name string `config:"name"`
//			Id   int    `config:"id"`
//		} `config:"channels"`
//	}
//
//	loader, err := orale.Load("my-app")
//	if err != nil {
//		panic(err)
//	}
//
//	var testConfig TestConfig
//	if err := loader.Get("", &testConfig); err != nil {
//		panic(err)
//	}
//
// ```
//
// As you can see in the example above, the TestConfig struct is populated with
// the loaded configuration values. The property names of each field are
// specified by the `config` tag. If the `config` tag is not specified, the
// property name is converted to snake case. For example `ConnectionUri` becomes
// `connection_uri` path.
func (l *Loader) Get(path string, target any) error {
	targetRefVal := reflect.ValueOf(target)
	if targetRefVal.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}
	targetRefVal = targetRefVal.Elem()

	return getFromLoader(l, path, targetRefVal, 0)
}

// MustGet is the same as Get except it panics if an error occurs.
func (l *Loader) MustGet(path string, target any) {
	err := l.Get(path, target)
	if err != nil {
		panic(err)
	}
}

// GetAll populates loaded configuration values into the target value.
// This should be a pointer type, usually a pointer to a struct. All loaded
// configuration values will be set into the target value. Any default values
// in the target value that are not present in the loaded configuration
// will be left unchanged.
func (l *Loader) GetAll(target any) error {
	return l.Get("", target)
}

// MustGetAll is the same as GetAll except it panics if an error occurs.
func (l *Loader) MustGetAll(target any) {
	l.MustGet("", target)
}

func getFromLoader(l *Loader, currentPath string, targetRefVal reflect.Value, index int) error {
	switch targetRefVal.Kind() {
	case reflect.Ptr:
		if targetRefVal.IsNil() {
			targetRefVal.Set(reflect.New(targetRefVal.Type().Elem()))
		}
		return getFromLoader(l, currentPath, targetRefVal.Elem(), 0)

	case reflect.Struct:
		typ := targetRefVal.Type()
		for i := 0; i < targetRefVal.NumField(); i++ {
			field := targetRefVal.Field(i)
			structField := typ.Field(i)

			// Check if the field is exported
			if !field.CanSet() {
				continue
			}

			// Handle anonymous struct fields (embedded structs)
			if structField.Anonymous && field.Kind() == reflect.Struct {
				fieldTag := structField.Tag.Get("config")
				var embeddedPath string
				if fieldTag != "" {
					if currentPath != "" {
						embeddedPath = currentPath + "." + fieldTag
					} else {
						embeddedPath = fieldTag
					}
				} else {
					// If no 'config' tag, use currentPath (fields are promoted)
					embeddedPath = currentPath
				}
				// Recursively process the embedded struct
				if err := getFromLoader(l, embeddedPath, field, 0); err != nil {
					return err
				}
				continue
			}

			fieldTag := structField.Tag.Get("config")
			if fieldTag == "" {
				fieldTag = calDefaultFieldTag(structField.Name)
			}
			var fieldPath string
			if currentPath != "" {
				fieldPath = currentPath + "." + fieldTag
			} else {
				fieldPath = fieldTag
			}
			if err := getFromLoader(l, fieldPath, field, 0); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if targetRefVal.IsNil() {
			targetRefVal.Set(reflect.MakeSlice(targetRefVal.Type(), 0, 0))
		}
		valueLen, err := resolvePathLen(l, currentPath)
		if err != nil {
			return err
		}
		if valueLen > 0 {
			targetRefVal.Set(reflect.MakeSlice(targetRefVal.Type(), valueLen, valueLen))
			for i := 0; i < valueLen; i += 1 {
				if err := getFromLoader(l, fmt.Sprintf("%s[%d]", currentPath, i), targetRefVal.Index(i), 0); err != nil {
					return err
				}
			}
		} else {
			value, err := resolveValue(l, currentPath)
			if err != nil {
				return err
			}
			if value != nil {
				targetRefVal.Set(reflect.MakeSlice(targetRefVal.Type(), len(value), len(value)))
				for i := 0; i < len(value); i += 1 {
					if err := getFromLoader(l, currentPath, targetRefVal.Index(i), i); err != nil {
						return err
					}
				}
			} else {
				targetRefVal.Set(reflect.MakeSlice(targetRefVal.Type(), 0, 0))
			}
		}

	case reflect.String:
		value, err := resolveValue(l, currentPath)
		if err != nil {
			return err
		}
		if len(value) > index {
			strValue, ok := intoString(value[index])
			if ok {
				targetRefVal.SetString(strValue)
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := resolveValue(l, currentPath)
		if err != nil {
			return err
		}
		if len(value) > index {
			int64Value, ok := intoInt64(value[index])
			if ok {
				targetRefVal.SetInt(int64Value)
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := resolveValue(l, currentPath)
		if err != nil {
			return err
		}
		if len(value) > index {
			uint64Value, ok := intoUint64(value[index])
			if ok && len(value) > 0 {
				targetRefVal.SetUint(uint64Value)
			}
		}

	case reflect.Float32, reflect.Float64:
		value, err := resolveValue(l, currentPath)
		if err != nil {
			return err
		}
		if len(value) > index {
			float64Value, ok := intoFloat64(value[index])
			if ok {
				targetRefVal.SetFloat(float64Value)
			}
		}

	case reflect.Bool:
		value, err := resolveValue(l, currentPath)
		if err != nil {
			return err
		}
		if len(value) > index {
			if len(value) > 0 {
				val, ok := intoBool(value[index])
				if ok {
					targetRefVal.SetBool(val)
				}
			}
		}

	default:
		return fmt.Errorf("unsupported type %s", targetRefVal.Kind())
	}
	return nil
}

func resolveValue(l *Loader, targetPath string) ([]any, error) {
	if targetPath == "" {
		return nil, fmt.Errorf("target path cannot be empty")
	}
	if value, ok := l.FlagValues[targetPath]; ok {
		return value, nil
	} else if value, ok := l.EnvironmentValues[targetPath]; ok {
		return value, nil
	} else {
		for _, file := range l.ConfigurationFiles {
			if value, ok := file.Values[targetPath]; ok {
				return value, nil
			}
		}
	}
	return nil, nil
}

func resolvePathLen(l *Loader, targetPath string) (int, error) {
	if targetPath == "" {
		return 0, fmt.Errorf("target path cannot be empty")
	}

	flagPaths := map[string]bool{}
	for flagPath := range l.FlagValues {
		slicePath := getSlicePathFromSubjectAndTargetPaths(flagPath, targetPath)
		if slicePath != "" {
			flagPaths[slicePath] = true
		}
	}
	if len(flagPaths) != 0 {
		return len(flagPaths), nil
	}

	environmentPaths := map[string]bool{}
	for environmentPath := range l.EnvironmentValues {
		slicePath := getSlicePathFromSubjectAndTargetPaths(environmentPath, targetPath)
		if slicePath != "" {
			environmentPaths[slicePath] = true
		}
	}
	if len(environmentPaths) != 0 {
		return len(environmentPaths), nil
	}

	for _, file := range l.ConfigurationFiles {
		filePaths := map[string]bool{}
		for filePath := range file.Values {
			slicePath := getSlicePathFromSubjectAndTargetPaths(filePath, targetPath)
			if slicePath != "" {
				filePaths[slicePath] = true
			}
		}
		if len(filePaths) != 0 {
			return len(filePaths), nil
		}
	}

	return 0, nil
}

func getSlicePathFromSubjectAndTargetPaths(subjectPath, targetPath string) string {
	if len(subjectPath) < len(targetPath)+3 || !strings.HasPrefix(subjectPath, targetPath) {
		return ""
	}
	remainingPath := subjectPath[len(targetPath):]
	if remainingPath[0] != '[' {
		return ""
	}
	endIndexOffset := 0
	for i, r := range remainingPath {
		if r == ']' {
			endIndexOffset = i
			break
		}
	}
	return subjectPath[:len(targetPath)+endIndexOffset+1]
}

func calDefaultFieldTag(fieldName string) string {
	fieldTag := ""
	for i, r := range fieldName {
		if unicode.IsUpper(r) {
			if i != 0 {
				fieldTag += "_"
			}
			fieldTag += strings.ToLower(string(r))
		} else {
			fieldTag += string(r)
		}
	}
	return fieldTag
}

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
