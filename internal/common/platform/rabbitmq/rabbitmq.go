package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func New(amqpServerURL string) *amqp.Channel {
	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	if err != nil {
		panic(err)
	}

	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}

	err = channelRabbitMQ.ExchangeDeclare(
		"fleet.events",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	queue, err := channelRabbitMQ.QueueDeclare(
		"geofence_alerts",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	err = channelRabbitMQ.QueueBind(
		queue.Name,
		"",
		"fleet.events",
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	return channelRabbitMQ
}
