.PHONY: all dev kill build clean run-proxy upstream1 upstream2 run-simulator run-monitor

DIST := ./dist
APPS := proxy upstream simulator monitor
DOCKER_COMPOSE = docker-compose -f infra/docker-compose.yaml

all: build

build: $(APPS)

$(APPS):
	@echo "Building binary for $@..."
	@mkdir -p $(DIST)
	go build -o $(DIST)/$@ ./apps/$@/cmd/$@

run-proxy:
	go run ./apps/proxy/cmd/proxy

run-upstream1:
	UPSTREAM_NAME=alpha PORT=3001 go run ./apps/upstream/cmd/upstream

run-upstream2:
	UPSTREAM_NAME=beta PORT=3002 go run ./apps/upstream/cmd/upstream

run-simulator:
	go run ./apps/simulator/cmd/simulator

run-monitor:
	go run ./apps/monitor/cmd/monitor
    
dev: kill
	concurrently \
		"make run-upstream1" \
		"make run-upstream2" \
		"make run-proxy"

kill:
	@echo "Cleaning up ports"
	@for port in 3001 3002 8080; do \
		lsof -ti :$$port | xargs -r kill -9; \
	done

clean:
	rm -rf $(DIST)

dc-up: 
	$(DOCKER_COMPOSE) up -d

dc-up-build: 
	$(DOCKER_COMPOSE) up --build --remove-orphans

dc-down:
	$(DOCKER_COMPOSE) down

dc-build:
	$(DOCKER_COMPOSE) build

dc-logs:
	$(DOCKER_COMPOSE) logs -f