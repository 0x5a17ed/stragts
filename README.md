# stragts
[Go](https://go.dev) module providing a structured value parser for [reflect.StructTag](https://pkg.go.dev/reflect#StructTag). 


## 📦 Installation

```console
$ go get -u github.com/0x5a17ed/stragts@latest 
```


## 🤔 Usage

```go
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
```


## 🥇 Acknowledgments

The design and the implementation are roughly based on the idea and syntax of the lovely [github.com/muir/reflectutils](https://github.com/muir/reflectutils) module with an implementation based on the amazing [text/template/parse](https://github.com/golang/go/blob/0a1a092c4b56a1d4033372fbd07924dad8cbb50b/src/text/template/parse/). Both projects have been very inspirational.

The only reason for me to write my own module was because `reflectutils` was missing a few features, and I was always looking for an opportunity to write a lexer.
