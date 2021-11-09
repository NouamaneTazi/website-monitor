package cui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/NouamaneTazi/website-monitor/internal/config"
	"github.com/NouamaneTazi/website-monitor/internal/metrics"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// UI is a ui implementation of UI interface.
type UI struct {
	Title      *widgets.Paragraph
	Status     *widgets.Paragraph
	StatsTable *widgets.Table
	Alerts     *widgets.List
}

// Init creates widgets, sets sizes and labels.
func (t *UI) Init() error {
	if err := ui.Init(); err != nil {
		return err
	}
	termWidth, termHeight := ui.TerminalDimensions()

	t.Title = func() *widgets.Paragraph {
		p := widgets.NewParagraph()
		p.Border = true
		p.Title = "Websites Monitor"
		p.SetRect(0, 0, termWidth/2, 3)
		return p
	}()
	t.Status = func() *widgets.Paragraph {
		p := widgets.NewParagraph()
		p.Border = true
		p.Title = "Status"
		p.SetRect(termWidth/2, 0, termWidth, 3)
		return p
	}()
	t.StatsTable = func() *widgets.Table {
		table := widgets.NewTable()
		table.Rows = [][]string{
			{"website",
				"Polling Interval",
				"Status code count",
				"Availability",
				"ConnectDuration",
				"FirstByteDuration"},
		}
		table.TextStyle = ui.NewStyle(ui.ColorWhite)
		table.RowSeparator = false
		return table
	}()
	t.Alerts = func() *widgets.List {
		l := widgets.NewList()
		l.Title = "Alerts [Press Arrow Keys to navigate]"
		l.Rows = []string{}
		l.TextStyle = ui.NewStyle(ui.ColorYellow)
		l.WrapText = true
		return l
	}()

	grid := ui.NewGrid()
	grid.SetRect(0, 3, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.6, t.StatsTable),
		ui.NewRow(0.4, t.Alerts),
	)

	ui.Render(grid)
	return nil
}

// Update updates UI widgets from UIData.
func (t *UI) UpdateUI(data []*metrics.Metrics, refreshInterval time.Duration) {
	// Lock so only one reader or writer goroutines at a time can access the map.
	for _, m := range data {
		m.Mu.RLock()
		defer m.Mu.RUnlock()
	}

	/* -------------------------------------------------------------------------- */
	/*                                   HEADERS                                  */
	/* -------------------------------------------------------------------------- */
	t.Title.Text = fmt.Sprintf("monitoring %d websites, press q to quit", len(data))
	t.Status.Text = fmt.Sprintf("Last update: %v (refreshes every %vs)", data[0].LastTimestamp.Format(time.Stamp), refreshInterval.Seconds())

	/* -------------------------------------------------------------------------- */
	/*                                MIDDLE TABLE                                */
	/* -------------------------------------------------------------------------- */
	// Update stats table
	t.StatsTable.Rows = t.StatsTable.Rows[:1]
	var agg *metrics.IntervalAggData
	for _, stat := range data {
		switch refreshInterval {
		case config.ShortUIRefreshInterval:
			agg = stat.AggData.Short
		case config.LongUIRefreshInterval:
			agg = stat.AggData.Long
		}

		// Update stat row in table
		t.StatsTable.Rows = append(t.StatsTable.Rows,
			[]string{stat.Url,
				fmt.Sprintf("%vs", stat.PollingInterval.Seconds()),
				strings.Join(formatStatusCodeCount(agg.StatusCodesCount), ""),
				strconv.FormatFloat(agg.Availability*100, 'f', 2, 64) + "%",
				fmt.Sprintf("%dms (%dms)", agg.ConnectDuration[0], agg.ConnectDuration[1]),
				fmt.Sprintf("%dms (%dms)", agg.FirstByteDuration[0], agg.FirstByteDuration[1]),
			})
	}

	// Colors table in different color depending on refreshInterval
	switch refreshInterval {
	case config.LongUIRefreshInterval:
		t.Title.TextStyle = ui.NewStyle(ui.ColorCyan)
		t.StatsTable.TextStyle = ui.NewStyle(ui.ColorCyan)
	case config.ShortUIRefreshInterval:
		t.Title.TextStyle = ui.Theme.Default
		t.StatsTable.TextStyle = ui.Theme.Table.Text
	}

	/* -------------------------------------------------------------------------- */
	/*                                   ALERTS                                   */
	/* -------------------------------------------------------------------------- */
	// previous number of alerts
	oldAlertRowsLen := len(t.Alerts.Rows)

	// update alerts
	for _, stat := range data {
		if stat.Alert.WebsiteWasDown {
			t.Alerts.Rows = append(t.Alerts.Rows, fmt.Sprintf("[Website %v is down. availability=%.2f, time=%v](fg:red)", stat.Url, stat.Alert.Availability, time.Now().Format("2006-01-02 15:04:05")))
		}
		if stat.Alert.WebsiteHasRecovered {
			t.Alerts.Rows = append(t.Alerts.Rows, fmt.Sprintf("[Website %v has recovered. availability=%.2f, time=%v](fg:green)", stat.Url, stat.Alert.Availability, time.Now().Format("2006-01-02 15:04:05")))
		}
	}
	// if there's new alerts scrolldown
	if len(t.Alerts.Rows) != oldAlertRowsLen {
		t.Alerts.ScrollPageDown()
	}

	// Rerender widgets
	var widgets []ui.Drawable
	widgets = append(widgets, t.Title, t.Status, t.StatsTable, t.Alerts)
	ui.Render(widgets...)

}

// formatStatusCodeCount format status code count in a better format
func formatStatusCodeCount(statusCodesMap map[int]int) []string {
	// Format status code count
	var statusCodeCount []string
	for code, count := range statusCodesMap {
		statusCodeCount = append(statusCodeCount, fmt.Sprintf("(%v: %v)", code, count))
	}
	return statusCodeCount
}

// Close shuts down UI module.
func (*UI) Close() {
	ui.Close()
}
