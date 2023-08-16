FROM golang:1.16-alpine as builder

WORKDIR /app

# Descargamos las dependencias específicas que has mencionado.
RUN go get github.com/gin-gonic/gin
RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/swaggo/gin-swagger
RUN go get github.com/swaggo/gin-swagger/swaggerFiles
RUN go get go.opentelemetry.io/otel
RUN go get go.opentelemetry.io/otel/exporters/stdout
RUN go get go.opentelemetry.io/otel/trace

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
