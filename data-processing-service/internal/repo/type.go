package repo

import (
	"data-processing-service/internal/models"
	"time"
)

type UpdateTagsOpts struct {
	ID            int32
	Name          *string
	DeviceId      *int32
	Regexp        *string
	CompareType   *string
	Value         *string
	ArrayIndex    *int32
	Subject       *string
	SeverityLevel *string
}

type ReadTagsResult struct {
	Tags []models.Tag
}

type MessagesGetAllByPeriodOpts struct {
	StartTime time.Time
	EndTime   time.Time
}

type GetCountByMessageTypeResult struct {
	Count []models.CountByDeviceID
}
