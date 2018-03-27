package parse

import (
	"unicode"
	"unicode/utf8"

	"github.com/faiface/funky/expr"
)

type Token struct {
	SourceInfo *expr.SourceInfo
	Value      string
}

func Tokenize(filename, s string) []Token {
	var tokens []Token
	for token := range tokenize(filename, s) {
		tokens = append(tokens, token)
	}
	return tokens
}

func tokenize(filename, s string) <-chan Token {
	ch := make(chan Token)

	go func() {
		defer close(ch)

		si := &expr.SourceInfo{
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
				return
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
				ch <- Token{SourceInfo: si, Value: string(r)}
				s = s[size:]
				si = updateSI(si, r)
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
			ch <- Token{SourceInfo: tokenSI, Value: value}
		}
	}()

	return ch
}

func updateSIInPlace(si *expr.SourceInfo, r rune) {
	if r == '\n' {
		si.Line++
		si.Column = 1
	} else {
		si.Column++
	}
}

func copySI(si *expr.SourceInfo) *expr.SourceInfo {
	newSI := &expr.SourceInfo{}
	*newSI = *si
	return newSI
}

func updateSI(si *expr.SourceInfo, r rune) *expr.SourceInfo {
	newSI := copySI(si)
	updateSIInPlace(newSI, r)
	return newSI
}
