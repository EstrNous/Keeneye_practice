FROM golang:1.26-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /keeneye_app ./app/cmd/server/main.go

FROM alpine:3.19

RUN apk update && apk add --no-cache tzdata curl

WORKDIR /root/

COPY --from=builder /keeneye_app .

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/healthz || exit 1
CMD ["./keenye_app"]