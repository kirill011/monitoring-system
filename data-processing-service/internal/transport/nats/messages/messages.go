package messagelisteners

import (
	"data-processing-service/internal/models"
	"data-processing-service/internal/services"
	pbdevices "data-processing-service/proto/devices"
	pbmessages "data-processing-service/proto/messages"
	pbnotification "data-processing-service/proto/notification"
	"time"

	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	saveMessageSubject = "msg.save"
	sendNotifySubject  = "notify.send"

	readDevicesSubject    = "devices.read"
	devicesUpdatedSubject = "devices.updated"

	reportGetAllByPeriod        = "report.get_all_by_period"
	reportGetAllByDeviceId      = "report.get_all_by_device_id"
	reportGetCountByMessageType = "report.get_count_by_message_type"
	reportGetMonthReport        = "report.get_month_report"

	dataProcessingQueue = "data-processing"
)

type NatsListeners struct {
	natsConn        *nats.Conn
	messagesService services.Messages
	timeout         time.Duration
	log             *zap.Logger
}

type Config struct {
	NatsConn        *nats.Conn
	MessagesService services.Messages
	Timeout         time.Duration
	Log             *zap.Logger
}

func NewListener(cfg Config) *NatsListeners {
	return &NatsListeners{
		natsConn:        cfg.NatsConn,
		messagesService: cfg.MessagesService,
		log:             cfg.Log,
		timeout:         cfg.Timeout,
	}
}

func (n *NatsListeners) Listen() error {
	_, err := n.natsConn.QueueSubscribe(saveMessageSubject, dataProcessingQueue, n.createHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+saveMessageSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(devicesUpdatedSubject, dataProcessingQueue, n.updateDevicesHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+devicesUpdatedSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(reportGetAllByPeriod, dataProcessingQueue, n.getAllByPeriodHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+reportGetAllByPeriod+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(reportGetAllByDeviceId, dataProcessingQueue, n.getAllByDeviceIdHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+reportGetAllByDeviceId+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(reportGetCountByMessageType, dataProcessingQueue, n.getCountByMessageTypeHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+reportGetCountByMessageType+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(reportGetMonthReport, dataProcessingQueue, n.getMonthReport)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+reportGetMonthReport+"): %w", err)
	}

	n.updateDevicesHandler(nil)

	return nil
}

func (n *NatsListeners) updateDevicesHandler(msg *nats.Msg) {
	result, err := n.natsConn.Request(readDevicesSubject, nil, n.timeout)
	if err != nil {
		n.log.Error("n.natsConn.Request", zap.Error(err))
		return
	}

	var response pbdevices.ReadResp
	err = proto.Unmarshal(result.Data, &response)
	if err != nil {
		n.log.Error("proto.Unmarshal", zap.Error(err))
		return
	}

	if response.Error != "" {
		n.log.Error("response.Error", zap.String("Error", response.Error))
		return
	}

	var devices []models.Device
	for _, device := range response.Devices {
		devices = append(devices, models.Device{
			ID:          device.ID,
			Name:        device.Name,
			DeviceType:  device.DeviceType,
			Address:     device.Address,
			Responsible: device.Responsible,
		})
	}

	n.messagesService.SetDevices(devices)
}

func (n *NatsListeners) createHandler(msg *nats.Msg) {
	var request pbmessages.MessageSave
	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
	}

	notify, needNotify, err := n.messagesService.Create(models.Message{
		DeviceId:    request.DeviceID,
		Message:     request.Message,
		MessageType: request.MessageType,
		Component:   request.Component,
	})
	if err != nil {
		n.log.Error("n.messagesService.Create", zap.Error(err))
		return
	}

	if !needNotify {
		return
	}

	binaryResp, err := proto.Marshal(&pbnotification.SendNotifyReq{
		DeviceID: request.DeviceID,
		Text:     notify.Text,
		Subject:  notify.Subject,
	})
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		return
	}

	if err := n.natsConn.Publish(sendNotifySubject, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return

	}
}

func (n *NatsListeners) getAllByPeriodHandler(msg *nats.Msg) {
	var request pbmessages.ReportGetAllByPeriodReq
	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)

		n.sendError(msg.Reply, &pbmessages.ReportGetAllByPeriodResp{Error: err.Error()})
		return
	}

	report, err := n.messagesService.GetAllByPeriod(services.MessagesGetAllByPeriodOpts{
		StartTime: request.StartTime.AsTime(),
		EndTime:   request.EndTime.AsTime(),
	})
	if err != nil {
		n.log.Error("n.messagesService.GetAllByPeriod", zap.Error(err))
		n.sendError(msg.Reply, &pbmessages.ReportGetAllByPeriodResp{Error: err.Error()})
		return
	}

	binaryResp, err := proto.Marshal(&pbmessages.ReportGetAllByPeriodResp{
		Report: convertGetAllByPeriodToProto(report),
	})
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbmessages.ReportGetAllByPeriodResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}

