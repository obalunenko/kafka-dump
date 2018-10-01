package main

import (
	"fmt"

	"github.com/oleg-balunenko/kafka-dump/config"
	"github.com/oleg-balunenko/kafka-dump/dumper"
)

var (
	version string
	build   string
	commit  string
)

func main() {

	fmt.Printf("Version info: %s:%s", version, build)
	fmt.Printf("commit: %s ", commit)

	service := config.LoadConfig()

	dumper.Start(service)

}
