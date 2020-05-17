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
	dep_kafka "github.com/danymarita/go-kafka/dep/kafka"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	_ "github.com/segmentio/kafka-go/snappy"
	"github.com/sirupsen/logrus"
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
	orderTopic = cfg.GetString("topics.order_req")
	sendEmailTopic = cfg.GetString("topics.send_email")

	log.Info().Msg(fmt.Sprintf("Running Kafka worker. ConsumerGroup %s Topic %s...", kafkaConsumerGroup, orderTopic))

	ctx, cancel := context.WithCancel(context.Background())
	idleConnectionClosed := make(chan struct{})
	run := true
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan

		log.Printf("Caught signal %v: terminating\n", signalChan)
		run = false
		cancel()
		close(idleConnectionClosed)
	}()

	brokers := strings.Split(kafkaBrokerUrl, ",")

	config := kafka.ReaderConfig{
		Brokers:         brokers,
		GroupID:         kafkaConsumerGroup,
		Topic:           orderTopic,
		MinBytes:        10e3,            // 10KB
		MaxBytes:        10e6,            // 10MB
		MaxWait:         1 * time.Second, // Maximum amount of time to wait for new data to come when fetching batches of messages from kafka.
		ReadLagInterval: -1,
	}

	reader := kafka.NewReader(config)
	defer reader.Close()

	kafkaProducer, err := dep_kafka.Configure(brokers, kafkaClientId, sendEmailTopic)
	if err != nil {
		log.Error().Str("error", err.Error()).Msg("unable to configure kafka producer")
		return
	}
	defer kafkaProducer.Close()

	var req app.OrderReq
	for run == true {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Error().Msgf("Error while receiving message: %s", err.Error())
			continue
		}

		fmt.Printf("message at topic/partition/offset %v/%v/%v: %s\n", m.Topic, m.Partition, m.Offset, string(m.Value))

		err = json.Unmarshal(m.Value, &req)
		if err != nil {
			log.Error().Msgf("error while unmarshalling message: %s", err.Error())
			continue
		}

		fmt.Printf("Order Processed, send email to %s (%s)\n", req.User.Name, req.User.Email)

		email := app.SendEmail{
			User: app.User{
				Name:  req.User.Name,
				Email: req.User.Email,
			},
			Message: fmt.Sprintf("Your order %s (%s) was processed", req.Product.Name, req.Product.Code),
		}
		emailInBytes, err := json.Marshal(email)
		if err != nil {
			fmt.Printf("Error Marshalling Order Request by %s for %s (%s)", req.User.Name, req.Product.Name, req.Product.Code)
			continue
		}
		err = dep_kafka.Push(ctx, nil, emailInBytes)
		if err != nil {
			fmt.Printf("Error Push Order Request by %s for %s (%s)", req.User.Name, req.Product.Name, req.Product.Code)
			continue
		}
	}
	<-idleConnectionClosed
	logrus.Infoln("[Worker] Bye")
}
