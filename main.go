package main

import (
	"fmt"
	"os"

	"github.com/orangekame3/oqx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cmd.ExitCode(err))
	}
}
