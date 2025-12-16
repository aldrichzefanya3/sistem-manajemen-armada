package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/common/platform/rabbitmq"
	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/model"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Info().Msg("No .env file found, using environment variables")
	} else {
		log.Info().Msg(".env file loaded")
	}

	exchange := os.Getenv("RABBIT_EXCHANGE")
	amqpServerURL := os.Getenv("AMQP_SERVER_URL")
	queue := os.Getenv("RABBIT_QUEUE")

	rabbitMQ := rabbitmq.New(amqpServerURL, exchange, queue)
	defer rabbitMQ.Close()

	messages, err := rabbitMQ.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to consume message")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for msg := range messages {
			var data model.EventGeofenceData
			if err := json.Unmarshal(msg.Body, &data); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal body, skipping message")
				continue
			}
			log.Info().Interface("data", data).Msg("Geofence alerts triggered..")
		}
	}()

	<-sigChan
	log.Info().Msg("Shutting down consumer")
}
