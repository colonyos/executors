package main

import (
	"github.com/colonyos/executors/echo/internal/cli"
	"github.com/colonyos/executors/echo/pkg/build"
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
