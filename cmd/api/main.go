package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/danymarita/go-kafka/dep/kafka"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var logger = log.With().Str("pkg", "main").Logger()

var (
	appName       string
	listenAddrApi string
	// Kafka
	kafkaBrokerUrl string
	kafkaVerbose   bool
	kafkaClientId  string
	kafkaTopic     string
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
	appName = cfg.GetString("app.name")
	listenAddrApi = fmt.Sprintf("%s:%s", cfg.GetString("app.host"), cfg.GetString("app.port"))
	// Kafka
	kafkaBrokerUrl = cfg.GetString("kafka.broker_url")
	kafkaVerbose = cfg.GetBool("kafka.verbose")
	kafkaClientId = cfg.GetString("kafka.client_id")
	kafkaTopic = cfg.GetString("kafka.topic")

	kafkaProducer, err := kafka.Configure(strings.Split(kafkaBrokerUrl, ","), kafkaClientId, kafkaTopic)
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

	form := &struct {
		Text string `form:"text" json:"text"`
	}{}

	ctx.Bind(form)
	formInBytes, err := json.Marshal(form)
	if err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("error while marshalling json: %s", err.Error()),
			},
		})

		ctx.Abort()
		return
	}

	err = kafka.Push(parent, nil, formInBytes)
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
		"data":    form,
	})
}
