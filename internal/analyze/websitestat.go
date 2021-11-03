package analyze

import (
	"math"
	"time"

	"github.com/NouamaneTazi/iseeu/internal/config"
	"github.com/NouamaneTazi/iseeu/internal/inspect"
)

// WebsiteStats represents interesting metrics about a website
type WebsiteStats struct {
	Url                 string
	StatusCodesCount    map[int]int
	Availability        float64
	websiteWasDown      bool
	WebsiteHasRecovered bool
	DNSLookup           [2]int // [avg, max]
	TCPConnection       [2]int // [avg, max]
	TLSHandshake        [2]int // [avg, max]
	ServerProcessing    [2]int // [avg, max]
	ContentTransfer     [2]int // [avg, max]
	NameLookup          [2]int // [avg, max]
	Connect             [2]int // [avg, max]
	PreTransfer         [2]int // [avg, max]
	StartTransfer       [2]int // [avg, max]
	Total               [2]int // [avg, max]
}

// calculateStats aggregates reports coming from inspectors and returns updated website stats
// the aggregation depends on the refresh interval (if short we keep only config.ShortStatsHistoryInterval..)
func (stat WebsiteStats) calculateStats(reports []*inspect.Report, refreshInterval time.Duration, url string) WebsiteStats {
	stat = WebsiteStats{StatusCodesCount: make(map[int]int), websiteWasDown: stat.websiteWasDown}
	stat.Url = url

	// keep only a number of reports depending on whether it's a long or short refresh
	var usefulNumOfReports int
	switch refreshInterval {
	case config.ShortUIRefreshInterval:
		usefulNumOfReports = int(config.ShortStatsHistoryInterval / config.UrlsPollingsIntervals[url])
	case config.LongUIRefreshInterval:
		usefulNumOfReports = int(config.LongStatsHistoryInterval / config.UrlsPollingsIntervals[url])
	}
	reports = reports[len(reports)-usefulNumOfReports:]

	// Aggregates the reports to have new stats
	for _, report := range reports {
		stat.StatusCodesCount[report.StatusCode]++
		// Calculate average and maximum of reports of last `config.ShortStatsHistoryInterval` or `config.LongStatsHistoryInterval` minutes
		stat.DNSLookup = updateAvgMax(stat.DNSLookup, report.DNSLookup)
		stat.TCPConnection = updateAvgMax(stat.TCPConnection, report.TCPConnection)
		stat.TLSHandshake = updateAvgMax(stat.TLSHandshake, report.TLSHandshake)
		stat.ServerProcessing = updateAvgMax(stat.ServerProcessing, report.ServerProcessing)
		stat.ContentTransfer = updateAvgMax(stat.ContentTransfer, report.ContentTransfer)
		stat.NameLookup = updateAvgMax(stat.NameLookup, report.NameLookup)
		stat.Connect = updateAvgMax(stat.Connect, report.Connect)
		stat.PreTransfer = updateAvgMax(stat.PreTransfer, report.PreTransfer)
		stat.StartTransfer = updateAvgMax(stat.StartTransfer, report.StartTransfer)
		stat.Total = updateAvgMax(stat.Total, report.Total)
	}
	if len(reports) == 0 {
		// TODO: handle case where polling interval is longer than stathistory interval
	} else {
		stat.Availability = float64(stat.StatusCodesCount[200]) / float64(len(reports))
		stat.DNSLookup[0] = stat.DNSLookup[0] / len(reports)
		stat.TCPConnection[0] = stat.TCPConnection[0] / len(reports)
		stat.TLSHandshake[0] = stat.TLSHandshake[0] / len(reports)
		stat.ServerProcessing[0] = stat.ServerProcessing[0] / len(reports)
		stat.ContentTransfer[0] = stat.ContentTransfer[0] / len(reports)
		stat.NameLookup[0] = stat.NameLookup[0] / len(reports)
		stat.Connect[0] = stat.Connect[0] / len(reports)
		stat.PreTransfer[0] = stat.PreTransfer[0] / len(reports)
		stat.StartTransfer[0] = stat.StartTransfer[0] / len(reports)
		stat.Total[0] = stat.Total[0] / len(reports)
	}

	stat.UpdateAlerting(refreshInterval)

	return stat
}

// UpdateAlerting handles the alerting logic
// Checks if website availability is below config.CriticalAvailability for the past config.ShortStatsHistoryInterval
// Checks if website availability has recovered
func (stat *WebsiteStats) UpdateAlerting(refreshInterval time.Duration) {
	switch refreshInterval {
	case config.ShortUIRefreshInterval:
		if stat.WebsiteHasRecovered {
			stat.WebsiteHasRecovered = false
		}
		if stat.websiteWasDown && stat.Availability >= config.CriticalAvailability {
			stat.websiteWasDown = false
			stat.WebsiteHasRecovered = true
		}
		if stat.Availability < config.CriticalAvailability {
			stat.websiteWasDown = true
		}
	}
}

// updateAvgMax keeps track of the avg and max of a metric
func updateAvgMax(metric [2]int, source time.Duration) [2]int {
	metric[0] += int(source.Milliseconds())
	metric[1] = int(math.Max(float64(metric[1]), float64(source.Milliseconds())))
	return metric
}
