package main

import (
	"flag"
	"fmt"
)

func main() {
	compress := flag.Bool("c", false, "Compress the output files")
	outDir := flag.String("o", "out", "Output directory. WARNING: Existing content will be deleted")
	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		indx := args[0]
		dr := &DumpReader{Compress: *compress, OutDir: *outDir}
		err := dr.readDump(indx)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}
