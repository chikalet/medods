
FROM golang:1.23-alpine as builder

WORKDIR /app

RUN apk add --no-cache git
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY . .
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org,direct
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/medods ./cmd/medods/main.go

FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache postgresql-client

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /app/bin/medods /app/medods
COPY --from=builder /app/migrations /app/migrations
COPY entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

EXPOSE 4000
ENTRYPOINT ["/app/entrypoint.sh"]