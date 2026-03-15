FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o slack-export .

FROM alpine:latest
RUN adduser -D appuser
WORKDIR /app
COPY --from=builder /app/slack-export .
USER appuser
ENTRYPOINT ["./slack-export"]
