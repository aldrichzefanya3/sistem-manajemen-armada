package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/common/util"
	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
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

	broker := os.Getenv("PUBSUB_BROKER")
	clientID := os.Getenv("PUBLISHER_CLIENT_ID")
	topic := os.Getenv("PUBSUB_TOPIC")

	if broker == "" || clientID == "" || topic == "" {
		log.Fatal().
			Str("PUBSUB_BROKER", broker).
			Str("PUBSUB_CLIENT_ID", clientID).
			Str("PUBSUB_TOPIC", topic).
			Msg("Missing required environment variables")
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetAutoReconnect(true).
		SetConnectRetry(true)

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msg("MQTT connection error")
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("Publisher started (interval: 2s)")

	for {
		select {
		case <-ticker.C:
			var mockData model.MockData
			vehicleID := "B1428YWV"

			isIterating := false
			for range 3 {
				mockData = model.MockData{
					VehicleID: vehicleID,
					Latitude:  util.RandomLatitude(),
					Longitude: util.RandomLongitude(),
					Timestamp: time.Now().Unix(),
				}
				isIterating = true
			}

			if !isIterating {
				mockData = model.MockData{
					VehicleID: util.RandomVehicleID(),
					Latitude:  util.RandomLatitude(),
					Longitude: util.RandomLongitude(),
					Timestamp: time.Now().Unix(),
				}
			}

			payload, err := json.Marshal(mockData)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal payload")
				continue
			}

			token := client.Publish(topic, 1, false, payload)
			token.Wait()

			if token.Error() != nil {
				log.Error().Err(token.Error()).Msg("Publish failed")
			} else {
				log.Info().
					Str("topic", topic).
					RawJSON("payload", payload).
					Msg("Message published")
			}

		case <-quit:
			log.Info().Msg("Shutting down publisher")
			client.Disconnect(250)
			return
		}
	}
}
