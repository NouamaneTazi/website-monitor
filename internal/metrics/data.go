package metrics

import (
	"math"
	"time"

	"github.com/NouamaneTazi/iseeu/internal/config"
	"github.com/NouamaneTazi/iseeu/internal/inspect"
)

// Metrics represents data to be passed to UI.
type Metrics struct {
	Url           string    // the url being monitored
	LastTimestamp time.Time // last updated time stamp
	reportc       chan *inspect.Report
	AggData       *AggData
	Alert         *Alert
}
type Alert struct {
	statuscodesc        chan int
	Availability        float64
	WebsiteWasDown      bool
	WebsiteHasRecovered bool
}

// AggData regroups the aggregated data that will be passed to UI
type AggData struct {
	Short, Long *IntervalAggData
}

// IntervalAggData aggregates data over `historyInterval`
type IntervalAggData struct {
	historyInterval   time.Duration
	StatusCodesCount  map[int]int
	statuscodesc      chan int
	Availability      float64
	ConnectDuration   [2]int // [avg, max] in milliseconds
	FirstByteDuration [2]int // [avg, max] in milliseconds
}

// NewMetrics inits and return new Metrics object.
func NewMetrics(reportc chan *inspect.Report, pollingInterval time.Duration) *Metrics {
	return &Metrics{
		reportc: reportc,
		AggData: &AggData{
			Short: &IntervalAggData{
				historyInterval:  config.ShortStatsHistoryInterval,
				statuscodesc:     make(chan int, int(pollingInterval/config.ShortStatsHistoryInterval)),
				StatusCodesCount: make(map[int]int),
			},
			Long: &IntervalAggData{historyInterval: config.LongStatsHistoryInterval,
				statuscodesc:     make(chan int, int(pollingInterval/config.LongStatsHistoryInterval)),
				StatusCodesCount: make(map[int]int)},
		},
		Alert: &Alert{
			statuscodesc: make(chan int, int(pollingInterval/config.WebsiteAlertInterval)),
		},
	}
}

func (m Metrics) ListenAndProcess() {
	// every `pollingInterval` this receives a report from Inspector
	for report := range m.reportc {
		// defines metrics url upon first report it gets
		if m.Url == "" {
			m.Url = report.Url
		}
		// update metrics data
		m.LastTimestamp = time.Now()
		m.AggData.update(report)
		m.Alert.update(report)
	}
}

func (agg AggData) update(newReport *inspect.Report) {
	agg.Short.update(newReport)
	agg.Long.update(newReport)
}
func (agg IntervalAggData) update(newReport *inspect.Report) {
	// update avg/max trackers
	numOfReports := int(newReport.PollingInterval / agg.historyInterval)
	agg.ConnectDuration = updateAvgMax(agg.ConnectDuration, newReport.ConnectDuration, numOfReports)
	agg.FirstByteDuration = updateAvgMax(agg.FirstByteDuration, newReport.FirstByteDuration, numOfReports)

	// update status count
	// only start dequeuing from channel after it becomes full
	if len(agg.statuscodesc) == cap(agg.statuscodesc) {
		statusCode := <-agg.statuscodesc
		agg.StatusCodesCount[statusCode]--
	}
	// note that statuscodesc is a buffered chan of capacity pollingInterval // config.[...]StatsHistoryInterval
	agg.statuscodesc <- newReport.StatusCode
	agg.StatusCodesCount[newReport.StatusCode]++

	// update availability
	agg.Availability = float64(agg.StatusCodesCount[200]) / float64(numOfReports)
}

// updateAvgMax keeps track of the avg and max of a metric
func updateAvgMax(aggMetric [2]int, newMetric time.Duration, numOfReports int) [2]int {
	aggMetric[0] += int(newMetric.Milliseconds()) / numOfReports
	aggMetric[1] = int(math.Max(float64(aggMetric[1]), float64(newMetric.Milliseconds())))
	return aggMetric
}

// UpdateAlerting handles the alerting logic
// Checks if website availability is below config.CriticalAvailability for the past config.WebsiteAlertInterval
// Checks if website availability has recovered
func (alert Alert) update(newReport *inspect.Report) {
	// update availability using alert.statuscodesc channel
	// only start dequeuing from channel after it becomes full
	if len(alert.statuscodesc) == cap(alert.statuscodesc) {
		statusCode := <-alert.statuscodesc
		if statusCode == 200 {
			alert.Availability -= 1 / float64(cap(alert.statuscodesc))
		}
	}
	// note that statuscodesc is a buffered chan of capacity pollingInterval // config.[...]StatsHistoryInterval
	alert.statuscodesc <- newReport.StatusCode
	if newReport.StatusCode == 200 {
		alert.Availability += 1 / float64(cap(alert.statuscodesc))
	}

	if alert.WebsiteHasRecovered {
		alert.WebsiteHasRecovered = false
	}
	if alert.WebsiteWasDown && alert.Availability >= config.CriticalAvailability {
		alert.WebsiteWasDown = false
		alert.WebsiteHasRecovered = true
	}
	if alert.Availability < config.CriticalAvailability {
		alert.WebsiteWasDown = true
	}
}
