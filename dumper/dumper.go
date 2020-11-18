package dumper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

func dumpMessage(outputDir string, msg *sarama.ConsumerMessage) error {
	// filename to use
	filename := generateFileName(msg)

	err := writeLineToFile(outputDir, msg.Value, filename, msg.Topic, msg.Partition)
	if err != nil {
		log.Errorf("Failed writing file for offset %v. Err: %v", msg.Offset, err)

		return err
	}

	return nil
}

func generateFileName(msg *sarama.ConsumerMessage) string {
	// timezone := svc.GetTimeZone()
	log.Debugf("Timestamp: %s", msg.BlockTimestamp)

	return fmt.Sprintf("%s_Partition_%d.txt", msg.Timestamp.Format("2006-01-02"), msg.Partition)
}

func writeLineToFile(outputDir string, line []byte, filename string, topic string, partition int32) error {
	fileLocation := filepath.Join(outputDir, topic, fmt.Sprintf("partition-%d", partition), filename)
	// create necessary dirs
	if err := os.MkdirAll(filepath.Dir(fileLocation), 0o700); err != nil {
		log.Errorf("failed creating all dirs at %v for offset %v", filepath.Dir(fileLocation), err)

		return fmt.Errorf("failed create dir: %w", err)
	}

	var f *os.File

	if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
		// path/to/whatever does not exist.
		log.Infof("Will be create new file: %s", fileLocation)

		f, err = os.OpenFile(fileLocation, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Infof("Will be used existed file: %s", fileLocation)

		f, err = os.OpenFile(fileLocation, os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	if _, err := f.Write(line); err != nil {
		panic(err)
	}

	return nil
}
