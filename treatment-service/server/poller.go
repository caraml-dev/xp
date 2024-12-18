package server

import (
	"log"
	"time"

	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type Poller struct {
	config   config.PollerConfig
	storage  *models.LocalStorage
	stopChan chan struct{}
}

func NewPoller(cfg *config.PollerConfig, storage *models.LocalStorage) *Poller {
	return &Poller{
		config:   *cfg,
		storage:  storage,
		stopChan: make(chan struct{}),
	}
}

func (p *Poller) Start() {
	ticker := time.NewTicker(p.config.PollInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				p.storage.Init()
				log.Println("Polling...")
			case <-p.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *Poller) Stop() {
	close(p.stopChan)
}
