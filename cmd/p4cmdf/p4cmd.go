package main

import (
	"fmt"
	"os"

	"github.com/rcowham/go-libp4"
)

func main() {
	p4 := p4.NewP4("", "")
	if result, err := p4.RunP(os.Args[1:]); err != nil {
		fmt.Printf("Error: %v %v", err, result)
	} else {
		for _, r := range result {
			for k, v := range r {
				fmt.Printf("%v: %v\n", k, v)
			}
			fmt.Printf("\n")
		}
		fmt.Printf("\nResult: %v\n", result)
	}
}
