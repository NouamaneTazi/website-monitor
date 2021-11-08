package metrics

import (
	"math"
	"sync"
	"time"

	"github.com/NouamaneTazi/iseeu/internal/config"
	"github.com/NouamaneTazi/iseeu/internal/inspect"
)

// Metrics contains url trace information and aggregation of reports over a short and a long interval
type Metrics struct {
	Url             string                 // the url being monitored
	PollingInterval time.Duration          // the url's polling interval
	LastTimestamp   time.Time              // last updated time stamp
	reportc         <-chan *inspect.Report // the reports channel
	Mu              sync.RWMutex
	AggData         *AggData // aggregated data that will be passed to UI
	Alert           *Alert   // holds alerting logic
}

// Alert tracks url alerts
type Alert struct {
	statuscodesc        chan int
	Availability        float64
	WebsiteWasDown      bool
	WebsiteHasRecovered bool
}

// AggData regroups the aggregated data over a short and a long interval
type AggData struct {
	Short, Long *IntervalAggData
}

// IntervalAggData aggregates data over `historyInterval`
type IntervalAggData struct {
	historyInterval   time.Duration     // specifies duration of relevant reports history
	numOfAggReports   int               // specifies number of relevant reports (= historyInterval / PollingInterval)
	StatusCodesCount  map[int]int       // hold count of status codes of past reports
	statuscodesc      chan int          // channel to update status codes count (we don't need queue)
	reportQueue       []*inspect.Report // stores last (historyInterval / PollingInterval) reports to calculate aggregated metrics
	Availability      float64           // Website availability (%)
	ConnectDuration   [2]int            // [avg, max] in milliseconds
	FirstByteDuration [2]int            // [avg, max] in milliseconds
}

// NewMetrics inits and return new Metrics object.
func NewMetrics(reportc <-chan *inspect.Report, pollingInterval time.Duration) *Metrics {

	// Use queues to find max values in metrics
	shortReportQueue := make([]*inspect.Report, 0, int(config.ShortStatsHistoryInterval/pollingInterval))
	longReportQueue := make([]*inspect.Report, 0, int(config.LongStatsHistoryInterval/pollingInterval))

	return &Metrics{
		PollingInterval: pollingInterval,
		reportc:         reportc,
		AggData: &AggData{
			Short: &IntervalAggData{
				historyInterval:  config.ShortStatsHistoryInterval,
				numOfAggReports:  int(config.ShortStatsHistoryInterval / pollingInterval),
				statuscodesc:     make(chan int, int(config.ShortStatsHistoryInterval/pollingInterval)),
				reportQueue:      shortReportQueue,
				StatusCodesCount: make(map[int]int),
			},
			Long: &IntervalAggData{
				historyInterval:  config.LongStatsHistoryInterval,
				numOfAggReports:  int(config.LongStatsHistoryInterval / pollingInterval),
				statuscodesc:     make(chan int, int(config.LongStatsHistoryInterval/pollingInterval)),
				reportQueue:      longReportQueue,
				StatusCodesCount: make(map[int]int)},
		},
		Alert: &Alert{
			statuscodesc: make(chan int, int(config.WebsiteAlertInterval/pollingInterval)),
		},
	}
}

// ListenAndProcess listens for incoming reports and updates metrics
func (m *Metrics) ListenAndProcess() {
	// every `pollingInterval` this receives a report from Inspector
	for report := range m.reportc {
		// update metrics data
		m.update(report)
	}
}

// update updates metrics aggregated data and alerts from a report
func (m *Metrics) update(newReport *inspect.Report) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	// defines metrics url upon first report it gets
	if m.Url == "" {
		m.Url = newReport.Url
	}
	m.LastTimestamp = time.Now()
	m.AggData.update(newReport)
	m.Alert.update(newReport)
}

// update updates `AggData` data from a report
func (agg *AggData) update(newReport *inspect.Report) {
	agg.Short.aggregate(newReport)
	agg.Long.aggregate(newReport)
}

