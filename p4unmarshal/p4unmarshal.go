package main

import (
	"fmt"
	"os"

	p4 "github.com/brettbates/p4go"
)

func main() {
	p4 := p4.NewP4()
	result, err := p4.Run(os.Args[1:])
	if err != nil {
		fmt.Printf("Error: %v %v\n", err, result)
	}
	for _, r := range result {
		for k, v := range r {
			fmt.Printf("%v: %v\n", k, v)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\nResult: %v\n", result)
}
