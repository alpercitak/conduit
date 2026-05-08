.PHONY: proxy upstream1 upstream2 dev

proxy:
	go run ./apps/proxy/cmd/conduit

upstream1:
	UPSTREAM_NAME=alpha PORT=3001 \
	go run ./apps/mock-upstream/cmd/mock-upstream

upstream2:
	UPSTREAM_NAME=beta PORT=3002 \
	go run ./apps/mock-upstream/cmd/mock-upstream

dev:
	concurrently \
		"make upstream1" \
		"make upstream2" \
		"make proxy"

kill:
	-kill -9 $(lsof -t -i :3001) 2>/dev/null || true
	-kill -9 $(lsof -t -i :3002) 2>/dev/null || true
	-kill -9 $(lsof -t -i :8080) 2>/dev/null || true