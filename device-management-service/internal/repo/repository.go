package repo

import (
	"context"
	"device-management-service/internal/models"
)

type Devices interface {
	BeginTx(ctx context.Context) (Devices, error)
	Commit() error
	Rollback() error

	Create(opts models.Device) (models.Device, error)
	Read(ctx context.Context) (ReadDevicesResult, error)
	Update(ctx context.Context, opts UpdateDeviceOpts) error
	Delete(ctx context.Context, id int32) error
	GetResponsible(ctx context.Context) ([]GetResponsibleResult, error)
}
