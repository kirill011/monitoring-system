package natslisteners

import (
	"fmt"
	smtpsender "notification-service/internal/transport/smtp"
	pbgetmail "notification-service/proto/getmail"
	pbgetresposible "notification-service/proto/getresponsible"
	pbsendnotify "notification-service/proto/sendnotify"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	sendNotifySubject = "notify.send"
	notifycationQueue = "notification"

	getResponsibleSubject = "devices.get_responsible"
	getEmailSubject       = "users.get_email"

	emailSubject = "PROBLEM"
	defaultEmail = "noreply@monitoring.com"
)

func (n *NatsListeners) listen() error {
	_, err := n.natsConn.QueueSubscribe(sendNotifySubject, notifycationQueue, n.sendNotifyHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+sendNotifySubject+"): %w", err)
	}

	return nil
}

func (n *NatsListeners) sendNotifyHandler(msg *nats.Msg) {
	// Get Notifycation
	var notifycatioRequest pbsendnotify.SendNotifyReq

	if err := proto.Unmarshal(msg.Data, &notifycatioRequest); err != nil {
		n.log.Error("proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
		return
	}

	// Get Responsible
	responsibleReq := pbgetresposible.GetResponsibleReq{
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

	var responsibleResp pbgetresposible.GetResponsibleResp
	if err := proto.Unmarshal(responsibleRespMsg.Data, &responsibleResp); err != nil {
		n.log.Error("proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", getResponsibleSubject),
		)
		return
	}

	// Get Email
	emailReq := pbgetmail.GetEmailReq{
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

	var emailResp pbgetmail.GetEmailResp
	if err := proto.Unmarshal(emailRespMsg.Data, &emailResp); err != nil {
		n.log.Error("proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", getEmailSubject),
		)
		return
	}

	email := smtpsender.Email{
		Subject: emailSubject,
		Body:    notifycatioRequest.Text,
		From:    defaultEmail,
		To:      emailResp.GetEmail(),
	}

	// send notification
	err = n.smtpSender.Send(&email)
	if err != nil {
		n.log.Error("n.smtpSender.Send", zap.Error(err),
			zap.Any("Email", email),
		)
	}
}
