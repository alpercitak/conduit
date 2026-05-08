.PHONY: all dev kill build clean run-proxy upstream1 upstream2 run-simulator run-monitor

DIST := ./dist
APPS := conduit mock-upstream simulator monitor

all: build

build: $(APPS)

$(APPS):
	@echo "Building binary for $@..."
	@mkdir -p $(DIST)
	go build -o $(DIST)/$@ ./apps/$(shell [ "$@" = "conduit" ] && echo "proxy" || echo "$@")/cmd/$@

run-proxy:
	go run ./apps/proxy/cmd/conduit

run-upstream1:
	UPSTREAM_NAME=alpha PORT=3001 go run ./apps/mock-upstream/cmd/mock-upstream

run-upstream2:
	UPSTREAM_NAME=beta PORT=3002 go run ./apps/mock-upstream/cmd/mock-upstream

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
	@echo "Cleaning up ports 3001, 3002, 8080..."
	@lsof -t -i :3001,3002,8080 | xargs kill -9 2>/dev/null || true

clean:
	rm -rf $(DIST)