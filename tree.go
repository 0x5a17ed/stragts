package stragts

import (
	"fmt"
	"strconv"
)

// tree is the representation of a single parsed tag.
type tree struct {
	root *listNode // top-level root of the tree.

	// Parsing only; cleared after parse.
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
}

// undo backs the input stream up one token.
func (t *tree) undo() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *tree) backup2(t1 item) {
	t.token[1] = t1
	t.peekCount = 2
}

// next returns the next token.
func (t *tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.item()
	}
	return t.token[t.peekCount]
}

func (t *tree) peek() item {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.item()
	return t.token[0]
}

// errorf formats the error and terminates processing.
func (t *tree) errorf(format string, args ...any) {
	t.root = nil
	panic(fmt.Errorf(format, args...))
}

func (t *tree) error(err error) {
	t.errorf("%s", err)
}

// unexpected complains about the token and terminates processing.
func (t *tree) unexpected(token item) {
	t.errorf("unexpected %s", token)
}

func (t *tree) startParse(text string) (tree *tree, err error) {
	t.lex = lex(text)
	t.parse()
	return t, nil
}

func (t *tree) parse() {
	t.root = t.newList(t.peek().pos)
Loop:
	for t.peek().typ != itemEOF {
		t.root.append(t.argument())

		token := t.next()
		switch token.typ {
		case itemArgumentSeparator:
			continue Loop
		case itemEOF:
			break Loop
		}
	}
}

// Argument:
//
//	("!"|"~") identifier
//	(identifier "=")? simpleValue
//
func (t *tree) argument() *argumentNode {
	if pt := t.peek().typ; pt == itemDisable || pt == itemEnable {
		return t.switchArgument()
	}

	if t.peek().typ == itemIdentifier {
		token := t.next()
		separatorToken := t.peek()
		t.backup2(token)

		if separatorToken.typ == itemAssign {
			ident := t.identifier()
			t.next()
			return t.newArgument(ident.pos, ident, t.argumentValue())
		}
	}

	argument := t.argumentValue()
	return t.newArgument(argument.getPosition(), nil, argument)
}

func (t *tree) argumentValue() node {
	prev := t.next()
	nextToken := t.peek()
	t.backup2(prev)
	if nextToken.typ == itemListSeparator {
		return t.slice()
	}
	return t.simpleValue()
}

func (t *tree) simpleValue() node {
	switch t.peek().typ {
	case itemNil:
		return t.nil()
	case itemBool:
		return t.bool()
	case itemIdentifier:
		return t.identifier()
	case itemString:
		return t.string()
	case itemNumber:
		return t.number()
	default:
		t.unexpected(t.next())
		return nil
	}
}

func (t *tree) nil() *nilNode {
	return t.newNil(t.next().pos)
}

func (t *tree) bool() node {
	token := t.next()
	if token.val == "true" {
		return t.newBool(token.pos, true)
	}
	return t.newBool(token.pos, false)
}

func (t *tree) identifier() *identifierNode {
	token := t.next()
	return t.newIdentifier(token.pos, token.val)
}

func (t *tree) string() *stringNode {
	token := t.next()
	s, err := strconv.Unquote("\"" + token.val[1:len(token.val)-1] + "\"")
	if err != nil {
		t.error(err)
	}

	return t.newString(token.pos, token.val, s)
}

func (t *tree) number() node {
	token := t.next()
	number, err := t.newNumber(token.pos, token.val)
	if err != nil {
		t.error(err)
	}
	return number
}

func (t *tree) switchArgument() *argumentNode {
	sn := t.switchNode()
	return t.newArgument(sn.pos, sn.ident, sn)
}

func (t *tree) switchNode() *switchNode {
	var value bool

	prefix := t.next()
	switch prefix.typ {
	case itemDisable:
		value = false
	case itemEnable:
		value = true
	default:
		t.unexpected(prefix)
	}

	if t.peek().typ != itemIdentifier {
		t.unexpected(t.next())
	}
	return t.newSwitch(prefix.pos, t.identifier(), value)
}

func (t *tree) slice() node {
	items := []node{t.simpleValue()}
	for t.next().typ == itemListSeparator {
		items = append(items, t.simpleValue())
	}
	t.undo()
	return t.newSlice(items[0].getPosition(), items)
}

func newTree() *tree                  { return &tree{} }
func Parse(inp string) (*tree, error) { return newTree().startParse(inp) }
