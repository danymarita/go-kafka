package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	_ "github.com/segmentio/kafka-go/snappy"
	"github.com/spf13/viper"
)

var (
	// kafka
	kafkaBrokerUrl     string
	kafkaVerbose       bool
	kafkaTopic         string
	kafkaConsumerGroup string
	kafkaClientId      string
)

func readViperConfig() *viper.Viper {
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("./params")
	v.AddConfigPath("/opt/go-kafka/params")
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	v.AddConfigPath(fmt.Sprintf("%s/../params", basepath))
	v.SetConfigName("gokafka")

	err := v.ReadInConfig()
	if err == nil {
		fmt.Printf("Using config file: %s \n\n", v.ConfigFileUsed())
	} else {
		panic(fmt.Errorf("Config error: %s", err))
	}

	return v
}

func main() {
	cfg := readViperConfig()
	// Kafka
	kafkaBrokerUrl = cfg.GetString("kafka.broker_url")
	kafkaVerbose = cfg.GetBool("kafka.verbose")
	kafkaClientId = cfg.GetString("kafka.client_id")
	kafkaConsumerGroup = cfg.GetString("kafka.consumer_group")
	kafkaTopic = cfg.GetString("kafka.topic")

	log.Info().Msg(fmt.Sprintf("Running Kafka worker. ConsumerGroup %s Topic %s...", kafkaConsumerGroup, kafkaTopic))

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	brokers := strings.Split(kafkaBrokerUrl, ",")

	config := kafka.ReaderConfig{
		Brokers:         brokers,
		GroupID:         kafkaConsumerGroup,
		Topic:           kafkaTopic,
		MinBytes:        10e3,            // 10KB
		MaxBytes:        10e6,            // 10MB
		MaxWait:         1 * time.Second, // Maximum amount of time to wait for new data to come when fetching batches of messages from kafka.
		ReadLagInterval: -1,
	}

	reader := kafka.NewReader(config)
	defer reader.Close()

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Error().Msgf("error while receiving message: %s", err.Error())
			continue
		}

		if err != nil {
			log.Error().Msgf("error while receiving message: %s", err.Error())
			continue
		}

		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s\n", m.Topic, m.Partition, m.Offset, string(m.Value))
	}
}
