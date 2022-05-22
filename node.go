package stragts

import (
	"fmt"
	"strconv"
	"strings"
)

// pos represents a byte position in the original input text from which
// this template was parsed.
type pos int

func (p pos) getPosition() pos {
	return p
}

type node interface {
	getType() nodeType

	getPosition() pos

	String() string

	writeTo(*strings.Builder)
}

// nodeType identifies the type of the parse tree node.
type nodeType int

func (t nodeType) getType() nodeType { return t }

const (
	_              nodeType = iota // Plain text.
	nodeList                       // A list of nodes.
	nodeArgument                   // An argument.
	nodeNil                        // An untyped nil value.
	nodeBool                       // A boolean value.
	nodeNumber                     // A numerical value.
	nodeString                     // A string value.
	nodeIdentifier                 // An identifier value.
	nodeSwitch                     // An disable field value.
	nodeSlice                      // A slice value.
)

type baseNode struct {
	nodeType
	pos
}

func newBaseNode(nodeType nodeType, pos pos) baseNode {
	return baseNode{nodeType: nodeType, pos: pos}
}

// listNode holds a sequence of nodes.
type listNode struct {
	baseNode
	nodes []*argumentNode
}

func (n *listNode) String() string {
	var sb strings.Builder
	n.writeTo(&sb)
	return sb.String()
}

func (n *listNode) writeTo(sb *strings.Builder) {
	if len(n.nodes) > 0 {
		n.nodes[0].writeTo(sb)
		for _, c := range n.nodes[1:] {
			sb.WriteByte(',')
			c.writeTo(sb)
		}
	}
}

func (n *listNode) append(a *argumentNode) {
	n.nodes = append(n.nodes, a)
}

func (t *tree) newList(pos pos) *listNode {
	return &listNode{baseNode: newBaseNode(nodeList, pos)}
}

// argumentNode holds a single argument.
type argumentNode struct {
	baseNode
	ident *identifierNode
	value node
}

func (n *argumentNode) String() string {
	var sb strings.Builder
	n.writeTo(&sb)
	return sb.String()
}

func (n *argumentNode) writeTo(sb *strings.Builder) {
	if n.ident != nil && n.value.getType() != nodeSwitch {
		sb.WriteString(n.ident.String())
		sb.WriteByte('=')
	}
	sb.WriteString(n.value.String())
}

func (t *tree) newArgument(pos pos, ident *identifierNode, value node) *argumentNode {
	return &argumentNode{baseNode: newBaseNode(nodeArgument, pos), ident: ident, value: value}
}

// identifierNode holds an identifier.
type identifierNode struct {
	baseNode
	value string
}

func (n *identifierNode) String() string              { return n.value }
func (n *identifierNode) writeTo(sb *strings.Builder) { sb.WriteString(n.String()) }

func (t *tree) newIdentifier(pos pos, ident string) *identifierNode {
	return &identifierNode{baseNode: newBaseNode(nodeIdentifier, pos), value: ident}
}

// nilNode holds the special identifier 'nil' representing an untyped nil constant.
type nilNode struct{ baseNode }

func (n *nilNode) String() string              { return "nil" }
func (n *nilNode) writeTo(sb *strings.Builder) { sb.WriteString(n.String()) }

func (t *tree) newNil(pos pos) *nilNode {
	return &nilNode{baseNode: newBaseNode(nodeNil, pos)}
}

// boolNode holds a boolean constant.
type boolNode struct {
	baseNode
	value bool // value of the boolean constant.
}

func (n *boolNode) String() string {
	if n.value {
		return "true"
	}
	return "false"
}

func (n *boolNode) writeTo(sb *strings.Builder) {
	sb.WriteString(n.String())
}

func (t *tree) newBool(pos pos, value bool) *boolNode {
	return &boolNode{baseNode: newBaseNode(nodeBool, pos), value: value}
}

// stringNode holds a string constant. The value has been "unquoted".
type stringNode struct {
	baseNode
	Quoted string // The original text of the string, with quotes.
	Text   string // The string, after quote processing.
}

