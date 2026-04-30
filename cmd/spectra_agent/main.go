package main

import (
	"os"
)

func main() {
	cmd := NewRootCommand()
	os.Exit(cmd.Execute())
}
