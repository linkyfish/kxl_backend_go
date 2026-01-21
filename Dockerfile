FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/kxl-api ./cmd/api

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /app/kxl-api /app/kxl-api
COPY --from=builder /app/config /app/config
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static

# Uploads are stored on disk; mount a volume in production.
RUN mkdir -p /app/uploads

ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8787

EXPOSE 8787

ENTRYPOINT ["/app/kxl-api"]

