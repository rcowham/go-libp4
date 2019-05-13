package main

import (
	"fmt"
	"os"

	"github.com/rcowham/go-libp4"
)

func main() {
	p4 := p4.NewP4("", "")
	if results, err := p4.Run(os.Args[1:]); err != nil {
		fmt.Printf("Error: %v %s", err, string(results))
	} else {
		fmt.Printf("Results: %v", string(results))
	}
}
