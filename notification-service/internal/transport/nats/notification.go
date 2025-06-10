package natslisteners

import (
	"fmt"
	httpsender "notification-service/internal/transport/http"
	smtpsender "notification-service/internal/transport/smtp"
	pbdevices "notification-service/proto/devices"
	pbnotification "notification-service/proto/notification"
	pbusers "notification-service/proto/users"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	sendNotifySubject = "monitoring.notify.send"
	notifycationQueue = "notification"

	getResponsibleSubject = "devices.get_responsible"
	getEmailSubject       = "users.get_email"

	emailSubject = "PROBLEM"
	defaultEmail = "noreply@monitoring.com"
)

func (n *NatsListeners) listen() error {
	_, err := n.js.QueueSubscribe(sendNotifySubject, notifycationQueue, n.sendNotifyHandler)
	if err != nil {
		return fmt.Errorf("n.js.Subscribe("+sendNotifySubject+"): %w", err)
	}

	return nil
}

func (n *NatsListeners) sendNotifyHandler(msg *nats.Msg) {
	// Get Notifycation
	var notifycatioRequest pbnotification.SendNotifyReq

	if err := proto.Unmarshal(msg.Data, &notifycatioRequest); err != nil {
		n.log.Error("proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
		return
	}

	// Get Responsible
	responsibleReq := pbdevices.GetResponsibleReq{
		DeviceID: notifycatioRequest.DeviceID,
	}

	responsibleReqBytes, err := proto.Marshal(&responsibleReq)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err), zap.Any("variable", responsibleReq))
		return
	}

	responsibleRespMsg, err := n.natsConn.Request(getResponsibleSubject, responsibleReqBytes, n.timeout)
	if err != nil {
		n.log.Error("n.natsConn.Request", zap.Error(err),
			zap.Binary("Data", responsibleReqBytes),
			zap.String("Subject", getResponsibleSubject),
		)
		return
	}

	var responsibleResp pbdevices.GetResponsibleResp
	if err := proto.Unmarshal(responsibleRespMsg.Data, &responsibleResp); err != nil {
		n.log.Error("proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", getResponsibleSubject),
		)
		return
	}

	if len(responsibleResp.GetResponsibleID()) == 0 {
		n.log.Warn("responsible not found",
			zap.Any("responsibleID", responsibleResp.GetResponsibleID()),
			zap.Int32("DeviceID", notifycatioRequest.DeviceID),
		)
		return
	}

	// Get Email
	emailReq := pbusers.GetEmailReq{
		UserID: responsibleResp.GetResponsibleID(),
	}

	emailReqBytes, err := proto.Marshal(&emailReq)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err), zap.Any("variable", emailReq))
		return
	}

	emailRespMsg, err := n.natsConn.Request(getEmailSubject, emailReqBytes, n.timeout)
	if err != nil {
		n.log.Error("n.natsConn.Request", zap.Error(err),
			zap.Binary("Data", emailReqBytes),
			zap.String("Subject", getEmailSubject),
		)
		return
	}

	var emailResp pbusers.GetEmailResp
	if err := proto.Unmarshal(emailRespMsg.Data, &emailResp); err != nil {
		n.log.Error("proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", getEmailSubject),
		)
		return
	}

	if len(emailResp.GetEmail()) == 0 {
		n.log.Warn("responsible not found",
			zap.Any("responsibleID", emailResp.GetEmail()),
			zap.Int32("DeviceID", notifycatioRequest.DeviceID),
		)
		return
	}

	email := smtpsender.Email{
		Subject: notifycatioRequest.GetSubject(),
		Body:    notifycatioRequest.Text,
		From:    defaultEmail,
		To:      emailResp.GetEmail(),
	}

	httpEmail := httpsender.Email{
		Subject: notifycatioRequest.GetSubject(),
		Body:    notifycatioRequest.Text,
		From:    defaultEmail,
		To:      emailResp.GetEmail(),
	}

	n.log.Debug("send notification", zap.Any("email", email))

	// send notification
	err = n.smtpSender.Send(&email)
	if err != nil {
		n.log.Error("n.smtpSender.Send", zap.Error(err),
			zap.Any("Email", email),
		)
	}

	err = n.httpSender.Send(&httpEmail)
	if err != nil {
		n.log.Error("n.httpSender.Send", zap.Error(err),
			zap.Any("Email", email),
		)
	}
}
