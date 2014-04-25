package main

import (
	"flag"
	"os"

	"github.com/zond/docile"
)

func main() {
	dst := flag.String("dst", "", "Where to put the source code version of comments")
	pack := flag.String("pack", "", "Go package to generate comment source for")

	flag.Parse()

	if *dst == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *pack == "" {
		flag.Usage()
		os.Exit(2)
	}

	if err := docile.Generate(*pack, *dst); err != nil {
		panic(err)
	}
}
