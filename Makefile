.PHONY: run-app
run-app: ## run API service 
	go build -o ./bin/app ./cmd/app/
	./bin/app

.PHONY: run-subs
run-subs: ## run Subscriber service 
	go build -o ./bin/subscriber ./cmd/subscriber/
	./bin/subscriber

.PHONY: run-event
run-event: ## run event geofence
	go build -o ./bin/event ./cmd/event/
	./bin/event