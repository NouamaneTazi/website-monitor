/*
Copyright 2021, 2021 the ISeeU contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NouamaneTazi/iseeu/internal/config"
	"github.com/NouamaneTazi/iseeu/internal/cui"
	"github.com/NouamaneTazi/iseeu/internal/inspect"
	"github.com/NouamaneTazi/iseeu/internal/metrics"
)

func main() {
	// Parse urls and polling intervals and options, and updates `config`
	flag.DurationVar(&config.ShortUIRefreshInterval, "sui", 2*time.Second, "Short refreshing UI interval (in seconds)")
	flag.DurationVar(&config.LongUIRefreshInterval, "lui", 10*time.Second, "Long refreshing UI interval (in seconds)")
	flag.DurationVar(&config.ShortStatsHistoryInterval, "sstats", 10*time.Second, "Short history interval (in minutes)")
	flag.DurationVar(&config.LongStatsHistoryInterval, "lstats", 60*time.Second, "Long history interval (in minutes)")
	flag.DurationVar(&config.WebsiteAlertInterval, "alertint", 60*time.Second,
		"Shows alert if website is down for `WebsiteAlertInterval` minutes")
	parse()

	// initiate array holding metrics. each metrics corresponds to one URL
	// metrics are updated over time through `ListenAndProcess()` method
	// note: we could have used a channel of capacity one to always keep the latest metric ready
	// but I think this is simpler and more understandable
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

	// show metrics in UI
	cui.HandleCUI(stats)
	select {}
}

// parse parses urls and validates flags
func parse() {
	flag.Parse()
	tail := flag.Args()

	// validates the format `URL POLLING_INTERVAL`
	if len(tail) > 0 && len(tail)%2 == 0 {
		for i := 0; i < len(tail); i += 2 {
			pollingInterval, err := strconv.Atoi(tail[i+1])
			if err != nil {
				fmt.Println("Error converting polling interval to int", err)
				os.Exit(2)
			}
			// update config
			config.UrlsPollingsIntervals[parseURL(tail[i])] = time.Duration(pollingInterval) * time.Second
		}
	} else {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL1 POLLING_INTERVAL1 URL2 POLLING_INTERVAL2\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "OPTIONS:")
		flag.PrintDefaults()
		log.Fatal("Urls must be provided with their respective polling intervals.")
	}
}

// parseURL reassembles the URL into a valid URL string
func parseURL(uri string) string {
	if !strings.Contains(uri, "://") && !strings.HasPrefix(uri, "//") {
		uri = "//" + uri
	}

	url, err := url.Parse(uri)
	if err != nil {
		log.Panicf("could not parse url %q: %v", uri, err)
	}
	if url.Scheme == "" {
		url.Scheme = "http"
		if !strings.HasSuffix(url.Host, ":80") {
			url.Scheme += "s"
		}
	}

	return url.String()
}
