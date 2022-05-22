package stragts

import (
	"fmt"
	"reflect"
)

type Tag struct {
	Name  string
	Value string
}

func applyNode(f reflect.Value, n node) {
	if f.Kind() == reflect.Pointer && f.IsZero() {
		f.Set(reflect.New(f.Type().Elem()))
		f = f.Elem()
	}
	switch nv := n.(type) {
	case *boolNode:
		f.SetBool(nv.value)
	case *switchNode:
		f.SetBool(nv.value.value)
	case *identifierNode:
		f.SetString(nv.value)
	case *nilNode:
		f.Set(reflect.Zero(f.Type()))
	case *numberNode:
		switch {
		case nv.IsInt:
			f.SetInt(nv.Int64)
		case nv.IsUint:
			f.SetUint(nv.Uint64)
		case nv.IsFloat:
			f.SetFloat(nv.Float64)
		}
	case *sliceNode:
		f.Set(reflect.MakeSlice(f.Type(), len(nv.values), len(nv.values)))
		for i, el := range nv.values {
			applyNode(f.Index(i), el)
		}
	case *stringNode:
		f.SetString(nv.Text)
	}
}

func (tag Tag) Fill(model any) error {
	// Ensure we're working directly on a reference to a structure
	// value that is held by the call side and not a copy.
	m := reflect.ValueOf(model)
	if !m.IsValid() || m.IsNil() || m.Type().Kind() != reflect.Ptr || m.Type().Elem().Kind() != reflect.Struct {
		return fmt.Errorf("fill target must be a pointer to a struct, not %T", model)
	}
	m = m.Elem()

	// Skip parsing the tag value if it is just a dash.
	if tag.Value == "-" {
		return nil
	}

	// Processing of the tag flags, nothing special here.
	p, err := parseValue(tag.Value)
	if err != nil {
		return err
	}

	for i, n := range p.indexed {
		applyNode(m.Field(i), n)
	}

	for k, v := range p.keyword {
		f := m.FieldByNameFunc(func(s string) bool {
			return toKebabCase(s) == k
		})
		applyNode(f, v)
	}

	return nil
}

func Lookup(tags reflect.StructTag, key string) (t *Tag, ok bool) {
	value, ok := tags.Lookup(key)
	if ok {
		t = &Tag{Name: key, Value: value}
	}
	return
}
