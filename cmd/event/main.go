package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/common/platform/rabbitmq"
	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/common/util"
	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/model"
	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Info().Msg("No .env file found, using environment variables")
	} else {
		log.Info().Msg(".env file loaded")
	}

	rabbitMQ := rabbitmq.New(os.Getenv("AMQP_SERVER_URL"))
	defer rabbitMQ.Close()

	centerLat := -6.2088
	centerLon := 106.8456
	radius := 50.0

	mockVehicleLat := -6.2088
	mockVehicleLon := 106.8456

	distance := util.CalculateRadius(centerLat, centerLon, mockVehicleLat, mockVehicleLon)

	if distance <= radius {
		eventGeofence := model.EventGeofenceData{
			VehicleID: "B1234XYZ",
			Event:     "geofence_entry",
			Location: model.Location{
				Latitude:  -6.2088,
				Longitude: 106.8456,
			},
			Timestamp: 1715003456,
		}

		body, _ := json.Marshal(eventGeofence)

		err = rabbitMQ.Publish("fleet.events", "", false, false, amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to publish event")
		} else {
			log.Info().Interface("data", body).Msg("Succesfully publish event geofence")
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Info().Msg("Shutdown signal received. Exiting...")
}
