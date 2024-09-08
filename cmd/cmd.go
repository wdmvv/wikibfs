package cmd

import (
	"flag"
	"wikibfs/internal/vault"
)

func ArgParse() {
	start := flag.String("s", "", "start of the search")
	end := flag.String("e", "", "end of the search")

	flag.Parse()
	// TODO: a lot of validation
	vault.Config.Cmd.Start = *start
	vault.Config.Cmd.End = *end
}
