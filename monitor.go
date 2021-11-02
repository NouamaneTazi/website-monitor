package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gizak/termui/v3"
)

var (
	shortUIRefreshInterval    = flag.Duration("sui", 2*time.Second, "Short refreshing UI interval (in seconds)")
	longUIRefreshInterval     = flag.Duration("lui", 10*time.Second, "Long refreshing UI interval (in seconds)")
	shortStatsHistoryInterval = flag.Duration("sstats", 10*time.Second, "Short history interval (in minutes)")
	longStatsHistoryInterval  = flag.Duration("lstats", 60*time.Second, "Long history interval (in minutes)")
	urlsPollingsIntervals     = make(map[string]time.Duration) // maps urls to their corresponding polling interval
	maxHistoryPerURL          = 1 * time.Minute                // max stats history duration
	criticalAvailability      = 0.8                            // availability of websites below which we show an alert
)

func main() {

	// Parse urls and polling intervals and options
	flag.Parse()
	tail := flag.Args()
	if len(tail) > 0 && len(tail)%2 == 0 {
		for i := 0; i < len(tail); i += 2 {
			pollingInterval, err := strconv.Atoi(tail[i+1])
			if err != nil {
				fmt.Println("Error converting polling interval to int", err)
				os.Exit(2)
			}
			urlsPollingsIntervals[parseURL(tail[i])] = time.Duration(pollingInterval) * time.Second
		}
	} else {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL1 POLLING_INTERVAL1 URL2 POLLING_INTERVAL2\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "OPTIONS:")
		flag.PrintDefaults() //TODO: better usage
		log.Fatal("Urls must be provided with their respective polling intervals.")
	}

	// Init the inspectors, where each inspector monitors a single URL
	inspectorsList := make([]*Inspector, 0, len(urlsPollingsIntervals))
	for url, pollingInterval := range urlsPollingsIntervals {
		inspector := NewInspector(URL(url), intervalInspection(pollingInterval))
		inspectorsList = append(inspectorsList, inspector)

		// Init website monitoring
		go inspector.startLoop()
	}

	// Init UIData
	data := NewUIData(inspectorsList)

	// Start proper UI
	var ui UI
	if err := ui.Init(); err != nil {
		log.Fatal("Failed to start CLI %v", err)
	}
	defer ui.Close()

	// Ticker that refreshes UI
	shortTick := time.NewTicker(*shortUIRefreshInterval)

	var counter int
	uiEvents := termui.PollEvents()
	for {
		select {
		case <-shortTick.C:
			counter++
			lenRows := len(ui.Alerts.Rows)
			if counter%int(*longUIRefreshInterval / *shortUIRefreshInterval) != 0 {
				UpdateUI(ui, data, shortUIRefreshInterval)
			} else {
				UpdateUI(ui, data, longUIRefreshInterval)
			}
			if ui.Alerts.SelectedRow == lenRows-1 || counter < 2 {
				ui.Alerts.ScrollPageDown()
				termui.Render(ui.Alerts)
			}

		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "j", "<Down>":
				ui.Alerts.ScrollDown()
			case "k", "<Up>":
				ui.Alerts.ScrollUp()
			case "<C-d>":
				ui.Alerts.ScrollHalfPageDown()
			case "<C-u>":
				ui.Alerts.ScrollHalfPageUp()
			case "<C-f>":
				ui.Alerts.ScrollPageDown()
			case "<C-b>":
				ui.Alerts.ScrollPageUp()
			case "<Home>":
				ui.Alerts.ScrollTop()
			case "G", "<End>":
				ui.Alerts.ScrollBottom()
			}

			termui.Render(ui.Alerts)
		}
	}

}

// UpdateUI collects data from inspectors and refreshes UI.
func UpdateUI(ui UI, data *UIData, interval *time.Duration) {
	data.updateData(interval)
	ui.Update(data, interval)
}
