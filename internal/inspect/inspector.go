package inspect

import (
	"log"
	"time"

	"github.com/NouamaneTazi/iseeu/internal/config"
	"github.com/gocolly/colly/v2"
)

// Report collects useful metrics from a single HTTP request made by an Inspector
type Report struct {
	url               string
	StatusCode        int
	ConnectDuration   time.Duration
	FirstByteDuration time.Duration
}

// Inspector monitors a url every polling interval
type Inspector struct {
	ticker    *time.Ticker     // periodic ticker
	url       string           // current URLs
	reportc   chan *Report     // channel used to report metrics
	collector *colly.Collector // colly collector which sends and traces HTTP requests
}

// create a new collector
func newTraceCollector(responseCb colly.ResponseCallback) *colly.Collector {
	collector := colly.NewCollector(colly.TraceHTTP(), colly.AllowURLRevisit())
	collector.OnResponse(responseCb)
	return collector
}

func NewInspector(url string, pollingInterval time.Duration) chan *Report {
	// TODO: can we modify calculations so that we keep track only of last one
	// number of reports to keep track of
	maxNumOfReports := int(config.MaxHistoryPerURL / pollingInterval)

	reportc := make(chan *Report, maxNumOfReports)
	// init new inspector
	inspector := &Inspector{
		ticker:  time.NewTicker(pollingInterval),
		reportc: reportc,
		url:     url,
		collector: newTraceCollector(func(resp *colly.Response) {
			if resp.Trace == nil {
				log.Print("Failed to initialize trace")
			}
			// create report from trace
			report := &Report{
				url:               url,
				StatusCode:        resp.StatusCode,
				ConnectDuration:   resp.Trace.ConnectDuration,
				FirstByteDuration: resp.Trace.FirstByteDuration,
			}

			// send report over to metrics for further analytics
			reportc <- report
		}),
	}

	// start monitoring
	go inspector.start()
	return reportc
}

func (inspector *Inspector) inspect() {
	err := inspector.collector.Visit(inspector.url)
	if err != nil {
		log.Printf("Failed to visit url %s", inspector.url)
	}
}

func (inspector *Inspector) start() {
	for {
		select {
		case <-inspector.ticker.C:
			// When the ticker fires, inspect url
			inspector.inspect()
		}
	}
}

// updateURLReports updates URL reports with useful metrics about website
// a single http request generates a single report
// we drop reports older than maxHistoryPerURL
// func (inspector *Inspector) updateURLReports(url string, report *Report) {
// 	queue := inspector.Reports
// 	queue = queue[1:] // TODO: make sure we reallocate memory
// 	inspector.Reports = append(queue, report)
// }
