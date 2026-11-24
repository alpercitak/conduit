.PHONY: all dev kill build clean run-proxy run-upstream1 run-upstream2 run-simulator deps up up-detached down logs help

DIST := ./dist
APPS := proxy upstream simulator
DOCKER_COMPOSE := docker-compose -f infra/docker-compose.yaml

all: build

build: deps $(APPS)

$(APPS):
	@echo "Building $@..."
	@mkdir -p $(DIST)
	@go build -o $(DIST)/$@ ./apps/$@/cmd/$@

deps:
	@go mod download

run-proxy:
	go run ./apps/proxy/cmd/proxy

run-upstream1:
	UPSTREAM_NAME=alpha PORT=3001 LATENCY_MS=30 JITTER_MS=50 ERROR_RATE=0.01 go run ./apps/upstream/cmd/upstream

run-upstream2:
	UPSTREAM_NAME=beta PORT=3002 LATENCY_MS=200 JITTER_MS=800 ERROR_RATE=0.15 go run ./apps/upstream/cmd/upstream

run-simulator:
	go run ./apps/simulator/cmd/simulator

dev: kill deps
	@echo "Starting local development stack..."
	@$(MAKE) run-upstream1 &
	@$(MAKE) run-upstream2 &
	@$(MAKE) run-simulator &
	@$(MAKE) run-proxy

kill:
	@echo "Cleaning up ports..."
	@for port in 3001 3002 8080; do \
		lsof -ti :$$port | xargs -r kill -9; \
	done

clean:
	rm -rf $(DIST)

up:
	$(DOCKER_COMPOSE) up --build --remove-orphans

up-detached:
	$(DOCKER_COMPOSE) up -d --build --remove-orphans

down:
	$(DOCKER_COMPOSE) down

logs:
	$(DOCKER_COMPOSE) logs -f

help:
	@echo "make up             - run full stack with logs"
	@echo "make up-detached    - run full stack in background"
	@echo "make dev            - run local development stack"
	@echo "make build          - build binaries"
	@echo "make clean          - remove built artifacts"
	@echo "make down           - stop docker stack"