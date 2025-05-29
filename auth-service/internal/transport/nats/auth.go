package natslisteners

import (
	"auth-service/internal/models"
	"auth-service/internal/services"
	pbapiusers "auth-service/proto/api-gateway/users"
	pbusers "auth-service/proto/users"
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	localID = "localID"

	getEmailSubject   = "users.get_email"
	usersQueue        = "users"
	authCreateSubject = "users.create"
	authReadSubject   = "users.read"
	authUpdateSubject = "users.update"
	authDeleteSubject = "users.delete"
	authSubject       = "users.auth"
)

func (n *NatsListeners) listen() error {
	_, err := n.natsConn.QueueSubscribe(getEmailSubject, usersQueue, n.getEmailHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+getEmailSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(authSubject, usersQueue, n.authHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+authSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(authCreateSubject, usersQueue, n.createHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+authCreateSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(authReadSubject, usersQueue, n.readHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+authReadSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(authUpdateSubject, usersQueue, n.updateHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+authUpdateSubject+"): %w", err)
	}

	_, err = n.natsConn.QueueSubscribe(authDeleteSubject, usersQueue, n.deleteHandler)
	if err != nil {
		return fmt.Errorf("n.natsConn.Subscribe("+authDeleteSubject+"): %w", err)
	}

	return nil
}

func (n *NatsListeners) getEmailHandler(msg *nats.Msg) {
	var request pbusers.GetEmailReq

	err := proto.Unmarshal(msg.Data, &request)
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

	email, err := n.authService.GetEmailsByIDs(context.Background(), request.GetUserID())
	if err != nil {
		n.log.Error("n.authService.GetEmailsByIDs", zap.Error(err))

		if err := n.natsConn.Publish(msg.Reply, nil); err != nil {
			n.log.Error("n.natsConn.Publish", zap.Error(err))
		}
		return
	}

	response := pbusers.GetEmailResp{
		Email: email,
	}

	responseBytes, err := proto.Marshal(&response)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))

		if err := n.natsConn.Publish(msg.Reply, nil); err != nil {
			n.log.Error("n.natsConn.Publish", zap.Error(err))
		}
		return
	}

	if err := n.natsConn.Publish(msg.Reply, responseBytes); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}

func (n *NatsListeners) authHandler(msg *nats.Msg) {
	var request pbapiusers.AuthReq

	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
	}

	userID, err := n.authService.Authorize(context.Background(), services.AuthorizeParams{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		n.log.Error("n.authService.Authorize", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.AuthResp{Error: err.Error()})
		return
	}

	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		localID: userID,
		"exp":   time.Now().Add(n.tokenLifeTime).Unix(),
	})

	token, err := jwt.SignedString([]byte(n.jwtKey))
	if err != nil {
		n.log.Error("jwt.SignedString", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.AuthResp{Error: err.Error()})
		return
	}

	resp := pbapiusers.AuthResp{
		Token: token,
	}

	binaryResp, err := proto.Marshal(&resp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.AuthResp{Error: err.Error()})
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

func (n *NatsListeners) createHandler(msg *nats.Msg) {
	var request pbapiusers.CreateReq

	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
	}

	created, err := n.authService.Create(context.Background(), services.CreateUserParams{
		Name:     request.User.Name,
		Email:    request.User.Email,
		Password: request.User.Password,
	})
	if err != nil {
		n.log.Error("n.authService.Authorize", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.CreateResp{Error: err.Error()})
		return
	}

	resp := pbapiusers.CreateResp{
		Created: &pbapiusers.User{
			ID:    int32(created.ID),
			Email: created.Email,
			Name:  created.Name,
		},
	}

	binaryResp, err := proto.Marshal(&resp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.CreateResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}

func (n *NatsListeners) readHandler(msg *nats.Msg) {
	data, err := n.authService.Read(context.Background())
	if err != nil {
		n.log.Error("n.authService.Authorize", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.ReadResp{Error: err.Error()})
		return
	}

	resp := pbapiusers.ReadResp{
		Users: convertUsersToProtoUsers(data.Users),
	}

	binaryResp, err := proto.Marshal(&resp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.ReadResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}

func convertUsersToProtoUsers(users []models.User) []*pbapiusers.User {
	var result []*pbapiusers.User

	for _, user := range users {
		var createdAt *timestamppb.Timestamp
		if user.CreatedAt != nil {
			createdAt = timestamppb.New(*user.CreatedAt)
		}

		var updatedAt *timestamppb.Timestamp
		if user.UpdatedAt != nil {
			updatedAt = timestamppb.New(*user.UpdatedAt)
		}

		result = append(result, &pbapiusers.User{
			ID:        int32(user.ID),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}
	return result
}

func (n *NatsListeners) updateHandler(msg *nats.Msg) {
	var request pbapiusers.UpdateReq

	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
	}

	n.log.Debug("updateHandler", zap.String("request", request.String()))

	err = n.authService.Update(context.Background(), services.UpdateUsersParams{
		ID:       int(request.User.GetID()),
		Name:     &request.User.Name,
		Email:    &request.User.Email,
		Password: &request.User.Password,
	})
	if err != nil {
		n.log.Error("n.authService.Authorize", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.UpdateResp{Error: err.Error()})
		return
	}

	resp := pbapiusers.UpdateResp{}

	binaryResp, err := proto.Marshal(&resp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.UpdateResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}

func (n *NatsListeners) deleteHandler(msg *nats.Msg) {
	var request pbapiusers.DeleteReq

	err := proto.Unmarshal(msg.Data, &request)
	if err != nil {
		n.log.Error(
			"proto.Unmarshal",
			zap.Error(err),
			zap.Binary("Data", msg.Data),
			zap.String("Subject", msg.Subject),
		)
	}

	err = n.authService.Delete(context.Background(), int(request.GetID()))
	if err != nil {
		n.log.Error("n.authService.Authorize", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.DeleteResp{Error: err.Error()})
		return
	}

	resp := pbapiusers.DeleteResp{}

	binaryResp, err := proto.Marshal(&resp)
	if err != nil {
		n.log.Error("proto.Marshal", zap.Error(err))
		n.sendError(msg.Reply, &pbapiusers.DeleteResp{Error: err.Error()})
		return
	}

	if err := n.natsConn.Publish(msg.Reply, binaryResp); err != nil {
		n.log.Error("n.natsConn.Publish", zap.Error(err))
		return
	}
}
