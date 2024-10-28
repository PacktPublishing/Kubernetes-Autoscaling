FROM --platform=$BUILDPLATFORM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -a -installsuffix cgo -o producer .

FROM alpine:3.20.3

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/producer .
RUN adduser -D appuser
USER appuser

CMD ["./producer"]