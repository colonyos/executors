package main

import (
	"github.com/colonyos/executors/backup/internal/cli"
	"github.com/colonyos/executors/backup/pkg/build"
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
