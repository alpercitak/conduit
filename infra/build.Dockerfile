FROM golang:1.26.3-alpine AS builder
WORKDIR /src

ARG APP_PATH

COPY go.work ./
COPY apps/ ./apps/

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /bin/app \
    ./apps/${APP_PATH}/cmd/${APP_PATH}/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN adduser -D conduituser
USER conduituser

COPY --from=builder /bin/app /app

ENTRYPOINT ["/app"]