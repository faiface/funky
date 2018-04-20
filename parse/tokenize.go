package parse

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/faiface/funky/parse/parseinfo"
)

var SpecialRunes = []rune{'(', ')', '[', ']', '{', '}', ',', ';', '\\', 'λ', '#'}

func IsSpecialRune(r rune) bool {
	for _, special := range SpecialRunes {
		if r == special {
			return true
		}
	}
	return false
}

type Token struct {
	SourceInfo *parseinfo.Source
	Value      string
}

func Tokenize(filename, s string) ([]Token, error) {
	var tokens []Token

	si := &parseinfo.Source{
		Filename: filename,
		Line:     1,
		Column:   1,
	}

	for {
		// skip whitespace
		for len(s) > 0 {
			r, size := utf8.DecodeRuneInString(s)
			if !unicode.IsSpace(r) {
				break
			}
			s = s[size:]
			updateSIInPlace(si, r)
		}

		if len(s) == 0 {
			break
		}

		// handle special runes and comments
		r, size := utf8.DecodeRuneInString(s)
		if r == '#' {
			// comment, skip until end of line
			for len(s) > 0 {
				r, size := utf8.DecodeRuneInString(s)
				if r == '\n' {
					break
				}
				s = s[size:]
				updateSIInPlace(si, r)
			}
			continue
		}
		if IsSpecialRune(r) {
			tokens = append(tokens, Token{SourceInfo: si, Value: string(r)})
			s = s[size:]
			si = updateSI(si, r)
			continue
		}

		// handle chars and strings
		if r == '\'' || r == '"' {
			quote := byte(r)
			quoteSI := copySI(si)
			s = s[1:]
			updateSIInPlace(si, r)

			var builder strings.Builder
			builder.WriteByte(quote)

			for len(s) > 0 && s[0] != quote {
				_, _, tail, err := strconv.UnquoteChar(s, quote)
				if err != nil {
					return nil, &Error{si, err.Error()}
				}

				for len(s) > len(tail) {
					r, size := utf8.DecodeRuneInString(s)
					s = s[size:]
					updateSIInPlace(si, r)
					builder.WriteRune(r)
				}
			}

			if len(s) == 0 {
				return nil, &Error{si, "unclosed char or string"}
			}

			s = s[1:] // closing quote
			updateSIInPlace(si, rune(quote))
			builder.WriteByte(quote)

			tokens = append(tokens, Token{quoteSI, builder.String()})
			continue
		}

		// accumulate token until whitespace or special rune
		value := ""
		for len(s) > 0 {
			r, size := utf8.DecodeRuneInString(s)
			if unicode.IsSpace(r) || IsSpecialRune(r) {
				break
			}
			value += string(r)
			s = s[size:]
			updateSIInPlace(si, r)
		}
		tokenSI := copySI(si)
		tokenSI.Column -= utf8.RuneCountInString(value)
		tokens = append(tokens, Token{SourceInfo: tokenSI, Value: value})
	}

	return tokens, nil
}

func updateSIInPlace(si *parseinfo.Source, r rune) {
	if r == '\n' {
		si.Line++
		si.Column = 1
	} else {
		si.Column++
	}
}

func copySI(si *parseinfo.Source) *parseinfo.Source {
	newSI := &parseinfo.Source{}
	*newSI = *si
	return newSI
}

func updateSI(si *parseinfo.Source, r rune) *parseinfo.Source {
	newSI := copySI(si)
	updateSIInPlace(newSI, r)
	return newSI
}
