package main

import (
	"fmt"
	"wikibfs/cmd"
	"wikibfs/internal/vault"
	"wikibfs/internal/wiki"
)

func main() {
	// extremely questionable
	// imagine i call this binary from god-knows-where, it might not find config relative to exec. path?
	// so for now launching from the same dir
	err := vault.Load("config.json")
	if err != nil {
		panic(err)
	}

	cmd.ArgParse()

	base := "https://en.wikipedia.org"
	depth, path, errs := wiki.Search(base, vault.Config.Cmd.Start, vault.Config.Cmd.End)
	if len(errs) != 0 {
		for err := range errs {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("The path from %s to %s is %d articles long:\n", vault.Config.Cmd.Start, vault.Config.Cmd.End, depth)
		for _, i := range path {
			fmt.Println("\t", base+i)
		}
	}
}
