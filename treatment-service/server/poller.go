package server

import (
	"log"
	"sync"
	"time"

	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type Poller struct {
	pollerConfig config.PollerConfig
	storage      *models.LocalStorage
	stopChan     chan struct{}
	stopOnce     sync.Once
}

func NewPoller(pollerConfig *config.PollerConfig, storage *models.LocalStorage) *Poller {
	return &Poller{
		pollerConfig: *pollerConfig,
		storage:      storage,
		stopChan:     make(chan struct{}),
	}
}

func (p *Poller) Start() {
	ticker := time.NewTicker(p.pollerConfig.PollInterval)
	go p.poll(ticker)
}

func (p *Poller) poll(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			err := p.storage.Init()
			log.Printf("Polling at %v with interval %v", time.Now(), p.pollerConfig.PollInterval)
			if err != nil {
				log.Printf("Error updating local storage: %v", err)
				return
			}
		case <-p.stopChan:
			ticker.Stop()
			return
		}
	}
}

func (p *Poller) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopChan)
	})
}
