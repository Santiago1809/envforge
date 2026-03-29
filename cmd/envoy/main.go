package main

import (
	"fmt"
	"os"

	"github.com/Santiago1809/envoy/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
