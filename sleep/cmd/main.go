package main

import (
	"github.com/colonyos/executors/sleep/internal/cli"
	"github.com/colonyos/executors/sleep/pkg/build"
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
