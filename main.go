package main

import (
	"flag"

	"github.com/stackexchange/haproxy-kubefigurator/cmd"
)

func main() {
	flag.CommandLine.Parse([]string{}) // quiets down kube client library logging
	cmd.Execute()
}
