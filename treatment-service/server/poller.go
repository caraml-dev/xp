package server

import (
	"log"
	"time"

	"github.com/caraml-dev/xp/treatment-service/config"
)

type Poller struct {
	config   config.PollerConfig
	stopChan chan struct{}
}

func NewPoller(cfg *config.PollerConfig) *Poller {
	return &Poller{
		config:   *cfg,
		stopChan: make(chan struct{}),
	}
}

func (p *Poller) Start() {
	ticker := time.NewTicker(p.config.PollInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				// Polling logic here
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
