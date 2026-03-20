FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/cdn-control ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/cdn-control /usr/local/bin/cdn-control
COPY configs/control-config.yaml /etc/cdn-control/config.yaml
EXPOSE 8090
CMD ["cdn-control", "-config", "/etc/cdn-control/config.yaml"]
