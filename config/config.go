package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/koding/multiconfig"
	log "github.com/sirupsen/logrus"
)

const (
	timeFormat = "150405"
	toolName   = "kafka-dumper"
)

// Config stores service config parameters.
type Config struct {
	kafkaVersion       sarama.KafkaVersion
	KafkaBrokers       []string `required:"true"`
	Topics             []string `required:"true"` // (example: '{"Topic1", "Topic2"}'
	OutputDir          string   `default:"OUTPUT_DATA"`
	KafkaClientID      string   `default:"kafka-dumper"`
	KafkaGroupID       string   `default:"kafka-dumper"`
	KafkaVersionString string   `default:"0.10.2.0"`
	Timezone           string   `default:"GMT"`
	Log                string   `default:"Info"`
	LocalLog           bool     `required:"false"` // if true  - will write log to stdout and to
	// file kafka-dump.log at OutputDir

	Overwrite bool `required:"false"` // if true - will create unique consumerID and messages will be received again
	Newest    bool `required:"false"` // if set true - will start dump all messages that appears in
	// kafka after start of tool

	Init bool `required:"false"`
}

// Help output for flags when program run with -h flag.
func setFlagsHelp() map[string]string {
	usageMsg := make(map[string]string)

	usageMsg["Init"] = "When true - creates initial config at usr.HomeDir/.config/kafka-dumper"

	usageMsg["KafkaClientID"] = "Kafka consumer group clientID"
	usageMsg["KafkaGroupID"] = "Kafka Consumer group Name"
	usageMsg["KafkaBrokers"] = "Kafka brokers address"
	usageMsg["Log"] = `Log level: All, Debug, Info, Error, Fatal, Panic, Warn`
	usageMsg["KafkaVersionString"] = `Kafka version`
	usageMsg["OutputDir"] = "Location of directory where kafka dump will be stored locally"
	usageMsg["Overwrite"] = `When select as true - 
	all previous dump in specified OutputDir will be overwritten. All kafka messages would be read again`
	usageMsg["Timezone"] = "Timezone that will be used for timestamps in messages"
	usageMsg["Topics"] = `List of all topics with specified message type which will be dumped`
	usageMsg["Log"] = `Log level that will be displayed (DEBUG, INFO, ERROR, WARN, FATAL"`
	usageMsg["LocalLog"] = `When true will write log to stdout and to file kafka-dump.log at OutputDir`
	usageMsg["Newest"] = `when set true - will start dump all messages that appears in kafka after start of tool`

	return usageMsg
}

// GetTimeZone - parses timezone string and return time.Location representation.
func (c *Config) GetTimeZone() *time.Location {
	timezone, err := time.LoadLocation(c.Timezone)
	if err != nil {
		log.Fatalf("Failed to set Timezone")
	}

	log.Infof("Timezone: %s", timezone)

	return timezone
}

// Make each tool run unique - add hostname of runner to KafkaClientID for Kafka Consumer.
func (c *Config) addHostnameToClientID() {
	hn, err := os.Hostname()
	if err != nil {
		hn = time.Now().Format(timeFormat)
		log.Infof("Failed to get hostname, using %s (%+v)", hn, err)
	}

	c.KafkaClientID = c.KafkaClientID + "-" + hn
}

// Start reading kafka messages from the beginning and overwrite already received data.
func (c *Config) overwriteMessages() {
	if c.Overwrite {
		log.Infof("All received Messages will be overwritten")

		c.KafkaGroupID += "-" + time.Now().Format(timeFormat)
		c.KafkaClientID += "-" + time.Now().Format(timeFormat)

		if err := os.RemoveAll(path.Join(c.OutputDir)); err != nil {
			log.Fatalf("Failed to remove all dirs: %v", err)
		}
	}
}

// LoadConfig loads configuration for service to struct Config and store topics to Topics map.
func LoadConfig() *Config {
	svcConfig := &Config{}

	log.Infof("Loading configuration\n")

	usr, errUser := user.Current()
	if errUser != nil {
		log.Fatal(errUser)
	}

	log.Infof("Current Username: %s. Home dir: %s", usr.Username, usr.HomeDir)
	configPath := path.Join(usr.HomeDir, ".config/", toolName)

	m := newConfig(path.Join(configPath, "config.toml"), "KafkaDump", true)

	if err := m.Load(svcConfig); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if svcConfig.Init {
		initServiceConfigFile()
	}

	// set logger
	setLogger(svcConfig)

	// Parse topics

	svcConfig.addHostnameToClientID()

	svcConfig.overwriteMessages()

	svcConfig.setKafkaVersion()

	if err := m.Validate(svcConfig); err != nil {
		log.Fatalf("Config struct is invalid: %v\n", err)
	}

	log.Infof("Configuration loaded\n")

	prettyConfig, err := json.MarshalIndent(svcConfig, "", "")
	if err != nil {
		log.Fatalf("Failed to marshal indent config: %v", err)
	}

	log.Infof("Current config:\n %s", string(prettyConfig))

	return svcConfig
}