// aggregate aggregates report for the past `agg.historyInterval` interval
func (agg *IntervalAggData) aggregate(newReport *inspect.Report) {
	// add newReport to reportQueue
	if len(agg.reportQueue) >= agg.numOfAggReports {
		// * note first element will be garbage collected when enough new elements are added to the slice to cause reallocation
		// check https://stackoverflow.com/questions/2818852/is-there-a-queue-implementation#comment103168917_26863706
		agg.reportQueue = agg.reportQueue[1:]
	}
	agg.reportQueue = append(agg.reportQueue, newReport)

	// update avg/max stats
	agg.updateAvgMax(agg.reportQueue)

	// update status count
	agg.updateStatusCount(newReport)

	// update availability
	agg.Availability = float64(agg.StatusCodesCount[200]) / float64(agg.numOfAggReports)
	agg.Availability = math.Round(agg.Availability*100) / 100
}

// updateAvgMax updates `IntervalAggData` with the aggregated avg and max of past reports
// * NOTE we could use a maxheap or maximum sliding window for O(1) time complexity here
func (agg *IntervalAggData) updateAvgMax(reportQueue []*inspect.Report) {
	// assuming reportQueue has the newReport
	agg.ConnectDuration, agg.FirstByteDuration = [2]int{0, 0}, [2]int{0, 0}

	// TODO: this can be refactored
	for _, report := range reportQueue {
		// update ConnectDuration (-1 means there has been an error)
		if report.ConnectDuration != -1 {
			if int(report.ConnectDuration) > agg.ConnectDuration[1] {
				agg.ConnectDuration[1] = int(report.ConnectDuration.Milliseconds())
			}
			agg.ConnectDuration[0] += int(report.ConnectDuration.Milliseconds()) / agg.numOfAggReports
		}

		// update FirstByteDuration (-1 means there has been an error)
		if report.FirstByteDuration != -1 {
			if int(report.FirstByteDuration) > agg.FirstByteDuration[1] {
				agg.FirstByteDuration[1] = int(report.FirstByteDuration.Milliseconds())
			}
			agg.FirstByteDuration[0] += int(report.FirstByteDuration.Milliseconds()) / agg.numOfAggReports
		}
	}
}

// updateStatusCount updates status count using `agg.statuscodesc` channel (no need for a queue here)
func (agg *IntervalAggData) updateStatusCount(newReport *inspect.Report) {
	// only start dequeuing from channel after it becomes full
	if len(agg.statuscodesc) == cap(agg.statuscodesc) {
		statusCode := <-agg.statuscodesc
		if _, ok := agg.StatusCodesCount[statusCode]; ok {
			agg.StatusCodesCount[statusCode]--
		}
	}
	// note that statuscodesc is a buffered chan of capacity `numOfAggReports`
	agg.statuscodesc <- newReport.StatusCode
	agg.StatusCodesCount[newReport.StatusCode]++
}

// update handles the alerting logic
// Alert if website availability is below config.CriticalAvailability for the past config.WebsiteAlertInterval
// Alert if website has recovered
func (alert *Alert) update(newReport *inspect.Report) {
	oldAvailability := alert.Availability
	// update availability using alert.statuscodesc channel
	// only start dequeuing from channel after it becomes full
	if len(alert.statuscodesc) == cap(alert.statuscodesc) {
		statusCode := <-alert.statuscodesc
		if statusCode == 200 {
			alert.Availability -= 1 / float64(cap(alert.statuscodesc))
		}
	}
	// note that statuscodesc is a buffered chan of capacity `numOfAggReports`
	alert.statuscodesc <- newReport.StatusCode
	if newReport.StatusCode == 200 {
		alert.Availability += 1 / float64(cap(alert.statuscodesc))
	}
	alert.Availability = math.Round(alert.Availability*100) / 100
	if alert.WebsiteHasRecovered {
		alert.WebsiteHasRecovered = false
	}
	if oldAvailability < config.CriticalAvailability && alert.Availability >= config.CriticalAvailability {
		alert.WebsiteHasRecovered = true
	}
	if newReport.StatusCode == 200 {
		alert.WebsiteWasDown = false
	} else {
		alert.WebsiteWasDown = true
	}
}

// updateAvg keeps track of the avg of a metric
// note: this method only uses newest and oldest metric, and doesn't need a queue
// func updateAvgDEPRECATED(aggMetric int, newMetric time.Duration, deprMetric time.Duration, numOfReports int) int {
// 	if deprMetric != -1 {
// 		aggMetric -= int(newMetric.Milliseconds()) / numOfReports
// 	}
// 	aggMetric += int(newMetric.Milliseconds()) / numOfReports
// 	return aggMetric
// }
