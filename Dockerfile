# --- Build Stage ---
FROM golang:1.24.2-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w " -o /app/build/web ./cmd/web

FROM alpine:latest AS final

WORKDIR /app

RUN mkdir -p /app/bin

COPY --from=builder /app/build/web ./bin/web
COPY  .env .

RUN mkdir -p /app/web/static /app/web/templates 
COPY web/static/ ./web/static/
COPY web/templates/ ./web/templates/

EXPOSE 8080

CMD ["./bin/web"]
