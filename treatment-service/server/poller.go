package server

import (
	"log"
	"time"

	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type Poller struct {
	pollerConfig config.ManagementServicePollerConfig
	localStorage *models.LocalStorage
	stopChannel  chan struct{}
}

// NewPoller creates a new Poller instance with the given configuration and local storage.
// pollerConfig: configuration for the poller
// localStorage: local storage to be used by the poller
func NewPoller(pollerConfig config.ManagementServicePollerConfig, localStorage *models.LocalStorage) *Poller {
	return &Poller{
		pollerConfig: pollerConfig,
		localStorage: localStorage,
		stopChannel:  make(chan struct{}),
	}
}

func (p *Poller) Start() {
	ticker := time.NewTicker(p.pollerConfig.PollInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := p.Refresh()
				log.Printf("Polling at %v with interval %v", time.Now(), p.pollerConfig.PollInterval)
				if err != nil {
					log.Printf("Error updating local storage: %v", err)
					continue
				}
			case <-p.stopChannel:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *Poller) Stop() {
	close(p.stopChannel)
}

func (p *Poller) Refresh() error {
	err := p.localStorage.Init()
	return err
}
