package analyze

import (
	"time"

	"github.com/NouamaneTazi/iseeu/internal/inspect"
)

// UIData represents data to be passed to UI.
type UIData struct {
	inspectorsList    []*inspect.Inspector
	WebsitesStatsList []WebsiteStats
	LastTimestamp     time.Time // last updated time stamp
}

// NewUIData inits and return new ui data object.
func NewUIData(inspectorsList []*inspect.Inspector) *UIData {
	return &UIData{inspectorsList: inspectorsList, WebsitesStatsList: make([]WebsiteStats, len(inspectorsList))}
}

// UpdateData collects data from inspectors and updates websites stats
func (data *UIData) UpdateData(interval time.Duration) {
	for i := 0; i < len(data.inspectorsList); i++ {
		inspector := data.inspectorsList[i]
		reports := inspector.Reports
		data.WebsitesStatsList[i] = data.WebsitesStatsList[i].calculateStats(reports, interval, inspector.Url)
	}
	// Set last updated data time
	data.LastTimestamp = time.Now()
}
