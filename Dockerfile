FROM golang:1.16-alpine as builder

WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN go get go.opentelemetry.io/otel@v0.20.0
RUN go get go.opentelemetry.io/otel/exporters/stdout@v0.20.0
RUN go get go.opentelemetry.io/otel/sdk/trace@v0.20.0
RUN go get go.opentelemetry.io/otel/trace@v0.20.0
RUN go get go.opentelemetry.io/otel/attribute@v0.20.0
RUN go get go.opentelemetry.io/otel/codes@v0.20.0
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
