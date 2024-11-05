# Build Stage
FROM golang:1.23.2-alpine3.20 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/main .
COPY app.yml .
COPY start.sh .
COPY db/migration ./db/migration

EXPOSE 8080
CMD ["/app/main"]
#can be removed - left for reference
ENTRYPOINT ["/app/start.sh"]