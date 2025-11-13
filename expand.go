package orale

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func expandValue(l *Loader, rawValue any) (any, error) {
	value, ok := rawValue.(string)
	if !ok {
		return rawValue, nil
	}

	if _, ok := l.expandTargets[value]; ok {
		return nil, errors.New("circular variable expansion detected")
	}
	l.expandTargets[value] = struct{}{}
	defer delete(l.expandTargets, value)

	if expandedValue, ok := l.expandCache[value]; ok {
		// If we determine that a value is the same as it's expanded form we do
		// not add it's value to the map because the key is the same, so we
		// simply return the key (the unexpanded value) instead.
		if expandedValue == "" {
			return value, nil
		}
		return expandedValue, nil
	}

	isExpanded := false
	in := []rune(value)
	out := []rune{}

	for i := 0; i < len(in); i += 1 {

		switch {
		// ${...} - Environment Variable
		case in[i] == '$' && i < len(in)-3 && in[i+1] == '{':
			s := i
			i += 2
			e := -1
			for ; i < len(in); i += 1 {
				if in[i] == '}' {
					e = i
					break
				}
			}
			if e == -1 {
				return nil, expandError(value, s, i, fmt.Errorf("environment variable has no closing brace"))
			}

			k := string(in[s+2 : e])
			v := os.Getenv(k)

			out = append(out, []rune(v)...)
			isExpanded = true

		// %{...} - Config Variable
		case in[i] == '%' && i < len(in)-3 && in[i+1] == '{':
			s := i
			i += 2
			e := -1
			for ; i < len(in); i += 1 {
				if in[i] == '}' {
					e = i
					break
				}
			}
			if e == -1 {
				return nil, expandError(value, s, i, fmt.Errorf("configuration variable has no closing brace"))
			}

			p := string(in[s+2 : e])
			vs, err := resolveValue(l, p)
			if err != nil {
				return nil, expandError(value, s, e, fmt.Errorf("configuration variable could not be resolved: %w", err))
			}

			v := ""
			if len(vs) >= 1 {
				va := vs[0]
				if vs, ok := va.(string); ok {
					v = vs
				}
			}

			out = append(out, []rune(v)...)
			isExpanded = true

		// @{...} - File Variable
		case in[i] == '@' && i < len(in)-3 && in[i+1] == '{':
			s := i
			i += 2
			e := -1
			for ; i < len(in); i += 1 {
				if in[i] == '}' {
					e = i
					break
				}
			}
			if e == -1 {
				return nil, expandError(value, s, i, fmt.Errorf("file variable has no closing brace"))
			}

			rp := string(in[s+2 : e])
			ap, err := filepath.Abs(rp)
			if err != nil {
				return nil, expandError(value, s, e, fmt.Errorf("file variable path could not be made absolute: %w", err))
			}

			fv, err := os.ReadFile(ap)
			if err != nil {
				return nil, expandError(value, s, e, fmt.Errorf("file variable path could not be read: %w", err))
			}
			if !utf8.Valid(fv) {
				return nil, expandError(value, s, e, fmt.Errorf("file variable path contents are not valid UTF-8"))
			}
			v := strings.TrimSuffix(string(fv), "\n")

			out = append(out, []rune(v)...)
			isExpanded = true

		// Non-Variable Characters
		default:
			out = append(out, in[i])
		}
	}

	finalValue := string(out)

	if isExpanded {
		l.expandCache[value] = finalValue
	} else {
		l.expandCache[value] = ""
	}

	return finalValue, nil
}

func expandError(value string, startIndex, endIndex int, err error) error {
	cL := 5
	preStart := max(startIndex-cL, 0)
	postEnd := min(endIndex+cL, len(value))
	snippet := value[preStart:postEnd]
	if len(snippet) > 20 {
		snipPre := snippet[:9]
		snipPost := snippet[len(snippet)-8:]
		snippet = snipPre + "..." + snipPost
	}
	return fmt.Errorf("expand error '%s': %w", snippet, err)
}
