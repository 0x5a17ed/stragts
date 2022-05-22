// go:build examples

package main

import (
	"fmt"
	"reflect"

	"github.com/0x5a17ed/stragts"
)

type TagStruct struct {
	Index    *string
	Priority *int
}

func HandleStruct(data any) {
	v := reflect.ValueOf(data)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		tag, found := stragts.Lookup(t.Field(i).Tag, "norm")
		if !found {
			continue
		}

		var tagStruct TagStruct
		if err := tag.Fill(&tagStruct); err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		fmt.Println("index:", *tagStruct.Index, "priority:", *tagStruct.Priority)
	}
}

func main() {
	type TaggedStruct struct {
		Name   string `norm:"index=idx_member,priority=2"`
		Number string `norm:"index=idx_member,priority=1"`
	}
	HandleStruct(TaggedStruct{})
}
