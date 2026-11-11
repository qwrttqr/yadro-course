package events

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"yadro.com/course/search/core"
)

const (
	updateTopic        = "xkcd.db.updated"
	indexUpdateTimeout = 5 * time.Second
)

type Broker struct {
	connection   *nats.Conn
	subscription *nats.Subscription
	log          *slog.Logger
}

func New(address string, searcher core.Searcher, log *slog.Logger) (*Broker, error) {
	nc, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}
	log.Debug("connected to broker", "address", address)

	sub, err := nc.Subscribe(updateTopic, func(msg *nats.Msg) {
		ctx, cancel := context.WithTimeout(context.Background(), indexUpdateTimeout)
		defer cancel()
		log.Debug("db update event", "message", msg.Data)
		if err = searcher.BuildIndex(ctx); err != nil {
			log.Error("index re-build failed", "error", err)
		}
	})
	if err != nil {
		return nil, err
	}

	return &Broker{connection: nc, subscription: sub, log: log}, nil
}

func (b *Broker) Close() {
	if err := b.subscription.Unsubscribe(); err != nil {
		b.log.Error("could not unsubscribe", "error", err)
	}
	b.connection.Close()
}
