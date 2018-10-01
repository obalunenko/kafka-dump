package main

import (
	"github.com/oleg-balunenko/kafka-dump/config"
	"github.com/oleg-balunenko/kafka-dump/dumper"
)

func main() {

	service := config.LoadConfig()

	dumper.Start(service)

}
