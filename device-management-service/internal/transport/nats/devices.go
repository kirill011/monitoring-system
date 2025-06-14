package natslisteners

import (
	"context"
	"device-management-service/internal/models"
	"device-management-service/internal/services"
	pbapidevices "device-management-service/proto/api-gateway/devices"
	pbdevices "device-management-service/proto/devices"

	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	getResponsibleSubject = "devices.get_responsible"

	createDevicesSubject = "devices.create"
	readDevicesSubject   = "devices.read"
	updateDevicesSubject = "devices.update"
	deleteDevicesSubject = "devices.delete"

	deviceManagementQueue = "device-management"

	devicesUpdatedSubject = "devices.updated"
)

func (n *NatsListeners) listen() error {
	_, err := n.natsConn.QueueSubscribe(getResponsibleSubject, deviceManagementQueue, n.getResponsibleHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+getResponsibleSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(createDevicesSubject, deviceManagementQueue, n.createHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+createDevicesSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(readDevicesSubject, deviceManagementQueue, n.readHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+readDevicesSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(updateDevicesSubject, deviceManagementQueue, n.updateHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+updateDevicesSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(deleteDevicesSubject, deviceManagementQueue, n.deleteHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+deleteDevicesSubject+"): %w", err)
	}

	return nil
}

func (n *NatsListeners) PublishUpdateEvent() error {
	return n.natsConn.Publish(devicesUpdatedSubject, nil)
}

func (n *NatsListeners) getResponsibleHandler(msg *nats.Msg) {
	responsibles, err := n.devicesService.GetResponsible(context.Background())
	if err != nil {
		n.log.Error("devicesService.GetResponsible",
			zap.Error(err),
		)

		if err := n.natsConn.Publish(msg.Reply, nil); err != nil {
			n.log.Error("n.natsConn.Publish", zap.Error(err))
		}
		return
	}

	getResposibleResp := pbdevices.GetResponsibleResp{
		ResposiblesByDeviceID: make([]*pbdevices.ResposiblesByDeviceID, 0),
	}

	for _, responsible := range responsibles {
		getResposibleResp.ResposiblesByDeviceID = append(getResposibleResp.ResposiblesByDeviceID, &pbdevices.ResposiblesByDeviceID{
			DeviceID:      responsible.Device,
			ResponsibleID: responsible.Responsible,
		})
	}

	getResposibleRespBytes, err := proto.Marshal(&getResposibleResp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err), zap.Any("variable", getResposibleRespBytes))
		return
	}

	if err := n.natsConn.Publish(msg.Reply, getResposibleRespBytes); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
	}
}

func (n *NatsListeners) createHandler(msg *nats.Msg) {
	var request pbapidevices.CreateReq
	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
	}

	created, err := n.devicesService.Create(context.Background(),
		models.Device{
			Name:        request.Device.Name,
			DeviceType:  request.Device.DeviceType,
			Address:     request.Device.Address,
			Responsible: request.Device.Responsible,
		})
	if err != nil {
		n.log.Error("n.devicesService.Create", zap.Error(err))
		n.sendError(msg.Reply, &pbapidevices.CreateResp{Error: err.Error()})
		return
	}

	var createdAt *timestamppb.Timestamp
	if created.CreatedAt != nil {
		createdAt = timestamppb.New(*created.CreatedAt)
	}
	var updatedAt *timestamppb.Timestamp
	if created.UpdatedAt != nil {
		updatedAt = timestamppb.New(*created.UpdatedAt)
	}

	resp := pbapidevices.CreateResp{
		Created: &pbapidevices.Device{
			ID:          int32(created.ID),
			Name:        created.Name,
			DeviceType:  created.DeviceType,
			Address:     created.Address,
			Responsible: created.Responsible,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		},
	}

	binaryResp, err := proto.Marshal(&resp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbapidevices.CreateResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}

	if err := n.PublishUpdateEvent(); err != nil {
		n.log.Error("n.PublishUpdateEvent", zap.Error(err))
	}
}

func (n *NatsListeners) readHandler(msg *nats.Msg) {
	readResult, err := n.devicesService.Read(context.Background())
	if err != nil {
		n.log.Error("n.devicesService.Read",
			zap.Error(err),
		)

		n.sendError(msg.Reply, &pbapidevices.ReadResp{Error: err.Error()})
		return
	}

	readDevicesResp := pbapidevices.ReadResp{
		Devices: convertDevicesToProtoDevices(readResult.Devices),
	}

	readDevicesRespBytes, err := proto.Marshal(&readDevicesResp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err), zap.Any("variable", readDevicesRespBytes))

		n.sendError(msg.Reply, &pbapidevices.ReadResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, readDevicesRespBytes); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
	}
}

func (n *NatsListeners) sendError(subject string, message proto.Message) {
	binaryResp, err := proto.Marshal(message)
	if err != nil {
		n.log.Error("sendError: proto.Marshal", zap.Error(err))
		return
	}
	if err := n.natsConn.Publish(subject, binaryResp); err != nil {
		n.log.Error("sendError: n.natsConn.Publish", zap.Error(err))
		return
	}
}

func convertDevicesToProtoDevices(devices []models.Device) []*pbapidevices.Device {
	var result []*pbapidevices.Device
	for _, device := range devices {
		var createdAt *timestamppb.Timestamp
		if device.CreatedAt != nil {
			createdAt = timestamppb.New(*device.CreatedAt)
		}

		var updatedAt *timestamppb.Timestamp
		if device.UpdatedAt != nil {
			updatedAt = timestamppb.New(*device.UpdatedAt)
		}

		result = append(result, &pbapidevices.Device{
			ID:          device.ID,
			Name:        device.Name,
			DeviceType:  device.DeviceType,
			Address:     device.Address,
			Responsible: device.Responsible,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}
	return result
}

func (n *NatsListeners) updateHandler(msg *nats.Msg) {
	var request pbapidevices.UpdateReq
	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
	}

	var name *string
	if request.Device.GetName() != "" {
		name = &request.Device.Name
	}
	var deviceType *string
	if request.Device.GetDeviceType() != "" {
		deviceType = &request.Device.DeviceType
	}
	var address *string
	if request.Device.GetAddress() != "" {
		address = &request.Device.Address
	}

	err = n.devicesService.Update(context.Background(),
		services.UpdateDeviceParams{
			ID:          request.Device.GetID(),
			Name:        name,
			DeviceType:  deviceType,
			Address:     address,
			Responsible: request.Device.Responsible,
		})
	if err != nil {
		n.log.Error("n.devicesService.Update", zap.Error(err))
		n.sendError(msg.Reply, &pbapidevices.UpdateResp{Error: err.Error()})
		return
	}

	resp := pbapidevices.UpdateResp{}

	binaryResp, err := proto.Marshal(&resp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbapidevices.UpdateResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}

	if err := n.PublishUpdateEvent(); err != nil {
		n.log.Error("n.PublishUpdateEvent", zap.Error(err))
	}
}

func (n *NatsListeners) deleteHandler(msg *nats.Msg) {
	var request pbapidevices.DeleteReq
	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
	}

	err = n.devicesService.Delete(context.Background(), request.GetID())
	if err != nil {
		n.log.Error("n.devicesService.Delete", zap.Error(err))
		n.sendError(msg.Reply, &pbapidevices.DeleteResp{Error: err.Error()})
		return
	}

	resp := pbapidevices.DeleteResp{}

	binaryResp, err := proto.Marshal(&resp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbapidevices.DeleteResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}

	if err := n.PublishUpdateEvent(); err != nil {
		n.log.Error("n.PublishUpdateEvent", zap.Error(err))
	}
}
