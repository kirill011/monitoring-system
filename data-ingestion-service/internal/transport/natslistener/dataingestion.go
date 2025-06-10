package natslistener

import (
	"data-ingestion-service/internal/models"
	pbdevices "data-ingestion-service/proto/devices"
	pbmessages "data-ingestion-service/proto/messages"

	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	dataIngestionQueue = "data-ingestion"

	devicesUpdatedSubject = "devices.updated"
	readDevicesSubject    = "devices.read"

	saveMessageSubject = "monitoring.msg.save"
)

func (n *NatsListeners) listen() error {
	_, err := n.natsConn.QueueSubscribe(devicesUpdatedSubject, dataIngestionQueue, n.devicesUpdatedHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+devicesUpdatedSubject+"): %w", err)
	}

	err = n.publishDevicesRead()
	if err != nil {
		return fmt.Errorf("n.publishDevicesRead(): %w", err)
	}

	return nil
}

func (n *NatsListeners) devicesUpdatedHandler(msg *nats.Msg) {
	err := n.publishDevicesRead()
	if err != nil {
		n.log.Error("n.publishDevicesRead", zap.Error(err))
	}
}

func (n *NatsListeners) publishDevicesRead() error {
	reply, err := n.natsConn.Request(readDevicesSubject, nil, n.timeout)
	if err != nil {
		return fmt.Errorf("natsConn.Request: %w", err)
	}

	var devicesProto pbdevices.ReadResp
	err = proto.Unmarshal(reply.Data, &devicesProto)
	if err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	devices := make([]models.Device, 0)
	for _, device := range devicesProto.Devices {
		devices = append(devices, models.Device{
			ID:          device.ID,
			Name:        device.Name,
			DeviceType:  device.DeviceType,
			Address:     device.Address,
			Responsible: device.Responsible,
			CreatedAt:   device.CreatedAt.AsTime(),
			UpdatedAt:   device.UpdatedAt.AsTime(),
		})
	}

	n.devicesService.UpdateDevices(devices)
	return nil
}

func (n *NatsListeners) PublishSaveMessage(message models.Message) error {
	deviceId, ok := n.devicesService.GetDeviceIDByIp(message.DeviceIP)
	if !ok {
		return fmt.Errorf("unknown ip")
	}
	pbMessage := pbmessages.MessageSave{
		DeviceID:    deviceId,
		Message:     message.Message,
		MessageType: message.MessageType,
		Component:   message.Component,
	}

	binaryMessage, err := proto.Marshal(&pbMessage)
	if err != nil {
		return fmt.Errorf("proto.Marshal: %w", err)
	}

	_, err = n.js.Publish(saveMessageSubject, binaryMessage)
	if err != nil {
		return fmt.Errorf("js.Publish: %w", err)
	}

	return nil
}
