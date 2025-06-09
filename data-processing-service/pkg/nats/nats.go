package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Nats struct {
	NatsConn *nats.Conn
	log      *zap.Logger
}

func NewNats(natsConnString string, log *zap.Logger) (*Nats, error) {
	natsConn := &Nats{
		log: log,
	}
	var err error
	natsConn.NatsConn, err = nats.Connect(natsConnString, nats.ErrorHandler(natsConn.natsErrHandler))
	if err != nil {
		return nil, fmt.Errorf("nats.Connect: %w", err)
	}

	return natsConn, nil
}

func (n *Nats) natsErrHandler(nc *nats.Conn, sub *nats.Subscription, natsErr error) {
	if natsErr == nats.ErrSlowConsumer {
		pendingMsgs, _, err := sub.Pending()
		if err != nil {
			n.log.Error("couldn't get pending messages", zap.Error(err))
			return
		}
		n.log.Error("slow consumer", zap.Int("pendingMsgs", pendingMsgs), zap.String("subject", sub.Subject))
	}
}

func (s *Nats) Close(_ context.Context) error {
	s.NatsConn.Close()
	return nil
}
