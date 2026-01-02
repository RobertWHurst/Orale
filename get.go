package orale

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
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
		targetRefType := targetRefVal.Type()

		if targetRefType == reflect.TypeOf(time.Time{}) {
			value, err := resolveValue(l, currentPath)
			if err != nil {
				return err
			}
			if len(value) > index {
				timeValue, ok := intoTime(value[index])
				if ok {
					targetRefVal.Set(reflect.ValueOf(timeValue))
				}
			}
			return nil
		}

		for i := 0; i < targetRefVal.NumField(); i++ {
			field := targetRefVal.Field(i)
			structField := targetRefType.Field(i)

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
				if err := getFromLoader(l, fmt.Sprintf("%s.%d", currentPath, i), targetRefVal.Index(i), 0); err != nil {
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
		targetRefType := targetRefVal.Type()

		if targetRefType == reflect.TypeOf(time.Duration(0)) {
			value, err := resolveValue(l, currentPath)
			if err != nil {
				return err
			}
			if len(value) > index {
				duration, ok := intoDuration(value[index])
				if ok {
					targetRefVal.Set(reflect.ValueOf(duration))
				}
			}
			return nil
		}

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

	var value []any
	source := ""
	sourcePath := ""
	if v, ok := l.FlagValues[targetPath]; ok {
		value = v
		source = "flag"
	} else if v, ok := l.EnvironmentValues[targetPath]; ok {
		value = v
		source = "environment"
	} else {
		for _, file := range l.ConfigurationFiles {
			if v, ok := file.Values[targetPath]; ok {
				value = v
				source = "file"

				sourcePath = targetPath
				if workingDir, err := os.Getwd(); err == nil {
					if relPath, err := filepath.Rel(workingDir, sourcePath); err == nil {
						sourcePath = relPath
					}
				}
				break
			}
		}
	}

	var err error
	for i, v := range value {
		if strValue, ok := v.(string); ok && isEncrypted(strValue) {
			value[i], err = decryptValue(strValue)
			if err != nil {
				if sourcePath != "" {
					return nil, fmt.Errorf("cannot decrypt value from '%s(%s)': %w", source, sourcePath, err)
				}
				return nil, fmt.Errorf("cannot decrypt value from '%s': %w", source, err)
			}
		}

		value[i], err = expandValue(l, value[i])
		if err != nil {
			if sourcePath != "" {
				return nil, fmt.Errorf("cannot expand value from '%s(%s)': %w", source, sourcePath, err)
			}
			return nil, fmt.Errorf("cannot expand value from '%s': %w", source, err)
		}
	}

	return value, nil
}

func resolvePathLen(l *Loader, targetPath string) (int, error) {
	if targetPath == "" {
		return 0, fmt.Errorf("target path cannot be empty")
	}

	maxIndex := -1
	for flagPath := range l.FlagValues {
		slicePath := getSlicePathFromSubjectAndTargetPaths(flagPath, targetPath)
		if slicePath == "" {
			continue
		}
		lastDotIndex := strings.LastIndex(slicePath, ".")
		if lastDotIndex != -1 {
			if index, err := strconv.Atoi(slicePath[lastDotIndex+1:]); err == nil {
				if index > maxIndex {
					maxIndex = index
				}
			}
		}
	}
	if maxIndex >= 0 {
		return maxIndex + 1, nil
	}

	maxIndex = -1
	for environmentPath := range l.EnvironmentValues {
		slicePath := getSlicePathFromSubjectAndTargetPaths(environmentPath, targetPath)
		if slicePath != "" {
			lastDotIndex := strings.LastIndex(slicePath, ".")
			if lastDotIndex != -1 {
				if index, err := strconv.Atoi(slicePath[lastDotIndex+1:]); err == nil {
					if index > maxIndex {
						maxIndex = index
					}
				}
			}
		}
	}
	if maxIndex >= 0 {
		return maxIndex + 1, nil
	}

	for _, file := range l.ConfigurationFiles {
		maxIndex = -1
		for filePath := range file.Values {
			slicePath := getSlicePathFromSubjectAndTargetPaths(filePath, targetPath)
			if slicePath != "" {
				lastDotIndex := strings.LastIndex(slicePath, ".")
				if lastDotIndex != -1 {
					if index, err := strconv.Atoi(slicePath[lastDotIndex+1:]); err == nil {
						if index > maxIndex {
							maxIndex = index
						}
					}
				}
			}
		}
		if maxIndex >= 0 {
			return maxIndex + 1, nil
		}
	}

	return 0, nil
}

func getSlicePathFromSubjectAndTargetPaths(subjectPath, targetPath string) string {
	if len(subjectPath) < len(targetPath)+2 || !strings.HasPrefix(subjectPath, targetPath) {
		return ""
	}
	remainingPath := subjectPath[len(targetPath):]
	if remainingPath[0] != '.' {
		return ""
	}
	endIndexOffset := 1
	for i := 1; i < len(remainingPath); i += 1 {
		if remainingPath[i] >= '0' && remainingPath[i] <= '9' {
			endIndexOffset = i + 1
		} else {
			break
		}
	}
	if endIndexOffset == 1 {
		return ""
	}
	return subjectPath[:len(targetPath)+endIndexOffset]
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
