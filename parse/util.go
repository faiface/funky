package parse

import "unicode/utf8"

var SpecialRunes = []rune{'(', ')', '[', ']', '{', '}', ',', ';', '\\', 'λ', '#'}

func IsSpecialRune(r rune) bool {
	for _, special := range SpecialRunes {
		if r == special {
			return true
		}
	}
	return false
}

func IsSpecial(s string) bool {
	if utf8.RuneCountInString(s) != 1 {
		return false
	}
	return IsSpecialRune([]rune(s)[0])
}
