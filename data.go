package main

import (
	"time"
)

// UIData represents data to be passed to UI.
type UIData struct {
	inspectorsList    []*Inspector
	WebsitesStatsList []WebsiteStats
	LastTimestamp     time.Time // last updated time stamp
}

// NewUIData inits and return new ui data object.
func NewUIData(inspectorsList []*Inspector) *UIData {
	return &UIData{inspectorsList: inspectorsList, WebsitesStatsList: make([]WebsiteStats, len(inspectorsList))}
}

// updateData collects data from inspectors and updates websites stats
func (data *UIData) updateData(interval *time.Duration) {
	for i := 0; i < len(data.inspectorsList); i++ {
		inspector := data.inspectorsList[i]
		reports := inspector.reports
		data.WebsitesStatsList[i] = data.WebsitesStatsList[i].calculateStats(reports, interval, inspector.url)
	}
	// Set last updated data time
	data.LastTimestamp = time.Now()
}
