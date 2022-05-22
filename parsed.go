package stragts

import (
	"errors"
)

var (
	ErrPositionalAfterKeyword = errors.New("positional argument after keywords")
)

type parsed struct {
	indexed []node
	keyword map[string]node
}

func parseValue(inp string) (*parsed, error) {
	t, err := Parse(inp)
	if err != nil {
		return nil, err
	}

	p := &parsed{keyword: map[string]node{}}
	for _, n := range t.root.nodes {
		if n.ident == nil {
			if len(p.keyword) != 0 {
				return nil, ErrPositionalAfterKeyword
			}
			p.indexed = append(p.indexed, n.value)
		} else {
			p.keyword[n.ident.String()] = n.value
		}
	}

	return p, nil
}
