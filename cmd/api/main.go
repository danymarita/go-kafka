package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/danymarita/go-kafka/dep/app"

	"github.com/danymarita/go-kafka/dep/kafka"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var logger = log.With().Str("pkg", "main").Logger()

var (
	appName       string
	listenAddrApi string
	// Kafka
	kafkaBrokerUrl string
	kafkaVerbose   bool
	kafkaClientId  string
	orderTopic     string
)

func main() {
	cfg := app.ReadViperConfig()
	appName = cfg.GetString("app.name")
	listenAddrApi = fmt.Sprintf("%s:%s", cfg.GetString("app.host"), cfg.GetString("app.port"))
	// Kafka
	kafkaBrokerUrl = cfg.GetString("kafka.broker_url")
	kafkaVerbose = cfg.GetBool("kafka.verbose")
	kafkaClientId = cfg.GetString("kafka.client_id")
	orderTopic = cfg.GetString("topics.order_req")

	kafkaProducer, err := kafka.Configure(strings.Split(kafkaBrokerUrl, ","), kafkaClientId, orderTopic)
	if err != nil {
		logger.Error().Str("error", err.Error()).Msg("unable to configure kafka")
		return
	}
	defer kafkaProducer.Close()

	var errChan = make(chan error, 1)
	go func() {
		log.Info().Msgf("starting server at %s", listenAddrApi)
		errChan <- server(listenAddrApi)
	}()

	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalChan:
		logger.Info().Msg("got an interrupt, exiting...")
	case err := <-errChan:
		if err != nil {
			logger.Error().Err(err).Msg("error while running api, exiting...")
		}
	}
}

func server(listenAddr string) (err error) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.POST("/api/v1/data", postDataToKafka)
	// list routes for debugging purpose
	for _, routeInfo := range router.Routes() {
		logger.Debug().
			Str("path", routeInfo.Path).
			Str("handler", routeInfo.Handler).
			Str("method", routeInfo.Method).
			Msg("registered routes")
	}

	if err = router.Run(listenAddr); err != nil {

	}
	logger.Info().Msg(fmt.Sprintf("%s running at %s...", appName, listenAddr))
	return
}

func postDataToKafka(ctx *gin.Context) {
	parent := context.Background()
	defer parent.Done()

	order := &app.OrderReq{}

	ctx.Bind(order)
	orderInBytes, err := json.Marshal(order)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("error while marshalling json: %s", err.Error()),
			},
		})

		ctx.Abort()
		return
	}

	err = kafka.Push(parent, nil, orderInBytes)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("error while push message to kafka: %s", err.Error()),
			},
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success push data into kafka",
		"data":    order,
	})
}
