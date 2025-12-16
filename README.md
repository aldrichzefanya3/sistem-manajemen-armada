# Whatsapp Service Golang


## Setup

1. Clone this repository
2. Run `go mod download`
3. Install `make` if you don't have this, so you can running each services easily
4. Run `make run-app` for starting endpoints service
   Run `make run-subs` for starting MQTT subscriber
   Run `make run-event` for starting event geofence service

 or simply you can use docker so you don't need to setup everthing

 1. Clone this repository
 2. Install docker && docker-compose if you don't have it.
 3. Run `docker-compose build`
 4. Run `docker-compose up`

### ENV
- There's env that need to be setup if you are not using docker
  DATABASE_URL=your-full-database-connection-string
  PORT=
  PUBLISHER_CLIENT_ID=example-pub-client-id
  PUBSUB_BROKER=tcp://localhost:1883
  PUBSUB_TOPIC=example-topic
  SUBSCRIBER_PORT=
  SUBSCRIBER_CLIENT_ID=example-sub-client-id
  AMQP_SERVER_URL=amqp://guest:guest@localhost:5672/
