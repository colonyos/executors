package main

import (
	"github.com/colonyos/executors/ml/internal/cli"
	"github.com/colonyos/executors/ml/pkg/build"
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
