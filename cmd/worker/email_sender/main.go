package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/danymarita/go-kafka/dep/app"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	_ "github.com/segmentio/kafka-go/snappy"
)

var (
	// kafka
	kafkaBrokerUrl     string
	kafkaVerbose       bool
	orderTopic         string
	sendEmailTopic     string
	kafkaConsumerGroup string
	kafkaClientId      string
)

func main() {
	cfg := app.ReadViperConfig()
	// Kafka
	kafkaBrokerUrl = cfg.GetString("kafka.broker_url")
	kafkaVerbose = cfg.GetBool("kafka.verbose")
	kafkaClientId = cfg.GetString("kafka.client_id")
	kafkaConsumerGroup = cfg.GetString("kafka.consumer_group")
	sendEmailTopic = cfg.GetString("topics.send_email")

	log.Info().Msg(fmt.Sprintf("Running Kafka worker. ConsumerGroup %s Topic %s...", kafkaConsumerGroup, sendEmailTopic))

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	brokers := strings.Split(kafkaBrokerUrl, ",")

	config := kafka.ReaderConfig{
		Brokers:         brokers,
		GroupID:         kafkaConsumerGroup,
		Topic:           sendEmailTopic,
		MinBytes:        10e3,            // 10KB
		MaxBytes:        10e6,            // 10MB
		MaxWait:         1 * time.Second, // Maximum amount of time to wait for new data to come when fetching batches of messages from kafka.
		ReadLagInterval: -1,
	}

	reader := kafka.NewReader(config)
	defer reader.Close()

	var req app.SendEmail
	ctx := context.Background()

	emailClient := app.NewEmailClient(cfg.GetString("mailtrap.username"), cfg.GetString("mailtrap.password"), cfg.GetString("mailtrap.host"), cfg.GetString("mailtrap.port"))
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Error().Msgf("Error while receiving message: %s", err.Error())
			continue
		}

		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s\n", m.Topic, m.Partition, m.Offset, string(m.Value))

		err = json.Unmarshal(m.Value, &req)
		if err != nil {
			log.Error().Msgf("Error while unmarshalling message: %s", err.Error())
			continue
		}

		fmt.Printf("Order Processed, send email to %s (%s). Message is %s\n", req.User.Name, req.User.Email, req.Message)
		err = emailClient.Send(cfg.GetString("mailtrap.from"), []string{req.User.Email}, []byte("From: "+cfg.GetString("mailtrap.from")+"\r\nTo: "+req.User.Email+"\r\nSubject: Order Processed\r\n\r\n"+req.Message+"\r\n"))
	}
}
