package stragts

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var nodeStringer = map[nodeType]string{
	nodeList:       "list",
	nodeArgument:   "arg",
	nodeNil:        "nil",
	nodeBool:       "bool",
	nodeNumber:     "number",
	nodeString:     "string",
	nodeIdentifier: "ident",
	nodeSwitch:     "switch",
	nodeSlice:      "slice",
}

func graphNode(sb *strings.Builder, n node) {
	sb.WriteByte('<')
	sb.WriteString(nodeStringer[n.getType()])
	switch v := n.(type) {
	case *nilNode:
	case *switchNode:
		sb.WriteByte(' ')
		sb.WriteByte(map[bool]byte{false: '!', true: '~'}[v.value.value])
		graphNode(sb, v.ident)
	case *sliceNode:
		sb.WriteString(fmt.Sprintf("#%d ", len(v.values)))

		for _, c := range v.values {
			graphNode(sb, c)
		}
	default:
		sb.WriteByte(' ')
		sb.WriteString(n.String())
	}
	sb.WriteByte('>')
}

func graph(root *listNode) string {
	var sb strings.Builder

	for i, n := range root.nodes {
		sb.WriteByte('<')
		sb.WriteString(nodeStringer[n.getType()])

		sb.WriteString(" [")
		if n.ident != nil {
			sb.WriteByte(':')
			sb.WriteString(n.ident.String())
		} else {
			sb.WriteByte('#')
			sb.WriteString(strconv.Itoa(i))
		}
		sb.WriteString("]=")
		graphNode(&sb, n.value)
		sb.WriteByte('>')
	}

	return sb.String()
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		inp     string
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{inp: "foo", want: "<arg [#0]=<ident foo>>", wantErr: assert.NoError},
		{inp: "foo,baa", want: "<arg [#0]=<ident foo>><arg [#1]=<ident baa>>", wantErr: assert.NoError},
		{inp: "foo;baa",
			want: "<arg [#0]=<slice#2 <ident foo><ident baa>>>", wantErr: assert.NoError},

		{inp: "!foo", want: "<arg [:foo]=<switch !<ident foo>>>", wantErr: assert.NoError},
		{inp: "~foo", want: "<arg [:foo]=<switch ~<ident foo>>>", wantErr: assert.NoError},

		{inp: "nil", want: "<arg [#0]=<nil>>", wantErr: assert.NoError},
		{inp: "nil;nil;nil", want: "<arg [#0]=<slice#3 <nil><nil><nil>>>", wantErr: assert.NoError},
		{inp: "foo=nil", want: "<arg [:foo]=<nil>>", wantErr: assert.NoError},

		{inp: "true", want: "<arg [#0]=<bool true>>", wantErr: assert.NoError},
		{inp: "true;true;true", want: "<arg [#0]=<slice#3 <bool true><bool true><bool true>>>", wantErr: assert.NoError},
		{inp: "foo=true", want: "<arg [:foo]=<bool true>>", wantErr: assert.NoError},

		{inp: "false",
			want: "<arg [#0]=<bool false>>", wantErr: assert.NoError},
		{inp: "false;false;false",
			want:    "<arg [#0]=<slice#3 <bool false><bool false><bool false>>>",
			wantErr: assert.NoError},
		{inp: "foo=false",
			want:    "<arg [:foo]=<bool false>>",
			wantErr: assert.NoError},

		{inp: "'hello world'",
			want:    "<arg [#0]=<string 'hello world'>>",
			wantErr: assert.NoError},
		{inp: "'hello world';'foo bar'",
			want:    "<arg [#0]=<slice#2 <string 'hello world'><string 'foo bar'>>>",
			wantErr: assert.NoError},
		{inp: "foo='hello world'",
			want:    "<arg [:foo]=<string 'hello world'>>",
			wantErr: assert.NoError},
		{inp: "foo='hello world';'foo bar'",
			want:    "<arg [:foo]=<slice#2 <string 'hello world'><string 'foo bar'>>>",
			wantErr: assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.inp)
			if !tt.wantErr(t, err, fmt.Sprintf("Parse(%v)", tt.inp)) {
				return
			}
			assert.Equalf(t, tt.want, graph(got.root), "Parse(%v)", tt.inp)

			assert.Equalf(t, tt.inp, got.root.String(), "Parse(%v)", tt.inp)
		})
	}
}
