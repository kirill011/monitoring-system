package natslisteners

import (
	pbgetmail "auth-service/proto"
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	getEmailSubject = "users.get_email"
	usersQueue      = "users"
)

func (n *NatsListeners) listen() error {
	_, err := n.natsConn.QueueSubscribe(getEmailSubject, usersQueue, n.getEmailHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+getEmailSubject+"): %w", err)
	}

	return nil
}

func (n *NatsListeners) getEmailHandler(msg *nats.Msg) {
	var request pbgetmail.GetEmailReq

	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error("proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)

		if err := n.natsConn.Publish(msg.Reply, nil); err != nil {
			n.log.Error("n.natsConn.Publish", zap.Error(err))
		}
		return
	}

	email, err := n.authService.GetEmailsByIDs(context.Background(), request.GetUserID())
	if err != nil {
		n.log.Error("n.authService.GetEmailsByIDs", zap.Error(err))

		if err := n.natsConn.Publish(msg.Reply, nil); err != nil {
			n.log.Error("n.natsConn.Publish", zap.Error(err))
		}
		return
	}

	response := pbgetmail.GetEmailResp{
		Email: email,
	}

	responseBytes, err := proto.Marshal(&response)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))

		if err := n.natsConn.Publish(msg.Reply, nil); err != nil {
			n.log.Error("n.natsConn.Publish", zap.Error(err))
		}
		return
	}

	if err := n.natsConn.Publish(msg.Reply, responseBytes); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}
