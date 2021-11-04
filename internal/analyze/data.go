package analyze

import (
	"time"

	"github.com/NouamaneTazi/iseeu/internal/inspect"
)

// Metrics represents data to be passed to UI.
type Metrics struct {
	url               string // the url being monitored
	WebsitesStatsList []WebsiteStats
	LastTimestamp     time.Time // last updated time stamp
}

// NewUIData inits and return new ui data object.
func NewUIData(inspectorsList []*inspect.Inspector) *Metrics {
	// create
	return &Metrics{WebsitesStatsList: make([]WebsiteStats, len(inspectorsList))}
}

// UpdateData collects data from inspectors and updates websites stats
func (data *Metrics) UpdateData(interval time.Duration) {
	// every time we get data from reportChan
	//
	// for
	// data.WebsitesStatsList[i] = data.WebsitesStatsList[i].calculateStats(reports, interval, inspector.Url)

	// Set last updated data time
	data.LastTimestamp = time.Now()
}