func (n *stringNode) String() string              { return n.Quoted }
func (n *stringNode) writeTo(sb *strings.Builder) { sb.WriteString(n.String()) }

func (t *tree) newString(pos pos, orig, text string) *stringNode {
	return &stringNode{baseNode: newBaseNode(nodeString, pos), Quoted: orig, Text: text}
}

// sliceNode holds a slice value.
type sliceNode struct {
	baseNode
	values []node
}

func (n *sliceNode) String() string {
	var sb strings.Builder
	n.writeTo(&sb)
	return sb.String()
}

func (n *sliceNode) writeTo(sb *strings.Builder) {
	if len(n.values) > 0 {
		n.values[0].writeTo(sb)
		for _, c := range n.values[1:] {
			sb.WriteByte(';')
			c.writeTo(sb)
		}
	}
}

func (t *tree) newSlice(pos pos, values []node) *sliceNode {
	return &sliceNode{baseNode: newBaseNode(nodeSlice, pos), values: values}
}

// switchNode holds an enabling or disabling field value.
type switchNode struct {
	baseNode
	ident *identifierNode
	value *boolNode
}

func (n *switchNode) String() string {
	if n.value.value {
		return "~" + n.ident.String()
	}
	return "!" + n.ident.String()
}

func (n *switchNode) writeTo(sb *strings.Builder) {
	sb.WriteString(n.String())
}

func (t *tree) newSwitch(pos pos, ident *identifierNode, value bool) *switchNode {
	return &switchNode{baseNode: newBaseNode(nodeSwitch, pos), ident: ident, value: t.newBool(pos, value)}
}

// numberNode holds a number: signed or unsigned integer, float, or complex.
// The value is parsed and stored under all the types that can represent the value.
// This simulates in a small amount of code the behavior of Go's ideal constants.
type numberNode struct {
	baseNode
	IsInt   bool    // Number has an integral value.
	IsUint  bool    // Number has an unsigned integral value.
	IsFloat bool    // Number has a floating-point value.
	Int64   int64   // The signed integer value.
	Uint64  uint64  // The unsigned integer value.
	Float64 float64 // The floating-point value.
	Text    string  // The original textual representation from the input.
}

func (n *numberNode) String() string              { return n.Text }
func (n *numberNode) writeTo(sb *strings.Builder) { sb.WriteString(n.String()) }

func (t *tree) newNumber(pos pos, text string) (*numberNode, error) {
	n := &numberNode{baseNode: newBaseNode(nodeNumber, pos), Text: text}

	// Do integer test first so we get 0x123 etc.
	u, err := strconv.ParseUint(text, 0, 64) // will fail for -0; fixed below.
	if err == nil {
		n.IsUint = true
		n.Uint64 = u
	}
	i, err := strconv.ParseInt(text, 0, 64)
	if err == nil {
		n.IsInt = true
		n.Int64 = i
		if i == 0 {
			n.IsUint = true // in case of -0.
			n.Uint64 = u
		}
	}
	// If an integer extraction succeeded, promote the float.
	if n.IsInt {
		n.IsFloat = true
		n.Float64 = float64(n.Int64)
	} else if n.IsUint {
		n.IsFloat = true
		n.Float64 = float64(n.Uint64)
	} else {
		f, err := strconv.ParseFloat(text, 64)
		if err == nil {
			// If we parsed it as a float but it looks like an integer,
			// it's a huge number too large to fit in an int. Reject it.
			if !strings.ContainsAny(text, ".eEpP") {
				return nil, fmt.Errorf("integer overflow: %q", text)
			}
			n.IsFloat = true
			n.Float64 = f
			// If a floating-point extraction succeeded, extract the int if needed.
			if !n.IsInt && float64(int64(f)) == f {
				n.IsInt = true
				n.Int64 = int64(f)
			}
			if !n.IsUint && float64(uint64(f)) == f {
				n.IsUint = true
				n.Uint64 = uint64(f)
			}
		}
	}
	if !n.IsInt && !n.IsUint && !n.IsFloat {
		return nil, fmt.Errorf("illegal number syntax: %q", text)
	}
	return n, nil
}
