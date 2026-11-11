package events

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

const updateTopic = "xkcd.db.updated"

type Broker struct {
	connection *nats.Conn
	log        *slog.Logger
}

func New(address string, log *slog.Logger) (*Broker, error) {
	nc, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}
	log.Debug("connected to broker", "address", address)
	return &Broker{connection: nc, log: log}, nil
}

func (b *Broker) Close() {
	b.connection.Close()
}

func (b *Broker) NotifyDBUpdated() error {
	return b.connection.Publish(updateTopic, []byte("XKCD DB has been updated"))
}

func (b *Broker) NotifyDBCleaned() error {
	return b.connection.Publish(updateTopic, []byte("XKCD DB has been cleaned"))
}
