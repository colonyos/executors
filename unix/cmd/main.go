package main

import (
	"github.com/colonyos/executors/unix/internal/cli"
	"github.com/colonyos/executors/unix/pkg/build"
)

var (
	BuildVersion string = ""
	BuildTime    string = ""
)

func main() {
	build.BuildVersion = BuildVersion
	build.BuildTime = BuildTime
	cli.Execute()
}
