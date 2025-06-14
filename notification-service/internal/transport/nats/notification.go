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

	devicesUpdated        = "devices.updated"
	getResponsibleSubject = "devices.get_responsible"
	getEmailSubject       = "users.get_email"
	emailSubject          = "PROBLEM"
	defaultEmail          = "noreply@monitoring.com"
)

func (n *NatsListeners) listen() error {
	_, err := n.js.QueueSubscribe(sendNotifySubject, notifycationQueue, n.sendNotifyHandler)
	if err != nil {
		return fmt.Errorf("n.js.Subscribe("+sendNotifySubject+"): %w", err)
	}
	_, err = n.natsConn.QueueSubscribe(devicesUpdated, notifycationQueue, n.getResponsiblesHandler)

	n.getResponsiblesHandler(nil)
	return nil
}

func (n *NatsListeners) getResponsiblesHandler(msg *nats.Msg) {
	responsibleRespMsg, err := n.natsConn.Request(getResponsibleSubject, nil, n.timeout)
	if err != nil {
		n.log.Error("n.natsConn.Request", zap.Error(err),
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

	resposiblesByDeviceID := make(map[int32][]string)
	for _, responsibleResp := range responsibleResp.GetResposiblesByDeviceID() {
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
				zap.Int32("DeviceID", responsibleResp.DeviceID),
			)
			return
		}
		resposiblesByDeviceID[responsibleResp.DeviceID] = emailResp.GetEmail()
	}

	n.notificationService.SetResponsiblesByDeviceId(resposiblesByDeviceID)

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

	email := smtpsender.Email{
		Subject: notifycatioRequest.GetSubject(),
		Body:    notifycatioRequest.Text,
		From:    defaultEmail,
		To:      n.notificationService.GetResposibles(notifycatioRequest.DeviceID),
	}

	httpEmail := httpsender.Email{
		Subject: notifycatioRequest.GetSubject(),
		Body:    notifycatioRequest.Text,
		From:    defaultEmail,
		To:      n.notificationService.GetResposibles(notifycatioRequest.DeviceID),
	}

	n.log.Debug("send notification", zap.Any("email", email))

	// send notification
	// go func() {
	// 	err := n.smtpSender.Send(&email)
	// 	if err != nil {
	// 		n.log.Error("n.smtpSender.Send", zap.Error(err),
	// 			zap.Any("Email", email),
	// 		)
	// 	}
	// }()

	n.metrics.Inc()

	err := n.httpSender.Send(&httpEmail)
	if err != nil {
		n.log.Error("n.httpSender.Send", zap.Error(err),
			zap.Any("Email", email),
		)
	}
}
