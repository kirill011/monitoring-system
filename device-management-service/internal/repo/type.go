package repo

import "device-management-service/internal/models"

type UpdateDeviceOpts struct {
	ID          int
	Name        *string
	DeviceType  *string
	Address     *string
	Responsible []int
}

type ReadDevicesResult struct {
	Devices []models.Device
}
