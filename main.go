package main

import (
	"os"
	"runtime"
)

var buildVersion = "dev"

func main() {
	os.Exit(run(os.Args, runtime.GOOS))
}
