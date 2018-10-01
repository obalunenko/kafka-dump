package dumper

import (
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/oleg-balunenko/kafka-dump/config"
	log "github.com/sirupsen/logrus"
)

// Start - starts the dumper consumer loop and processing messages
func Start(cfg *config.Config) {

	// Create Kafka consumers
	kafkaConfig := cluster.NewConfig()

	kafkaConfig.Group.Return.Notifications = true

	kafkaConfig.ClientID = cfg.KafkaClientID

	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Version = cfg.KafkaVersion()
	if cfg.Newest {
		log.Infof("Will use OffsetNewest")
		kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	} else {
		log.Infof("Will use OffsetOldest")
		kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	}

	consumer, err := cluster.NewConsumer(cfg.KafkaBrokers, cfg.KafkaGroupID, cfg.Topics, kafkaConfig)

	if err != nil {
		log.Fatalf("Kafka connection failed. Err: %v", err)
	}

	log.Infof("consumer started\n")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, syscall.SIGTERM)

	// Count how many message processed
	msgCount := 0
	mu := &sync.Mutex{}

	// Get signal for finish

	for {
		log.Infof("Consumer loop started\n")
		select {
		case errConsumer := <-consumer.Errors():
			mu.Lock()
			msgCount++
			mu.Unlock()
			log.Errorf("consumer error: %v", errConsumer)

		case msg := <-consumer.Messages():
			if msg != nil {

				log.Infof("received message from topic [%s]:[part[%d];offset[%d];key[%s]]", msg.Topic, msg.Partition, msg.Offset, msg.Key)
				mu.Lock()
				msgCount++
				log.Debugf("Total amount of received messages: %d", msgCount)
				mu.Unlock()
				if err := dumpMessage(cfg, msg); err != nil {
					log.Fatalf("Failed to dump message: %v", err)
				}
			} else {
				msgCount++
				log.Warnf("Nil message received: %v", msg)
			}
			//tell kafka we are done with this message
			consumer.MarkOffset(msg, "")

		case rbe := <-consumer.Notifications():
			//Rebalancing
			js, err := json.Marshal(rbe)
			if err != nil {
				log.Errorf("Error when Marshal json from  Notifications channel: %v", err)
			}
			log.Infof("Rebalancing: %s", string(js))

		case consumerError := <-consumer.Errors():
			msgCount++
			log.Errorf("Received consumerError: %v ", consumerError)

		case <-signals:
			log.Infof("Got UNIX signal, shutting down")
			log.Infof("Total messages processed: %d", msgCount)
			return
		}
	}

}
