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

	fmt.Printf("Version info: %s:%s\n", version, build)
	fmt.Printf("commit: %s \n", commit)

	svcCfg := config.LoadConfig()

	dumper.Start(svcCfg.KafkaBrokers, svcCfg.KafkaGroupID, svcCfg.KafkaClientID, svcCfg.KafkaVersion(), svcCfg.Newest, svcCfg.Topics, svcCfg.OutputDir)

}
