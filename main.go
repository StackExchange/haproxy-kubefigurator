package main

import (
	"flag"

	"github.com/StackExchange/haproxy-kubefigurator/cmd"
)

func main() {
	// quiets down kube client library logging
	flag.CommandLine.Parse([]string{})
	cmd.Execute()
}
