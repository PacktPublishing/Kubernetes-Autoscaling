FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dummyapi .

FROM alpine:3.20.3

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/dummyapi .
RUN adduser -D appuser
USER appuser

CMD ["./dummyapi"]

EXPOSE 8080