package main

import (
	"gitlab.com/oleg.balunenko/kafka-dump/config"
	"gitlab.com/oleg.balunenko/kafka-dump/dumper"
)

func main() {

	service := config.LoadConfig()

	dumper.Start(service)

}
