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

	devicesQueue = "devices"

	devicesUpdatedSubject = "devices.updated"
)

func (n *NatsListeners) listen() error {
	_, err := n.natsConn.QueueSubscribe(getResponsibleSubject, devicesQueue, n.gerResponsibleHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+getResponsibleSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(createDevicesSubject, devicesQueue, n.createHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+createDevicesSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(readDevicesSubject, devicesQueue, n.readHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+readDevicesSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(updateDevicesSubject, devicesQueue, n.updateHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+updateDevicesSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(deleteDevicesSubject, devicesQueue, n.deleteHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+deleteDevicesSubject+"): %w", err)
	}

	return nil
}

func (n *NatsListeners) PublishUpdateEvent() error {
	return n.natsConn.Publish(devicesUpdatedSubject, nil)
}

func (n *NatsListeners) gerResponsibleHandler(msg *nats.Msg) {
	var getResposibleReq pbdevices.GetResponsibleReq
	err := proto.Unmarshal(msg.Data, &getResposibleReq)
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

	responsibles, err := n.devicesService.GetResponsible(context.Background(), int(getResposibleReq.GetDeviceID()))
	if err != nil {
		n.log.Error("devicesService.GetResponsible",
			zap.Error(err),
			zap.Int32("DeviceID", getResposibleReq.GetDeviceID()),
		)

		if err := n.natsConn.Publish(msg.Reply, nil); err != nil {
			n.log.Error("n.natsConn.Publish", zap.Error(err))
		}
		return
	}

	getResposibleResp := pbdevices.GetResponsibleResp{
		ResponsibleID: convertIntArrayToInt32(responsibles),
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

func convertIntArrayToInt32(value []int) []int32 {
	var result []int32
	for _, v := range value {
		result = append(result, int32(v))
	}
	return result
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
			Name:        &request.Device.Name,
			DeviceType:  &request.Device.DeviceType,
			Address:     &request.Device.Address,
			Responsible: convertInt32ArrayToInt(request.Device.Responsible),
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
			Name:        *created.Name,
			DeviceType:  *created.DeviceType,
			Address:     *created.Address,
			Responsible: convertIntArrayToInt32(created.Responsible),
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

func convertInt32ArrayToInt(value []int32) []int {
	if len(value) == 0 {
		return nil
	}
	var result []int
	for _, v := range value {
		result = append(result, int(v))
	}
	return result
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
			ID:          int32(device.ID),
			Name:        *device.Name,
			DeviceType:  *device.DeviceType,
			Address:     *device.Address,
			Responsible: convertIntArrayToInt32(device.Responsible),
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
			ID:          int(request.Device.GetID()),
			Name:        name,
			DeviceType:  deviceType,
			Address:     address,
			Responsible: convertInt32ArrayToInt(request.Device.Responsible),
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

	err = n.devicesService.Delete(context.Background(), int(request.GetID()))
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
