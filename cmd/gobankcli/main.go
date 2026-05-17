package main

import (
	"context"
	"fmt"
	"os"

	"gobankcli/internal/cmd"
)

var version = "dev"

func main() {
	if err := cmd.Run(context.Background(), os.Args[1:], version, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
