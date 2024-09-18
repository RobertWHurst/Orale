package orale

import "unicode"

func toCamelCase(s string) string {
	l := len(s)
	newS := []rune{}
	for i := 0; i < l; i += 1 {
		var nextChar rune
		currentChar := rune(s[i])
		if i+1 < l {
			nextChar = rune(s[i+1])
		}
		if currentChar != '_' && currentChar != '-' {
			newS = append(newS, currentChar)
		} else {
			newS = append(newS, unicode.ToUpper(nextChar))
			i += 1
		}
	}
	return string(newS)
}
