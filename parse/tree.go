package parse

import (
	"unicode"

	"github.com/faiface/funky/parse/parseinfo"
)

type Tree interface {
	SourceInfo() *parseinfo.Source
}

type (
	Literal struct {
		SI    *parseinfo.Source
		Value string
	}

	Paren struct {
		SI     *parseinfo.Source
		Type   string
		Inside Tree
	}

	Special struct {
		SI    *parseinfo.Source
		Type  string
		After Tree
	}

	Lambda struct {
		SI           *parseinfo.Source
		Bound, After Tree
	}

	Prefix struct {
		Left, Right Tree
	}

	Infix struct {
		Left, In, Right Tree
	}
)

func (l *Literal) SourceInfo() *parseinfo.Source { return l.SI }
func (p *Paren) SourceInfo() *parseinfo.Source   { return p.SI }
func (s *Special) SourceInfo() *parseinfo.Source { return s.SI }
func (l *Lambda) SourceInfo() *parseinfo.Source  { return l.SI }
func (p *Prefix) SourceInfo() *parseinfo.Source  { return p.Left.SourceInfo() }
func (i *Infix) SourceInfo() *parseinfo.Source   { return i.In.SourceInfo() }

func SingleTree(tokens []Token) (t Tree, end int, err error) {
	switch tokens[0].Value {
	case ")", "]", "}":
		return nil, 0, &Error{
			tokens[0].SourceInfo,
			"no matching opening parenthesis",
		}

	case "(", "[", "{":
		closing, ok := findClosingParen(tokens)
		if !ok {
			return nil, 0, &Error{
				tokens[0].SourceInfo,
				"no matching closing parenthesis",
			}
		}
		inside, err := MultiTree(tokens[1:closing])
		if err != nil {
			return nil, 0, err
		}
		if inside == nil {
			return nil, 0, &Error{
				tokens[0].SourceInfo,
				"nothing inside parentheses",
			}
		}
		paren := &Paren{Type: tokens[0].Value, Inside: inside}
		return paren, closing + 1, nil

	case "\\", "λ":
		if len(tokens) < 2 {
			return nil, 0, &Error{
				tokens[0].SourceInfo,
				"nothing to bind",
			}
		}
		bound, end, err := SingleTree(tokens[1:])
		if err != nil {
			return nil, 0, err
		}
		after, err := MultiTree(tokens[end+1:])
		if err != nil {
			return nil, 0, err
		}
		if after == nil {
			return nil, 0, &Error{
				tokens[0].SourceInfo,
				"nothing after lambda binding",
			}
		}
		return &Lambda{
			SI:    tokens[0].SourceInfo,
			Bound: bound,
			After: after,
		}, len(tokens), nil

	case ",", ";", ":", "|":
		after, err := MultiTree(tokens[1:])
		if err != nil {
			return nil, 0, err
		}
		return &Special{
			SI:    tokens[0].SourceInfo,
			Type:  tokens[0].Value,
			After: after,
		}, len(tokens), nil

	default:
		if !hasLetterOrDigit(tokens[0].Value) {
			after, err := MultiTree(tokens[1:])
			if err != nil {
				return nil, 0, err
			}
			return &Infix{
				In:    &Literal{SI: tokens[0].SourceInfo, Value: tokens[0].Value},
				Right: after,
			}, len(tokens), nil
		}
		return &Literal{
			SI:    tokens[0].SourceInfo,
			Value: tokens[0].Value,
		}, 1, nil
	}
}

func hasLetterOrDigit(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func MultiTree(tokens []Token) (Tree, error) {
	var t Tree

	for len(tokens) > 0 {
		single, end, err := SingleTree(tokens)
		tokens = tokens[end:]
		if err != nil {
			return nil, err
		}
		if t == nil {
			t = single
			continue
		}
		if infix, ok := single.(*Infix); ok {
			t = &Infix{
				Left:  t,
				In:    infix.In,
				Right: infix.Right,
			}
			continue
		}
		t = &Prefix{Left: t, Right: single}
	}

	return t, nil
}
