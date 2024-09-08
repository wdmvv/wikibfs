package main

import (
	"fmt"
	"wikibfs/internal/config"
	"wikibfs/internal/wiki"
)

func main() {
	// extremely questionable
	// imagine i call this binary from god-knows-where, it might not find config relative to exec. path?
	// so for now launching from the same dir
	err := config.Load("config.json")
	if err != nil {
		panic(err)
	}

	depth, path, errs := wiki.Search("https://en.wikipedia.org/wiki", "/Apple", "/Tree")
	if len(errs) != 0 {
		for err := range errs {
			fmt.Println(err)
		}
	} else {
		fmt.Println(path, depth)
	}
}
