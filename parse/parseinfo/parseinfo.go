package parseinfo

import "fmt"

type Source struct {
	Filename     string
	Line, Column int
}

func (s *Source) String() string {
	if s == nil {
		return "<unknown source>"
	}
	return fmt.Sprintf("%s:%d:%d", s.Filename, s.Line, s.Column)
}
