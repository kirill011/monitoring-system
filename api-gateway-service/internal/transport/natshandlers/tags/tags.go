package tags

import (
	pbtags "api-gateway-service/proto/api-gateway/tags"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type TagsHandler struct {
	natsConn *nats.Conn
	timeout  time.Duration
}

type Config struct {
	NatsConn *nats.Conn
	Timeout  time.Duration
}

func NewTagsHandlers(cfg Config) *TagsHandler {
	return &TagsHandler{
		natsConn: cfg.NatsConn,
		timeout:  cfg.Timeout,
	}
}

const (
	tagsCreateSubject = "tags.create"
)

func (n *TagsHandler) PublishCreate(req pbtags.CreateReq) (pbtags.CreateResp, error) {
	createReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return pbtags.CreateResp{}, fmt.Errorf("proto.Marshal: %w", err)
	}
	replyMsg, err := n.natsConn.Request(tagsCreateSubject, createReqBytes, n.timeout)
	if err != nil {
		return pbtags.CreateResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbtags.CreateResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbtags.CreateResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}
	if reply.Error != "" {
		return pbtags.CreateResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}

const (
	tagsReadSubject = "tags.read"
)

func (n *TagsHandler) PublishRead() (pbtags.ReadResp, error) {
	replyMsg, err := n.natsConn.Request(tagsReadSubject, nil, n.timeout)
	if err != nil {
		return pbtags.ReadResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbtags.ReadResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbtags.ReadResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return pbtags.ReadResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}

const (
	tagsUpdateSubject = "tags.update"
)

func (n *TagsHandler) PublishUpdate(req pbtags.UpdateReq) error {
	updateReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	replyMsg, err := n.natsConn.Request(tagsUpdateSubject, updateReqBytes, n.timeout)
	if err != nil {
		return fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbtags.UpdateResp
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
	tagsDeleteSubject = "tags.delete"
)

func (n *TagsHandler) PublishDelete(req pbtags.DeleteReq) error {
	deleteReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	replyMsg, err := n.natsConn.Request(tagsDeleteSubject, deleteReqBytes, n.timeout)
	if err != nil {
		return fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbtags.DeleteResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return nil
}
