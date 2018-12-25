package dumper

import (
	"encoding/json"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	log "github.com/sirupsen/logrus"
)

// Start - starts the dumper consumer loop and processing messages
func Start(kafkaBrokers []string, kafkaGroupID string, kafkaClientID string,
	kafkaVersion sarama.KafkaVersion, kafkaNewestOffset bool, kafkaTopics []string, outputDir string) {

	// Create Kafka consumers
	kafkaConfig := cluster.NewConfig()

	kafkaConfig.Group.Return.Notifications = true

	kafkaConfig.ClientID = kafkaClientID

	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Version = kafkaVersion
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	if kafkaNewestOffset {
		log.Infof("Will use OffsetNewest")
		kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	}

	consumer, err := cluster.NewConsumer(kafkaBrokers, kafkaGroupID, kafkaTopics, kafkaConfig)

	if err != nil {
		log.Fatalf("Kafka connection failed. Err: %v", err)
	}

	log.Infof("consumer started\n")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, syscall.SIGTERM)

	n := consumerLoop(consumer, outputDir, signals)
	log.Infof("Total messages processed: %d", n)

}

func consumerLoop(consumer *cluster.Consumer, outputDir string, signals chan os.Signal) (msgCount uint32) {

	// Get signal for finish

	for {
		log.Infof("Consumer loop started\n")
		select {
		case errConsumer := <-consumer.Errors():
			atomic.AddUint32(&msgCount, 1)
			log.Errorf("consumer error: %v", errConsumer)

		case msg := <-consumer.Messages():
			atomic.AddUint32(&msgCount, 1)

			if msg != nil {

				log.Infof("received message from topic [%s]:[part[%d];offset[%d];key[%s]]",
					msg.Topic, msg.Partition, msg.Offset, msg.Key)
				log.Debugf("Total amount of received messages: %d", atomic.LoadUint32(&msgCount))
				if err := dumpMessage(outputDir, msg); err != nil {
					log.Fatalf("Failed to dump message: %v", err)
				}
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
			atomic.AddUint32(&msgCount, 1)
			log.Errorf("Received consumerError: %v ", consumerError)

		case <-signals:
			log.Infof("Got UNIX signal, shutting down")
			return msgCount
		}
	}

}
