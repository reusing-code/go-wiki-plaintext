package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		indx := args[0]
		err := readDump(indx)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
