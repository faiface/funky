package parse

import (
	"fmt"
	"unicode"

	"github.com/faiface/funky/parse/parseinfo"
)

type Error struct {
	SourceInfo *parseinfo.Source
	Msg        string
}

func (err *Error) Error() string {
	return fmt.Sprintf("%v: %v", err.SourceInfo, err.Msg)
}

type Tree interface {
	String() string
	SourceInfo() *parseinfo.Source
}

type (
	Literal struct {
		SI    *parseinfo.Source
		Value string
	}

	Paren struct {
		SI     *parseinfo.Source
		Kind   string
		Inside Tree
	}

	Special struct {
		SI    *parseinfo.Source
		Kind  string
		After Tree
	}

	Binding struct {
		SI           *parseinfo.Source
		Kind         string
		Bound, After Tree
	}

	Prefix struct {
		Left, Right Tree
	}

	Infix struct {
		Left, In, Right Tree
	}
)

func (l *Literal) String() string { return l.Value }
func (p *Paren) String() string {
	switch p.Kind {
	case "(":
		return "(" + p.Inside.String() + ")"
	case "[":
		return "[" + p.Inside.String() + "]"
	case "{":
		return "{" + p.Inside.String() + "}"
	}
	panic("unreachable")
}
func (s *Special) String() string { return s.Kind + " " + s.After.String() }
func (l *Binding) String() string { return l.Kind + l.Bound.String() + " " + l.After.String() }
func (p *Prefix) String() string  { return p.Left.String() + " " + p.Right.String() }
func (i *Infix) String() string {
	switch {
	case i.Left == nil && i.Right == nil:
		return i.In.String()
	case i.Left == nil:
		return i.In.String() + " " + i.Right.String()
	case i.Right == nil:
		return i.Left.String() + " " + i.In.String()
	default:
		return i.Left.String() + " " + i.In.String() + " " + i.Right.String()
	}
}

func (l *Literal) SourceInfo() *parseinfo.Source { return l.SI }
func (p *Paren) SourceInfo() *parseinfo.Source   { return p.SI }
func (s *Special) SourceInfo() *parseinfo.Source { return s.SI }
func (l *Binding) SourceInfo() *parseinfo.Source { return l.SI }
func (p *Prefix) SourceInfo() *parseinfo.Source  { return p.Left.SourceInfo() }
func (i *Infix) SourceInfo() *parseinfo.Source   { return i.In.SourceInfo() }

func FindNextSpecialOrBinding(goAfterBindings bool, tree Tree, words ...string) (before, at, after Tree) {
	if tree == nil {
		return nil, nil, nil
	}

	switch tree := tree.(type) {
	case *Literal, *Paren:
		return tree, nil, nil

	case *Special:
		matches := false
		for _, s := range words {
			if tree.Kind == s {
				matches = true
				break
			}
		}
		if matches {
			return nil, tree, tree.After
		}
		afterBefore, afterAt, afterAfter := FindNextSpecialOrBinding(goAfterBindings, tree.After, words...)
		return &Special{
			SI:    tree.SI,
			Kind:  tree.Kind,
			After: afterBefore,
		}, afterAt, afterAfter

	case *Binding:
		matches := false
		for _, s := range words {
			if tree.Kind == s {
				matches = true
				break
			}
		}
		if matches {
			return nil, tree, tree.After
		}
		if !goAfterBindings {
			return tree, nil, nil
		}
		afterBefore, afterAt, afterAfter := FindNextSpecialOrBinding(goAfterBindings, tree.After, words...)
		return &Binding{
			SI:    tree.SI,
			Kind:  tree.Kind,
			Bound: tree.Bound,
			After: afterBefore,
		}, afterAt, afterAfter

	case *Prefix:
		// special can't be in the left
		rightBefore, rightAt, rightAfter := FindNextSpecialOrBinding(goAfterBindings, tree.Right, words...)
		if rightBefore == nil {
			return tree.Left, rightAt, rightAfter
		}
		return &Prefix{
			Left:  tree.Left,
			Right: rightBefore,
		}, rightAt, rightAfter

	case *Infix:
		// special can't be in the left or in
		rightBefore, rightAt, rightAfter := FindNextSpecialOrBinding(goAfterBindings, tree.Right, words...)
		return &Infix{
			Left:  tree.Left,
			In:    tree.In,
			Right: rightBefore,
		}, rightAt, rightAfter
	}

	panic("unreachable")
}

func Flatten(tree Tree) []Tree {
	var flat []Tree
	for t := range flatten(tree) {
		flat = append(flat, t)
	}
	return flat
}

func flatten(tree Tree) <-chan Tree {
	ch := make(chan Tree)
	go func() {
		flattenHelper(ch, tree)
		close(ch)
	}()
	return ch
}

func flattenHelper(ch chan<- Tree, tree Tree) {
	switch tree := tree.(type) {
	case *Literal, *Paren, *Special, *Binding:
		ch <- tree
	case *Prefix:
		flattenHelper(ch, tree.Left)
		flattenHelper(ch, tree.Right)
	case *Infix:
		ch <- &Infix{Left: nil, In: tree.In, Right: nil}
		ch <- tree.Left
		ch <- tree.Right
	}
}

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
		paren := &Paren{SI: tokens[0].SourceInfo, Kind: tokens[0].Value, Inside: inside}
		return paren, closing + 1, nil

	case "\\", "case":
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
				"nothing after binding",
			}
		}
		return &Binding{
			SI:    tokens[0].SourceInfo,
			Kind:  tokens[0].Value,
			Bound: bound,
			After: after,
		}, len(tokens), nil

	case ",", ";", ":", "|", "=", "record", "union", "alias", "func", "switch", "strict":
		after, err := MultiTree(tokens[1:])
		if err != nil {
			return nil, 0, err
		}
		return &Special{
			SI:    tokens[0].SourceInfo,
			Kind:  tokens[0].Value,
			After: after,
		}, len(tokens), nil

	default:
		switch LiteralKindOf(tokens[0].Value) {
		case LiteralIdentifier:
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
		}
		return &Literal{
			SI:    tokens[0].SourceInfo,
			Value: tokens[0].Value,
		}, 1, nil
	}
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

func hasLetterOrDigit(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return true
		}
	}
	return false
}

func findClosingParen(tokens []Token) (i int, ok bool) {
	round := 0  // ()
	square := 0 // []
	curly := 0  // {}
	for i := range tokens {
		switch tokens[i].Value {
		case "(":
			round++
		case ")":
			round--
		case "[":
			square++
		case "]":
			square--
		case "{":
			curly++
		case "}":
			curly--
		}
		if round < 0 || square < 0 || curly < 0 {
			return i, false
		}
		if round == 0 && square == 0 && curly == 0 {
			return i, true
		}
	}
	return len(tokens), false
}
