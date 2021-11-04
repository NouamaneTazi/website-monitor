package cui

import (
	"log"
	"time"

	"github.com/NouamaneTazi/iseeu/internal/config"
	"github.com/gizak/termui/v3"
)

// handleCUI creates CUI and handles keyboardBindings
func handleCUI() {
	var ui UI
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
			UpdateUI(ui, data, config.LongUIRefreshInterval)
			if ui.Alerts.SelectedRow == len(ui.Alerts.Rows)-1 || counter < 2 {
				ui.Alerts.ScrollPageDown()
				termui.Render(ui.Alerts)
			}
		default:
		}
		select {
		case <-shortTick.C:
			counter++
			lenRows := len(ui.Alerts.Rows)
			UpdateUI(ui, data, config.ShortUIRefreshInterval)
			if ui.Alerts.SelectedRow == lenRows-1 || counter < 2 {
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
