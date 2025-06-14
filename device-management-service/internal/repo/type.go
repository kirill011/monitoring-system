package repo

import "device-management-service/internal/models"

type UpdateDeviceOpts struct {
	ID          int32
	Name        *string
	DeviceType  *string
	Address     *string
	Responsible []int32
}

type ReadDevicesResult struct {
	Devices []models.Device
}

type GetResponsibleResult struct {
	ID           int32                   `db:"id"`
	Responsibles models.SqlJsonbIntArray `db:"responsibles"`
}
