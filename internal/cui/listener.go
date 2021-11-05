package cui

import (
	"log"
	"runtime/debug"
	"time"

	"github.com/NouamaneTazi/iseeu/internal/config"
	"github.com/NouamaneTazi/iseeu/internal/metrics"
	"github.com/gizak/termui/v3"
)

// handleCUI creates CUI and handles keyboardBindings
func HandleCUI(data []*metrics.Metrics) {
	var ui UI
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovering from panic:")
			debug.PrintStack()
		}
	}()
	if err := ui.Init(); err != nil {
		log.Fatalf("Failed to start CUI %v", err)
	}
	defer ui.Close()

	// Ticker that refreshes UI
	shortTick := time.NewTicker(config.ShortUIRefreshInterval)
	longTick := time.NewTicker(config.LongUIRefreshInterval)

	var counter int
	uiEvents := termui.PollEvents()
	for {
		select {
		case <-longTick.C:
			counter++
			ui.UpdateUI(data, config.LongUIRefreshInterval)
			// TODO: fix this
			if counter > 1 {
				ui.Alerts.ScrollPageDown()
				termui.Render(ui.Alerts)
			}
		case <-shortTick.C:
			counter++
			ui.UpdateUI(data, config.ShortUIRefreshInterval)
			if counter > 1 {
				ui.Alerts.ScrollPageDown()
				termui.Render(ui.Alerts)
			}

		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				// interrupt app gracefully
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
