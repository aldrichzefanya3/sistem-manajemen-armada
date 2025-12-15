package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/common/platform/postgres"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Info().Msg("No .env file found, using environment variables")
	} else {
		log.Info().Msg(".env file loaded")
	}

	dbPool := postgres.New(os.Getenv("DATABASE_URL"))
	defer dbPool.Close()

	port := os.Getenv("SUBSCRIBER_PORT")
	topic := os.Getenv("PUBSUB_TOPIC")

	server := &http.Server{
		Addr: fmt.Sprint(":", port),
	}

	opts := mqtt.NewClientOptions().
		AddBroker(os.Getenv("PUBSUB_BROKER")).
		SetClientID(os.Getenv("SUBSCRIBER_CLIENT_ID")).
		SetKeepAlive(60 * time.Second).
		SetAutoReconnect(true).
		SetConnectTimeout(30 * time.Second).
		SetOrderMatters(false)

	messageChan := make(chan []byte, 100)

	var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		log.Info().Str("topic", msg.Topic()).Msg("Received message")
		select {
		case messageChan <- msg.Payload():
			log.Info().Msg("Message queued for processing")
		case <-time.After(1 * time.Second):
			log.Warn().Msg("Message dropped (channel full or timeout)")
		}
	}

	opts.SetDefaultPublishHandler(messageHandler)

	client := mqtt.NewClient(opts)

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		token := client.Connect()
		if token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Msgf("MQTT connection attempt %d failed", i+1)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * 2 * time.Second)
				continue
			}
			log.Fatal().Msg("Max MQTT connection retries reached")
		}
		break
	}

	if token := client.Subscribe(topic, 1, messageHandler); token.Wait() && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msg("Failed to subscribe to topic")
	}
	log.Info().Str("topic", topic).Msg("Subscribed to topic")

	go func() {
		log.Info().Msgf("Subscriber server running on port :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	go worker(ctx, dbPool, messageChan)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutdown signal received")

	close(messageChan)

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown error")
	}

	log.Info().Msg("HTTP server shutdown complete")

	client.Disconnect(250)
	log.Info().Msg("MQTT client disconnected")

	log.Info().Msg("Subscriber server exited gracefully")
}

func worker(ctx context.Context, db *pgxpool.Pool, ch <-chan []byte) {
	for payload := range ch {
		ProcessMessage(ctx, db, payload)
	}
}
