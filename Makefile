.PHONY: proxy upstream dev

proxy:
	go run ./apps/proxy/cmd/conduit

upstream:
	UPSTREAM_NAME=alpha \
	go run ./apps/mock-upstream/cmd/mock-upstream

dev:
	(concurrently \
		"make upstream" \
		"make proxy")