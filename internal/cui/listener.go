package cui

import (
	"time"

	"github.com/NouamaneTazi/website-monitor/internal/config"
	"github.com/NouamaneTazi/website-monitor/internal/metrics"
	"github.com/gizak/termui/v3"
)

// handleCUI creates CUI and handles keyboardBindings
func HandleCUI(data []*metrics.Metrics) error {
	var ui UI

	if err := ui.Init(); err != nil {
		return err
	}
	defer ui.Close()

	// Ticker that refreshes UI
	shortTick := time.NewTicker(config.ShortUIRefreshInterval)
	longTick := time.NewTicker(config.LongUIRefreshInterval)

	// keyboard bindings
	uiEvents := termui.PollEvents()
	for {
		select {
		case <-longTick.C:
			ui.UpdateUI(data, config.LongUIRefreshInterval)

		case <-shortTick.C:
			ui.UpdateUI(data, config.ShortUIRefreshInterval)

		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				// TODO: interrupt app gracefully
				return nil
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
