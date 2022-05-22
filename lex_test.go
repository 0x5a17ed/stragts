package stragts

import (
	"testing"

	assertpkg "github.com/stretchr/testify/assert"
)

var (
	tEOF               = mkItem(itemEOF, "")
	tNil               = mkItem(itemNil, "nil")
	tTrue              = mkItem(itemBool, "true")
	tFalse             = mkItem(itemBool, "false")
	tDisable           = mkItem(itemDisable, "!")
	tEnable            = mkItem(itemEnable, "~")
	tAssign            = mkItem(itemAssign, "=")
	tSliceSeparator    = mkItem(itemListSeparator, ";")
	tArgumentSeparator = mkItem(itemArgumentSeparator, ",")
)

func mkItem(typ itemType, text string) item {
	return item{typ: typ, val: text}
}

func tIdentifier(text string) item { return mkItem(itemIdentifier, text) }
func tString(text string) item     { return mkItem(itemString, text) }
func tNumber(text string) item     { return mkItem(itemNumber, text) }

func equal(t *testing.T, i1, i2 []item) bool {
	assert := assertpkg.New(t)

	if !assert.Len(i2, len(i1)) {
		return false
	}
	for k := range i1 {
		if !assert.Equal(i1[k].typ, i2[k].typ, "%v != %v", i1, i2) {
			return false
		}
		if !assert.Equal(i1[k].val, i2[k].val, "%v != %v", i1, i2) {
			return false
		}
	}
	return true
}

func collect(inp string) (items []item) {
	for l := lex(inp); ; {
		item := l.item()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}

func Test_lex(t *testing.T) {
	tests := []struct {
		name string
		inp  string
		want []item
	}{
		{"emtpy", "", []item{tEOF}},

		{"identifier", "field", []item{tIdentifier("field"), tEOF}},
		{"true simpleValue", "true", []item{tTrue, tEOF}},
		{"false simpleValue", "false", []item{tFalse, tEOF}},
		{"nil simpleValue", "nil", []item{tNil, tEOF}},

		{"numeric simpleValue", "1234", []item{tNumber("1234"), tEOF}},
		{"string simpleValue", "'012foo'", []item{tString("'012foo'"), tEOF}},

		{"disabled field", "!foo", []item{tDisable, tIdentifier("foo"), tEOF}},
		{"enabled field", "~foo", []item{tEnable, tIdentifier("foo"), tEOF}},

		{"slice simpleValue", "foo;baa", []item{tIdentifier("foo"), tSliceSeparator, tIdentifier("baa"), tEOF}},
		{"nil slice", "nil;nil", []item{tNil, tSliceSeparator, tNil, tEOF}},

		{"assign bool false", "foo=nil", []item{
			tIdentifier("foo"), tAssign, tNil, tEOF,
		}},
		{"assign bool false", "foo=false", []item{
			tIdentifier("foo"), tAssign, tFalse, tEOF,
		}},
		{"assign bool true", "foo=true", []item{
			tIdentifier("foo"), tAssign, tTrue, tEOF,
		}},
		{"assign numeric", "foo=1234", []item{
			tIdentifier("foo"), tAssign, tNumber("1234"), tEOF,
		}},
		{"assign identifier", "foo=baa", []item{
			tIdentifier("foo"), tAssign, tIdentifier("baa"), tEOF,
		}},
		{"assign quote", "foo='012baa'", []item{
			tIdentifier("foo"), tAssign, tString("'012baa'"), tEOF,
		}},

		{"multiple#01", "one,two", []item{
			tIdentifier("one"), tArgumentSeparator, tIdentifier("two"), tEOF,
		}},
		{"multiple#01", "one,two='three',foo=true", []item{
			tIdentifier("one"), tArgumentSeparator, tIdentifier("two"), tAssign, tString("'three'"), tArgumentSeparator, tIdentifier("foo"), tAssign, tTrue, tEOF,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equal(t, tt.want, collect(tt.inp))
		})
	}
}
