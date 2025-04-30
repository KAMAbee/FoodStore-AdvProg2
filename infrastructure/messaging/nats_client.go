package messaging

import (
    "github.com/nats-io/nats.go"
    "log"
)

func NewNatsConnection(url string) (*nats.Conn, error) {
    log.Printf("Connecting to NATS server at %s", url)
    nc, err := nats.Connect(url)
    if err != nil {
        log.Printf("Failed to connect to NATS: %v", err)
        return nil, err
    }
    log.Printf("Connected to NATS server at %s", url)
    return nc, nil
}