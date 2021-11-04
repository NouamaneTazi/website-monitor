package inspect

import (
	"log"
	"time"

	"github.com/NouamaneTazi/iseeu/internal/config"
	"github.com/gocolly/colly/v2"
)

// Report collects useful metrics from a single HTTP request made by an Inspector
type Report struct {
	Url               string
	PollingInterval   time.Duration
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

// newTraceCollector creates a new collector which traces http requests
func newTraceCollector(responseCb colly.ResponseCallback) *colly.Collector {
	collector := colly.NewCollector(colly.TraceHTTP(), colly.AllowURLRevisit())
	collector.OnResponse(responseCb)
	return collector
}

// NewInspector initializes an Inspector
func NewInspector(url string, PollingInterval time.Duration) chan *Report {
	// TODO: can we modify calculations so that we keep track only of last one
	// number of reports to keep track of
	maxNumOfReports := int(config.LongStatsHistoryInterval / PollingInterval)
	reportc := make(chan *Report, maxNumOfReports)

	// init new inspector
	inspector := &Inspector{
		ticker:  time.NewTicker(PollingInterval),
		reportc: reportc,
		url:     url,
		collector: newTraceCollector(func(resp *colly.Response) {
			if resp.Trace == nil {
				log.Print("Failed to initialize trace")
			}
			// create report from trace
			report := &Report{
				Url:               url,
				PollingInterval:   PollingInterval,
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
	// log.Printf("Visiting %s", inspector.url)

	err := inspector.collector.Visit(inspector.url)
	if err != nil {
		// TODO: handle this (send suiting data report)
		log.Printf("Failed to visit url %s: %v", inspector.url, err)
	}
}

func (inspector *Inspector) start() {
	for {
		select {
		case <-inspector.ticker.C:
			// When the ticker fires, inspect url
			go inspector.inspect()
		}
	}
}
