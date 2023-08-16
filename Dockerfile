FROM golang:1.16-alpine as builder

WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN go mod tidy
RUN go mod download
RUN go get go.opentelemetry.io/otel
RUN go get go.opentelemetry.io/otel/exporters/stdout
RUN go get go.opentelemetry.io/otel/trace
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
