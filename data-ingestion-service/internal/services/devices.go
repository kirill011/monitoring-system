package services

import (
	"data-ingestion-service/internal/models"
	"sync"
)

type DeviceService struct {
}

type Config struct{}

func NewDeviceService(cfg Config) DeviceService {
	return DeviceService{}
}

var (
	devicesMutex sync.Mutex
	devicesByIp  = map[string]models.Device{}
)

func (ds *DeviceService) UpdateDevices(devices []models.Device) {
	devicesMutex.Lock()
	defer devicesMutex.Unlock()

	devicesByIp = make(map[string]models.Device, len(devices))
	for _, d := range devices {
		devicesByIp[d.Address] = d
	}
}

func (ds *DeviceService) GetDeviceIDByIp(address string) (int32, bool) {
	device, ok := devicesByIp[address]
	return device.ID, ok
}

func (ds *DeviceService) GetDevicesIPs() []string {
	res := make([]string, 0, len(devicesByIp))
	for ip := range devicesByIp {
		res = append(res, ip)
	}

	return res
}
