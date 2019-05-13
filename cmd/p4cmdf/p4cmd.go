package main

import (
	"fmt"
	"os"

	"github.com/rcowham/go-libp4"
)

func main() {
	p4 := p4.NewP4("", "")
	fmt.Printf("args: %v", os.Args[1:])
	if result, err := p4.RunP(os.Args[1:]); err != nil {
		fmt.Printf("Error: %v %v", err, result)
	} else {
		for _, v := range result {
			fmt.Printf("map map\n")
			for k, j := range v {
				fmt.Printf("%v: %v\n", k, j)
			}
			// switch r := v.(type) {
			// default:
			// 	fmt.Printf("unexpected type %T\n", r) // %T prints whatever type t has
			// case map[interface{}]interface{}:
			// 	fmt.Printf("map map\n")
			// 	for k, v := range r {
			// 		fmt.Printf("%v: %v\n", k, v)
			// 	}
			// case bool:
			// 	fmt.Printf("boolean %t\n", r) // t has type bool
			// case int:
			// 	fmt.Printf("integer %d\n", r) // t has type int
			// case *bool:
			// 	fmt.Printf("pointer to boolean %t\n", *r) // t has type *bool
			// case *int:
			// 	fmt.Printf("pointer to integer %d\n", *r) // t has type *int
			// }
			fmt.Printf("Results: %v", v)
		}

	}
}
