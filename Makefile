.PHONY: run-app
run-app: ## run API service 
	go build -o ./bin/app ./cmd/app/
	./bin/app

.PHONY: run-subs
run-subs: ## run Subscriber service 
	go build -o ./bin/subscriber ./cmd/subscriber/
	./bin/subscriber

.PHONY: run-pub
run-pub: ## run publish contains random mock data
	go run scripts/publish/mock_data.go