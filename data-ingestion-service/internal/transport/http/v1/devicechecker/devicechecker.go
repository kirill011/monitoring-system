package devicechecker

import (
	"data-ingestion-service/internal/models"
	"data-ingestion-service/internal/services"
	"data-ingestion-service/internal/transport/natslistener"
	"fmt"
	"net/http"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type devicechecekrHandler struct {
	deviceService services.DeviceService
	natsHandlers  *natslistener.NatsListeners

	deviceCheckPeriod int
	cron              *cron.Cron

	log *zap.Logger
}

type Config struct {
	DeviceService services.DeviceService
	NatsHandlers  *natslistener.NatsListeners

	DeviceCheckPeriod int
	Logger            *zap.Logger
}

func NewDeviceCheckerHandler(cfg *Config) *devicechecekrHandler {
	cron := cron.New(cron.WithSeconds())
	return &devicechecekrHandler{
		deviceService:     cfg.DeviceService,
		deviceCheckPeriod: cfg.DeviceCheckPeriod,
		natsHandlers:      cfg.NatsHandlers,
		cron:              cron,
		log:               cfg.Logger,
	}
}

func (dch *devicechecekrHandler) Start() {
	_, err := dch.cron.AddFunc(fmt.Sprintf("*/%d * * * * *", dch.deviceCheckPeriod),
		func() {
			ips := dch.deviceService.GetDevicesIPs()
			for _, ip := range ips {
				resp, err := http.Get(fmt.Sprintf("http://%s/healthcheck", ip))
				if err != nil {
					dch.natsHandlers.PublishSaveMessage(models.Message{
						DeviceIP:    ip,
						Message:     fmt.Sprintf("unable to connect to device %s", ip),
						MessageType: "error",
						Component:   "General",
					})
					return
				}

				if resp.StatusCode != http.StatusOK {
					dch.natsHandlers.PublishSaveMessage(models.Message{
						DeviceIP:    ip,
						Message:     fmt.Sprintf("device %s status is not OK", ip),
						MessageType: "error",
						Component:   "General",
					})
				}
			}
		},
	)
	if err != nil {
		dch.log.Error("error adding device checker to cron", zap.Error(err))
	}
}
