package main

import (
	"fmt"

	"github.com/obalunenko/kafka-dump/config"
	"github.com/obalunenko/kafka-dump/dumper"
)

var (
	version string
	date    string
	commit  string
)

func main() {
	fmt.Printf("Version info: %s:%s\n", version, date)
	fmt.Printf("commit: %s \n", commit)

	svcCfg := config.LoadConfig()

	dumper.Start(svcCfg.KafkaBrokers, svcCfg.KafkaGroupID, svcCfg.KafkaClientID,
		svcCfg.KafkaVersion(), svcCfg.Newest, svcCfg.Topics, svcCfg.OutputDir)
}
