.PHONY: build test-integration test-e2e lint docker-build docker-up docker-down migrate-up migrate-down load-test

build:
	go build -o bin/server ./cmd/server

test-integration:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down -v

test-e2e:
	./scripts/run_e2e_tests.sh

lint:
	golangci-lint run --config .golangci.yml ./...

docker-build:
	docker build -t pr-reviewer-service:latest .

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down
