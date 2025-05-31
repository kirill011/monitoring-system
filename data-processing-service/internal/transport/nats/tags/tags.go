package tagslistener

import (
	"context"
	"data-processing-service/internal/models"
	"data-processing-service/internal/services"
	pbapidevices "data-processing-service/proto/api-gateway/tags"

	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	createTagsSubject = "tags.create"
	readTagsSubject   = "tags.read"
	updateTagsSubject = "tags.update"
	deleteTagsSubject = "tags.delete"

	readDevicesSubject = "devices.read"

	tagsQueue = "devices"
)

type NatsListeners struct {
	natsConn    *nats.Conn
	tagsService services.Tags
	log         *zap.Logger
}

type Config struct {
	NatsConn    *nats.Conn
	TagsService services.Tags
	Log         *zap.Logger
}

func NewListener(cfg Config) *NatsListeners {
	return &NatsListeners{
		natsConn:    cfg.NatsConn,
		tagsService: cfg.TagsService,
		log:         cfg.Log,
	}
}

func (n *NatsListeners) Listen() error {
	_, err := n.natsConn.QueueSubscribe(createTagsSubject, tagsQueue, n.createHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+createTagsSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(readTagsSubject, tagsQueue, n.readHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+readTagsSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(updateTagsSubject, tagsQueue, n.updateHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+updateTagsSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(deleteTagsSubject, tagsQueue, n.deleteHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+deleteTagsSubject+"): %w", err)
	}

	return nil
}

func (n *NatsListeners) PublishReadDevices() {

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

	created, err := n.tagsService.Create(context.Background(),
		models.Tag{
			Name:          request.Tag.Name,
			DeviceId:      request.Tag.DeviceID,
			Regexp:        request.Tag.Regexp,
			CompareType:   request.Tag.CompareType,
			Value:         request.Tag.Value,
			ArrayIndex:    request.Tag.ArrayIndex,
			Subject:       request.Tag.Subject,
			SeverityLevel: request.Tag.SeverityLevel,
		})
	if err != nil {
		n.log.Error("n.tagsService.Create", zap.Error(err))
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
		Created: &pbapidevices.Tag{
			ID:            created.ID,
			Name:          created.Name,
			DeviceID:      created.DeviceId,
			Regexp:        created.Regexp,
			CompareType:   created.CompareType,
			Value:         created.Value,
			ArrayIndex:    created.ArrayIndex,
			Subject:       created.Subject,
			SeverityLevel: created.SeverityLevel,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
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

func (n *NatsListeners) readHandler(msg *nats.Msg) {
	readResult, err := n.tagsService.Read(context.Background())
	if err != nil {
		n.log.Error("n.tagsService.Read",
			zap.Error(err),
		)

		n.sendError(msg.Reply, &pbapidevices.ReadResp{Error: err.Error()})
		return
	}

	readTagsResp := pbapidevices.ReadResp{
		Tags: convertTagsToProtoTags(readResult.Tags),
	}

	readTagsRespBytes, err := proto.Marshal(&readTagsResp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err), zap.Any("variable", readTagsRespBytes))

		n.sendError(msg.Reply, &pbapidevices.ReadResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, readTagsRespBytes); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
	}
}

func convertTagsToProtoTags(devices []models.Tag) []*pbapidevices.Tag {
	var result []*pbapidevices.Tag
	for _, device := range devices {
		var createdAt *timestamppb.Timestamp
		if device.CreatedAt != nil {
			createdAt = timestamppb.New(*device.CreatedAt)
		}

		var updatedAt *timestamppb.Timestamp
		if device.UpdatedAt != nil {
			updatedAt = timestamppb.New(*device.UpdatedAt)
		}

		result = append(result, &pbapidevices.Tag{
			ID:            device.ID,
			Name:          device.Name,
			DeviceID:      device.DeviceId,
			Regexp:        device.Regexp,
			CompareType:   device.CompareType,
			Value:         device.Value,
			ArrayIndex:    device.ArrayIndex,
			Subject:       device.Subject,
			SeverityLevel: device.SeverityLevel,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
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
	if request.Tag.GetName() != "" {
		name = &request.Tag.Name
	}
	var deviceId *int32
	if request.Tag.GetDeviceID() != 0 {
		deviceId = &request.Tag.DeviceID
	}
	var regexp *string
	if request.Tag.GetRegexp() != "" {
		regexp = &request.Tag.Regexp
	}
	var compareType *string
	if request.Tag.GetCompareType() != "" {
		compareType = &request.Tag.CompareType
	}
	var value *string
	if request.Tag.GetValue() != "" {
		value = &request.Tag.Value
	}
	var arrayIndex *int32
	if request.Tag.GetArrayIndex() != 0 {
		arrayIndex = &request.Tag.ArrayIndex
	}

	var subject *string
	if request.Tag.GetSubject() != "" {
		subject = &request.Tag.Subject
	}

	var SeverityLevel *string
	if request.Tag.GetSeverityLevel() != "" {
		SeverityLevel = &request.Tag.SeverityLevel
	}

	err = n.tagsService.Update(context.Background(),
		services.UpdateParams{
			ID:            request.Tag.GetID(),
			Name:          name,
			DeviceId:      deviceId,
			Regexp:        regexp,
			CompareType:   compareType,
			Value:         value,
			ArrayIndex:    arrayIndex,
			Subject:       subject,
			SeverityLevel: SeverityLevel,
		})
	if err != nil {
		n.log.Error("n.tagsService.Update", zap.Error(err))
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

	err = n.tagsService.Delete(context.Background(), request.GetID())
	if err != nil {
		n.log.Error("n.tagsService.Delete", zap.Error(err))
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
}
