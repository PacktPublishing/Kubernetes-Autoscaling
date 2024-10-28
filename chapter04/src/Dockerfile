FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer .

FROM alpine:3.20.3

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/consumer .
RUN adduser -D appuser
USER appuser

CMD ["./consumer"]