func convertGetAllByPeriodToProto(report []services.ReportGetAllByPeriod) []*pbmessages.ReportGetAllByPeriod {
	resp := []*pbmessages.ReportGetAllByPeriod{}
	for _, item := range report {
		resp = append(resp, &pbmessages.ReportGetAllByPeriod{
			DeviceID:    item.DeviceID,
			Name:        item.Name,
			DeviceType:  item.DeviceType,
			Address:     item.Address,
			Responsible: item.Responsible,
			GotAt:       timestamppb.New(item.GotAt),
			Message:     item.Message,
			MessageType: item.MessageType,
		})
	}
	return resp
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

func (n *NatsListeners) getAllByDeviceIdHandler(msg *nats.Msg) {
	var request pbmessages.ReportGetAllByDeviceIdReq
	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)

		n.sendError(msg.Reply, &pbmessages.ReportGetAllByDeviceIdResp{Error: err.Error()})
		return
	}

	report, err := n.messagesService.GetAllByDeviceId(request.GetDeviceId())
	if err != nil {
		n.log.Error("n.messagesService.GetAllByDeviceId", zap.Error(err))
		n.sendError(msg.Reply, &pbmessages.ReportGetAllByDeviceIdResp{Error: err.Error()})
		return
	}

	binaryResp, err := proto.Marshal(&pbmessages.ReportGetAllByDeviceIdResp{
		Report: convertGetAllByDeviceIdToProto(report),
	})
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbmessages.ReportGetAllByDeviceIdResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}

func convertGetAllByDeviceIdToProto(report []services.ReportGetAllByDeviceId) []*pbmessages.ReportGetAllByDeviceId {
	resp := []*pbmessages.ReportGetAllByDeviceId{}
	for _, item := range report {
		resp = append(resp, &pbmessages.ReportGetAllByDeviceId{
			DeviceID:    item.DeviceID,
			Name:        item.Name,
			DeviceType:  item.DeviceType,
			Address:     item.Address,
			Responsible: item.Responsible,
			GotAt:       timestamppb.New(item.GotAt),
			Message:     item.Message,
			MessageType: item.MessageType,
		})
	}
	return resp
}

func (n *NatsListeners) getCountByMessageTypeHandler(msg *nats.Msg) {
	var request pbmessages.ReportGetCountByMessageTypeReq
	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)

		n.sendError(msg.Reply, &pbmessages.ReportGetCountByMessageTypeResp{Error: err.Error()})
		return
	}

	report, err := n.messagesService.GetCountByMessageType(request.GetMessageType())
	if err != nil {
		n.log.Error("n.messagesService.GetCountByMessageType", zap.Error(err))
		n.sendError(msg.Reply, &pbmessages.ReportGetCountByMessageTypeResp{Error: err.Error()})
		return
	}

	binaryResp, err := proto.Marshal(&pbmessages.ReportGetCountByMessageTypeResp{
		Report: convertGetCountByMessageTypeToProto(report),
	})
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbmessages.ReportGetCountByMessageTypeResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}

func convertGetCountByMessageTypeToProto(report []services.ReportGetCountByMessageType) []*pbmessages.ReportGetCountByMessageType {
	resp := []*pbmessages.ReportGetCountByMessageType{}
	for _, item := range report {
		resp = append(resp, &pbmessages.ReportGetCountByMessageType{
			DeviceID:    item.DeviceID,
			Name:        item.Name,
			DeviceType:  item.DeviceType,
			Address:     item.Address,
			Responsible: item.Responsible,
			Count:       item.Count,
		})
	}
	return resp
}

func (n *NatsListeners) getMonthReport(msg *nats.Msg) {
	report, err := n.messagesService.MonthReport()
	if err != nil {
		n.log.Error("n.messagesService.MonthReport", zap.Error(err))
		n.sendError(msg.Reply, &pbmessages.MonthReport{Error: err.Error()})
		return
	}

	binaryResp, err := proto.Marshal(&pbmessages.MonthReport{
		MonthReport: convertMonthReportToProto(report),
	})
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbmessages.MonthReport{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}

func convertMonthReportToProto(report []models.MonthReportRow) []*pbmessages.MonthReportRow {
	var resp []*pbmessages.MonthReportRow
	for _, item := range report {
		resp = append(resp, &pbmessages.MonthReportRow{
			DeviceId:            item.DeviceID,
			MessageType:         item.MessageType,
			ActiveDays:          item.ActiveDays,
			TotalMessages:       item.TotalMessages,
			AvgDailyMessages:    item.AvgDailyMessages,
			MaxDailyMessages:    item.MaxDailyMessages,
			MedianDailyMessages: item.MedianDailyMessages,
			TotalCritical:       item.TotalCritical,
			MaxDailyCritical:    item.MaxDailyCritical,
			MaxDailyComponents:  item.MaxDailyComponents,
			MostActiveComponent: item.MostActiveComponent.String,
			FirstCriticalTime:   timestamppb.New(item.FirstCriticalTime.Time),
			LastCriticalTime:    timestamppb.New(item.LastCriticalTime.Time),
			AvgCriticalInterval: item.AvgCriticalIntervalSec.Float64,
			CriticalPercentage:  item.CriticalPercentage,
			OverallVolumeRank:   item.OverallVolumeRank,
		})
	}
	return resp
}
