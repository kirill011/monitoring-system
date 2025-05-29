package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Nats struct {
	NatsConn *nats.Conn
}

func NewNats(natsConnString string) (*Nats, error) {
	natsConn, err := nats.Connect(natsConnString)
	if err != nil {
		return nil, fmt.Errorf("nats.Connect: %w", err)
	}

	return &Nats{
		NatsConn: natsConn,
	}, nil
}

func (s *Nats) Close(_ context.Context) error {
	s.NatsConn.Close()
	return nil
}
