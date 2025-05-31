package repo

import (
	"context"
	"data-processing-service/internal/models"
)

type Tags interface {
	BeginTx(ctx context.Context) (Tags, error)
	Commit() error
	Rollback() error

	Create(opts models.Tag) (models.Tag, error)
	Read(ctx context.Context) (ReadTagsResult, error)
	Update(ctx context.Context, opts UpdateTagsOpts) error
	Delete(ctx context.Context, id int32) error
}

type Messages interface {
	BeginTx(ctx context.Context) (Messages, error)
	Commit() error
	Rollback() error

	Create(opts models.Message) error
	GetAllByPeriod(opts MessagesGetAllByPeriodOpts) ([]models.Message, error)
	GetAllByDeviceId(deviceID int32) ([]models.Message, error)
	GetCountByMessageType(messageType string) (GetCountByMessageTypeResult, error)
	MonthReport() ([]models.MonthReportRow, error)
}
