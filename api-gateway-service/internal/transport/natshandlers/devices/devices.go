package devices

import (
	pbdevices "api-gateway-service/proto/api-gateway/devices"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type DevicesHandler struct {
	natsConn *nats.Conn
	timeout  time.Duration
}

type Config struct {
	NatsConn *nats.Conn
	Timeout  time.Duration
}

func NewDevicesHandlers(cfg Config) *DevicesHandler {
	return &DevicesHandler{
		natsConn: cfg.NatsConn,
		timeout:  cfg.Timeout,
	}
}

const (
	devicesCreateSubject = "devices.create"
)

func (n *DevicesHandler) PublishCreate(req pbdevices.CreateReq) (pbdevices.CreateResp, error) {
	createReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return pbdevices.CreateResp{}, fmt.Errorf("proto.Marshal: %w", err)
	}
	replyMsg, err := n.natsConn.Request(devicesCreateSubject, createReqBytes, n.timeout)
	if err != nil {
		return pbdevices.CreateResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbdevices.CreateResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbdevices.CreateResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}
	if reply.Error != "" {
		return pbdevices.CreateResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}

const (
	devicesReadSubject = "devices.read"
)

func (n *DevicesHandler) PublishRead() (pbdevices.ReadResp, error) {
	replyMsg, err := n.natsConn.Request(devicesReadSubject, nil, n.timeout)
	if err != nil {
		return pbdevices.ReadResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbdevices.ReadResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbdevices.ReadResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return pbdevices.ReadResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}

const (
	devicesUpdateSubject = "devices.update"
)

func (n *DevicesHandler) PublishUpdate(req pbdevices.UpdateReq) error {
	updateReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	replyMsg, err := n.natsConn.Request(devicesUpdateSubject, updateReqBytes, n.timeout)
	if err != nil {
		return fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbdevices.UpdateResp
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
	devicesDeleteSubject = "devices.delete"
)

func (n *DevicesHandler) PublishDelete(req pbdevices.DeleteReq) error {
	deleteReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	replyMsg, err := n.natsConn.Request(devicesDeleteSubject, deleteReqBytes, n.timeout)
	if err != nil {
		return fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbdevices.DeleteResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return nil
}
