package main

import (
	"testing"
	"time"

	"github.com/NouamaneTazi/iseeu/internal/config"
	"github.com/NouamaneTazi/iseeu/internal/inspect"
	"github.com/NouamaneTazi/iseeu/internal/metrics"
)

func initConfig() {
	config.ShortUIRefreshInterval = 2 * time.Second
	config.LongUIRefreshInterval = 10 * time.Second
	config.ShortStatsHistoryInterval = 10 * time.Second
	config.LongStatsHistoryInterval = 60 * time.Second
	config.WebsiteAlertInterval = 10 * time.Second
}
func TestAlerting(t *testing.T) {
	initConfig()
	reportc := make(chan *inspect.Report, 5)
	pollingInterval := 1 * time.Second
	errReport := &inspect.Report{
		Url:               "testurl",
		PollingInterval:   pollingInterval,
		StatusCode:        0, // any status that is different from 200
		ConnectDuration:   -1,
		FirstByteDuration: -1,
	}
	availableReport := &inspect.Report{
		Url:               "testurl",
		PollingInterval:   pollingInterval,
		StatusCode:        200,
		ConnectDuration:   10,
		FirstByteDuration: 5,
	}

	met := metrics.NewMetrics(reportc, pollingInterval)
	go met.ListenAndProcess() // listens for incoming reports
	alert := met.Alert

	if alert.WebsiteHasRecovered {
		t.Error("Website must be initialized as not recovered")
	}
	if alert.WebsiteWasDown {
		t.Error("Website must be initialized as not down")
	}

	for loop := 0; loop < 2; loop++ {
		// website doing BAD
		for i := 0; i < 10; i++ {
			reportc <- errReport
			time.Sleep(time.Millisecond) // Simulate pollinginterval, and forces writer to start before reader
			met.Mu.RLock()
			if alert.WebsiteHasRecovered {
				t.Error("Phase 1: Website hasn't recovered yet")
			}
			if !alert.WebsiteWasDown {
				t.Error("Phase 1: Website was down")
			}
			met.Mu.RUnlock()
		}

		// website recovering
		for i := 0; i < 7; i++ {
			reportc <- availableReport
			time.Sleep(time.Millisecond) // Simulate pollinginterval, and forces writer to start before reader
			met.Mu.RLock()
			if alert.WebsiteHasRecovered {
				t.Error("Phase 2: Website hasn't recovered yet")
			}
			if alert.WebsiteWasDown {
				t.Error("Phase 2: Website wasn't down")
			}
			met.Mu.RUnlock()
		}

		// website has recovered
		reportc <- availableReport
		time.Sleep(time.Millisecond) // Simulate pollinginterval, and forces writer to start before reader
		met.Mu.RLock()
		if !alert.WebsiteHasRecovered {
			t.Error("Phase 3: Website has recovered")
		}
		if alert.WebsiteWasDown {
			t.Error("Phase 3: Website wasn't down")
		}
		met.Mu.RUnlock()

		// website in good shape
		for i := 0; i < 10; i++ {
			reportc <- availableReport
			time.Sleep(time.Millisecond) // Simulate pollinginterval, and forces writer to start before reader
			met.Mu.RLock()
			if alert.WebsiteHasRecovered {
				t.Error("Phase 4: Website hasn't recovered")
			}

			if alert.WebsiteWasDown {
				t.Error("Phase 4: Website wasn't down")
			}
			met.Mu.RUnlock()
		}
	}
}
