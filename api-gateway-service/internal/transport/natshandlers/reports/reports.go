package reports

import (
	pbmessages "api-gateway-service/proto/api-gateway/messages"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type ReportsHandler struct {
	natsConn *nats.Conn
	timeout  time.Duration
}

type Config struct {
	NatsConn *nats.Conn
	Timeout  time.Duration
}

func NewReportsHandlers(cfg Config) *ReportsHandler {
	return &ReportsHandler{
		natsConn: cfg.NatsConn,
		timeout:  cfg.Timeout,
	}
}

const (
	getAllByDeviceId = "report.get_all_by_device_id"
)

func (n *ReportsHandler) PublishGetAllByDeviceId(req pbmessages.ReportGetAllByDeviceIdReq) (pbmessages.ReportGetAllByDeviceIdResp, error) {
	createReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return pbmessages.ReportGetAllByDeviceIdResp{}, fmt.Errorf("proto.Marshal: %w", err)
	}
	replyMsg, err := n.natsConn.Request(getAllByDeviceId, createReqBytes, n.timeout)
	if err != nil {
		return pbmessages.ReportGetAllByDeviceIdResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbmessages.ReportGetAllByDeviceIdResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbmessages.ReportGetAllByDeviceIdResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}
	if reply.Error != "" {
		return pbmessages.ReportGetAllByDeviceIdResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}

const (
	getAllByPeriodSubject = "report.get_all_by_period"
)

func (n *ReportsHandler) PublishGetAllByPeriod(req pbmessages.ReportGetAllByPeriodReq) (pbmessages.ReportGetAllByPeriodResp, error) {
	createReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return pbmessages.ReportGetAllByPeriodResp{}, fmt.Errorf("proto.Marshal: %w", err)
	}

	replyMsg, err := n.natsConn.Request(getAllByPeriodSubject, createReqBytes, n.timeout)
	if err != nil {
		return pbmessages.ReportGetAllByPeriodResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbmessages.ReportGetAllByPeriodResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbmessages.ReportGetAllByPeriodResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return pbmessages.ReportGetAllByPeriodResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}

const (
	getCountByMessageTypeSubject = "report.get_count_by_message_type"
)

func (n *ReportsHandler) PublishGetCountByMessageType(req pbmessages.ReportGetCountByMessageTypeReq) (pbmessages.ReportGetCountByMessageTypeResp, error) {
	createReqBytes, err := proto.Marshal(&req)
	if err != nil {
		return pbmessages.ReportGetCountByMessageTypeResp{}, fmt.Errorf("proto.Marshal: %w", err)
	}

	replyMsg, err := n.natsConn.Request(getCountByMessageTypeSubject, createReqBytes, n.timeout)
	if err != nil {
		return pbmessages.ReportGetCountByMessageTypeResp{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbmessages.ReportGetCountByMessageTypeResp
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbmessages.ReportGetCountByMessageTypeResp{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return pbmessages.ReportGetCountByMessageTypeResp{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}

const (
	getMonthReportSubject = "report.get_month_report"
)

func (n *ReportsHandler) PublishGetMonthReport() (pbmessages.MonthReport, error) {
	replyMsg, err := n.natsConn.Request(getMonthReportSubject, nil, n.timeout)
	if err != nil {
		return pbmessages.MonthReport{}, fmt.Errorf("natsConn.Request: %w", err)
	}

	var reply pbmessages.MonthReport
	err = proto.Unmarshal(replyMsg.Data, &reply)
	if err != nil {
		return pbmessages.MonthReport{}, fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if reply.Error != "" {
		return pbmessages.MonthReport{}, fmt.Errorf("reply.Error: %s", reply.Error)
	}
	return reply, nil
}
