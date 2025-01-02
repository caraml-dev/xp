package server

import (
	"log"
	"sync"
	"time"

	"github.com/caraml-dev/xp/treatment-service/config"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type Poller struct {
	pollerConfig config.ManagementServicePollerConfig
	localStorage *models.LocalStorage
	control      PollerControl
}

type PollerControl struct {
	stopChannel chan struct{}
	startOnce   sync.Once
	stopOnce    sync.Once
	waitGroup   sync.WaitGroup
}

// NewPoller creates a new Poller instance with the given configuration and local storage.
// pollerConfig: configuration for the poller
// localStorage: local storage to be used by the poller
func NewPoller(pollerConfig config.ManagementServicePollerConfig, localStorage *models.LocalStorage) *Poller {
	return &Poller{
		pollerConfig: pollerConfig,
		localStorage: localStorage,
		control: PollerControl{
			stopChannel: make(chan struct{}),
		},
	}
}

func (p *Poller) Start() {
	p.control.startOnce.Do(func() {
		ticker := time.NewTicker(p.pollerConfig.PollInterval)
		p.control.waitGroup.Add(1)
		go func() {
			defer p.control.waitGroup.Done()
			for {
				select {
				case <-ticker.C:
					err := p.Refresh()
					log.Printf("Polling at %v with interval %v", time.Now(), p.pollerConfig.PollInterval)
					if err != nil {
						log.Printf("Error updating local storage: %v", err)
						continue
					}
				case <-p.control.stopChannel:
					ticker.Stop()
					return
				}
			}
		}()
	})
}

func (p *Poller) Stop() {
	p.control.stopOnce.Do(func() {
		close(p.control.stopChannel)
		p.control.waitGroup.Wait()
	})
}

func (p *Poller) Refresh() error {
	err := p.localStorage.Init()
	return err
}
