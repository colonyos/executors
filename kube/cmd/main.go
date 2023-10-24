package main

import (
	"github.com/colonyos/executors/kube/internal/cli"
	"github.com/colonyos/executors/kube/pkg/build"
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
