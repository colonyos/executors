package main

import (
	"github.com/colonyos/executors/k8s/internal/cli"
	"github.com/colonyos/executors/k8s/pkg/build"
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
