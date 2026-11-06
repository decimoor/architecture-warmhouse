FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o temperature-api .

FROM alpine:3.20

RUN adduser -D -H -s /sbin/nologin appuser
COPY --from=builder /app/temperature-api /usr/local/bin/temperature-api

USER appuser
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/temperature-api"]
