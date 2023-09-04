package main

import (
	"github.com/colonyos/executors/hpc/internal/cli"
	"github.com/colonyos/executors/hpc/pkg/build"
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
