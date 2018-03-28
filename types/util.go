package types

import (
	"unicode"
	"unicode/utf8"
)

func IsConsName(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