// KafkaVersion setter.
func (c *Config) setKafkaVersion() {
	// Parse kafkaVersion
	if v, err := sarama.ParseKafkaVersion(c.KafkaVersionString); err == nil {
		c.kafkaVersion = v
	} else {
		log.Fatalf("Failed to parse kafkaVersion: %v", err)
	}
}

// KafkaVersion getter.
func (c *Config) KafkaVersion() sarama.KafkaVersion {
	return c.kafkaVersion
}

// Implementation of default loader for multiconfig.
func newConfig(path string, prefix string, camelCase bool) *multiconfig.DefaultLoader {
	var loaders []multiconfig.Loader

	// Read default values defined via tag fields "default"
	loaders = append(loaders, &multiconfig.TagLoader{})

	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Warnf("Provided local config file [%s] does not exist. Flags and Environment variables will be used ", path)
		} else {
			log.Infof("Local config file [%s] will be used", path)
			// Choose what while is passed
			if strings.HasSuffix(path, "toml") {
				log.Debugf("Toml detected")
				loaders = append(loaders, &multiconfig.TOMLLoader{
					Path:   path,
					Reader: nil,
				})
			}
		}
	}

	e := &multiconfig.EnvironmentLoader{
		Prefix:    prefix,
		CamelCase: camelCase,
	}

	usageMsg := setFlagsHelp()
	f := &multiconfig.FlagLoader{
		Prefix:        "",
		Flatten:       false,
		CamelCase:     false,
		EnvPrefix:     prefix,
		ErrorHandling: 0,
		Args:          nil,
		FlagUsageFunc: func(s string) string { return usageMsg[s] },
	}

	loaders = append(loaders, e, f)
	loader := multiconfig.MultiLoader(loaders...)

	d := &multiconfig.DefaultLoader{}
	d.Loader = loader
	d.Validator = multiconfig.MultiValidator(&multiconfig.RequiredValidator{})

	return d
}

// creates initial config file.
func initServiceConfigFile() {
	log.Infof("Creating initial config file...")

	usr, errUser := user.Current()
	if errUser != nil {
		log.Fatal(errUser)
	}

	log.Infof("Current Username: %s. Home dir: %s", usr.Username, usr.HomeDir)

	configPath := path.Join(usr.HomeDir, ".config", toolName, "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), os.ModePerm); err != nil {
		log.Fatalf("failed creating all dirs for config file [%s]: %v", filepath.Dir(configPath), err)
	}

	configFile, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o777)
	if err != nil {
		log.Fatalf("error opening config file: %v", err)
	}

	OutputDir := path.Join(usr.HomeDir, "Desktop", "KAFKA-DUMP", "OUTPUT")

	_, writeErr := configFile.WriteString(fmt.Sprintf(`OutputDir="%s"
Topics=["Topic1", "Topic2"]
KafkaClientID="kafka-dumper"
Consumer_Group="test-kafka-dump"
KafkaVersion="0.10.2.0"
KafkaBrokers=["localhost:9092"]
Timezone="Europe/Brussels"
Overwrite=true
Log="Debug"
LocalLog=true
Newest=false`, OutputDir))
	if writeErr != nil {
		log.Fatalf("Failed to write config file: %v", writeErr)
	}

	if err := configFile.Close(); err != nil {
		log.Fatalf("Failed to close config file: %v", err)
	}

	log.Infof("Local initial config file was created at [%s]", configPath)
	os.Exit(1)
}

func setLogger(cfg *Config) {
	formatter := &log.TextFormatter{
		ForceColors:               true,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             true,
		TimestampFormat:           "2006-01-02 15:04:05",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	}

	log.SetFormatter(formatter)

	lvl, err := log.ParseLevel(cfg.Log)
	if err != nil {
		lvl = log.InfoLevel
	}

	log.SetLevel(lvl)

	if cfg.LocalLog {
		// Open logfile
		logFileLoc := path.Join(cfg.OutputDir, "kafka-dump.log")
		if err := os.MkdirAll(filepath.Dir(logFileLoc), 0o700); err != nil {
			log.Fatalf("failed creating all dirs for logfile [%s]:  %v", filepath.Dir(logFileLoc), err)
		}

		logFile, err := os.OpenFile(logFileLoc, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}

		// create multiwriter for logs
		mw := io.MultiWriter(os.Stdout, logFile)

		log.SetOutput(mw)
	}
}
