package auth

import (
	pbusers "api-gateway-service/proto/api-gateway/users"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type AuthHandlers struct {
	natsConn *nats.Conn
	timeout  time.Duration
}

type Config struct {
	NatsConn *nats.Conn
	Timeout  time.Duration
}

func NewAuthHandlers(cfg Config) *AuthHandlers {
	return &AuthHandlers{
		natsConn: cfg.NatsConn,
		timeout:  cfg.Timeout,
	}
}

const (
	authCreateSubject = "users.create"
)

func (n *AuthHandlers) PublishAuthRegister(req pbusers.CreateReq) (pbusers.CreateResp, error) {
	createReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return pbusers.CreateResp{}, fmt.Errorf("proto.Marshal: %w", err)
	}
	replyMsg, err := n.natsConn.Request(authCreateSubject, createReqBytes, n.timeout)
	if err != nil {
		return pbusers.CreateResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbusers.CreateResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbusers.CreateResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}
	if reply.Error != "" {
		return pbusers.CreateResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}

const (
	authReadSubject = "users.read"
)

func (n *AuthHandlers) PublishAuthRead() (pbusers.ReadResp, error) {
	replyMsg, err := n.natsConn.Request(authReadSubject, nil, n.timeout)
	if err != nil {
		return pbusers.ReadResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbusers.ReadResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbusers.ReadResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return pbusers.ReadResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}

const (
	authUpdateSubject = "users.update"
)

func (n *AuthHandlers) PublishAuthUpdate(req pbusers.UpdateReq) error {
	updateReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	replyMsg, err := n.natsConn.Request(authUpdateSubject, updateReqBytes, n.timeout)
	if err != nil {
		return fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbusers.UpdateResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return nil
}

const (
	authDeleteSubject = "users.delete"
)

func (n *AuthHandlers) PublishAuthDelete(req pbusers.DeleteReq) error {
	deleteReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	replyMsg, err := n.natsConn.Request(authDeleteSubject, deleteReqBytes, n.timeout)
	if err != nil {
		return fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbusers.DeleteResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return nil
}

const (
	authSubject = "users.auth"
)

func (n *AuthHandlers) PublishAuth(req pbusers.AuthReq) (pbusers.AuthResp, error) {
	authReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return pbusers.AuthResp{}, fmt.Errorf("proto.Marshal: %w", err)
	}

	replyMsg, err := n.natsConn.Request(authSubject, authReqBytes, n.timeout)
	if err != nil {
		return pbusers.AuthResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbusers.AuthResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbusers.AuthResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return pbusers.AuthResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}
