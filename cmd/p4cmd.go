package main

import (
	"fmt"
	"os"

	p4 "github.com/brettbates/go-libp4"
)

func main() {
	p4 := p4.NewP4Params("localhost:1999", "perforce", "gg")
	if results, err := p4.Run(os.Args[1:]); err != nil {
		fmt.Printf("Error: %v %v", err, results)
	} else {
		fmt.Printf("Results: %v", results)
	}
}
