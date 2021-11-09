package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NouamaneTazi/website-monitor/internal/config"
	"github.com/NouamaneTazi/website-monitor/internal/cui"
	"github.com/NouamaneTazi/website-monitor/internal/inspect"
	"github.com/NouamaneTazi/website-monitor/internal/metrics"
)

func main() {
	// Parse urls and polling intervals and options
	flag.DurationVar(&config.ShortUIRefreshInterval, "sui", 2*time.Second, "Short refreshing UI interval (in seconds)")
	flag.DurationVar(&config.LongUIRefreshInterval, "lui", 10*time.Second, "Long refreshing UI interval (in seconds)")
	flag.DurationVar(&config.ShortStatsHistoryInterval, "sstats", 10*time.Second, "Short refreshes show stats for past `ShortStatsHistoryInterval` minutes")
	flag.DurationVar(&config.LongStatsHistoryInterval, "lstats", 60*time.Second,
		"Long refreshes show stats for past `LongStatsHistoryInterval` minutes")
	flag.DurationVar(&config.WebsiteAlertInterval, "alertint", 10*time.Second,
		"Shows alert if website is down for `WebsiteAlertInterval` minutes")
	flag.Float64Var(&config.CriticalAvailability, "crit", 0.8, "Availability of websites below which we show an alert")
	err := parse()
	if err != nil {
		log.Fatalln("Failed parsing command arguments: ", err)
	}

	// initiate array holding metrics. each metrics corresponds to one URL
	// metrics are updated over time through `ListenAndProcess()` method
	// * note: we could have used a channel of capacity one to always keep the latest metric ready
	// * but I think this is simpler and more understandable
	stats := make([]*metrics.Metrics, 0, len(config.UrlsPollingsIntervals))

	for url, pollingInterval := range config.UrlsPollingsIntervals {
		// Init the inspectors, where each inspector monitors a single URL
		// and sends back the trace report over the `reportc` channel
		reportc := inspect.NewInspector(url, pollingInterval)

		// init metrics server for each url and add them to `stats` array
		s := metrics.NewMetrics(reportc, pollingInterval)
		stats = append(stats, s)

		// start a goroutine for each url, which going to listen to `reportc` channel
		// process its reports, and updates the corresponding `Metrics`
		go s.ListenAndProcess()
	}

	// create CUI and handle keyboardBindings
	err = cui.HandleCUI(stats)
	if err != nil {
		log.Fatalf("Failed to start CUI %v", err)
	}
}

// parse parses urls and validates command format
func parse() error {
	flag.Parse()
	tail := flag.Args()
	if len(tail) > 0 && len(tail)%2 == 0 {
		for i := 0; i < len(tail); i += 2 {
			pollingInterval, err := strconv.Atoi(tail[i+1])
			if err != nil {
				return fmt.Errorf("error converting polling interval %v to int", tail[i+1])
			}
			url, err := parseURL(tail[i])
			if err != nil {
				return err
			}
			config.UrlsPollingsIntervals[url] = time.Duration(pollingInterval) * time.Second
		}
	} else {
		fmt.Fprintf(os.Stderr, "\nUsage: %s [OPTIONS] URL1 POLLING_INTERVAL1 URL2 POLLING_INTERVAL2\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -crit 0.3 -sui 1s google.com 2 http://google.fr 1\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "OPTIONS:")
		flag.PrintDefaults()
		return errors.New("urls must be provided with their respective polling intervals")
	}
	return nil
}

// parseURL reassembles the URL into a valid URL string
func parseURL(uri string) (string, error) {
	if !strings.Contains(uri, "://") && !strings.HasPrefix(uri, "//") {
		uri = "//" + uri
	}

	url, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if url.Scheme == "" {
		url.Scheme = "http"
		if !strings.HasSuffix(url.Host, ":80") {
			url.Scheme += "s"
		}
	}

	return url.String(), nil
}
