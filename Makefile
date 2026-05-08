.PHONY: proxy upstream1 upstream2 simulator monitor dev kill build clean

DIST=./dist

proxy:
	go run ./apps/proxy/cmd/conduit

upstream1:
	UPSTREAM_NAME=alpha PORT=3001 \
	go run ./apps/mock-upstream/cmd/mock-upstream

upstream2:
	UPSTREAM_NAME=beta PORT=3002 \
	go run ./apps/mock-upstream/cmd/mock-upstream

simulator:
	go run ./apps/simulator/cmd/simulator

monitor:
	go run ./apps/monitor/cmd/monitor
	
dev:
	concurrently \
		"make upstream1" \
		"make upstream2" \
		"make proxy"

kill:
	-kill -9 $(lsof -t -i :3001) 2>/dev/null || true
	-kill -9 $(lsof -t -i :3002) 2>/dev/null || true
	-kill -9 $(lsof -t -i :8080) 2>/dev/null || true

build:
	mkdir -p $(DIST)
	go build -o $(DIST)/conduit ./apps/proxy/cmd/conduit
	go build -o $(DIST)/mock-upstream ./apps/mock-upstream/cmd/mock-upstream
	go build -o $(DIST)/simulator ./apps/simulator/cmd/simulator
	go build -o $(DIST)/monitor ./apps/monitor/cmd/monitor

clean:
	rm -rf $(DIST)	