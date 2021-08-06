package main

import (
	"github.com/armory/armory-cli/cmd"
	"os"
)

func main() {
	if err := cmd.MainCommand().Execute(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
