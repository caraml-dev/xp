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
	localStorage *models.LocalStorage
	stopChannel  chan struct{}
	stopOnce     sync.Once
	waitGroup    sync.WaitGroup
}

// NewPoller creates a new Poller instance with the given configuration and local storage.
// pollerConfig: configuration for the poller
// localStorage: local storage to be used by the poller
func NewPoller(pollerConfig *config.PollerConfig, localStorage *models.LocalStorage) *Poller {
	return &Poller{
		pollerConfig: *pollerConfig,
		localStorage: localStorage,
		stopChannel:  make(chan struct{}),
	}
}

func (p *Poller) Start() {
	var startOnce sync.Once
	startOnce.Do(func() {
		ticker := time.NewTicker(p.pollerConfig.PollInterval)
		p.waitGroup.Add(1)
		go func() {
			defer p.waitGroup.Done()
			for {
				select {
				case <-ticker.C:
					err := p.localStorage.Init()
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
	})
}

func (p *Poller) Stop() {
	p.stopOnce.Do(func() {
		select {
		case <-p.stopChannel:
			// Already closed
		default:
			close(p.stopChannel)
			p.waitGroup.Wait()
		}
	})
}
