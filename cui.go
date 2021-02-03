package main

import (
	"fmt"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// UI is a ui implementation of UI interface.
type UI struct {
	Title      *widgets.Paragraph
	Status     *widgets.Paragraph
	StatsTable *widgets.Table
	RowStats   []*widgets.SparklineGroup
	Alerts     *widgets.List
}

// Init creates widgets, sets sizes and labels.
func (t *UI) Init() error {
	err := ui.Init()
	if err != nil {
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
				"Status code count",
				"Availability",
				"DNSLookup",
				"TCPConnection",
				"TLSHandshake",
				"ServerProcessing",
				"ContentTransfer",
				// "NameLookup",
				"Connect",
				"PreTransfer",
				"StartTransfer",
				"Total"},
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

	statsNames := []string{
		// "Status code count",
		"Availability",
		"DNSLookup",
		// "TCPConnection",
		// "TLSHandshake",
		// "ServerProcessing",
		"ContentTransfer",
		// "NameLookup",
		"Connect",
		// "PreTransfer",
		// "StartTransfer",
		"Total"}

	data := []float64{4, 2, 1, 6, 3, 9, 1, 4, 2, 15, 14, 9, 8, 6, 10, 13, 15, 12, 10, 5, 3, 6, 1, 7, 10, 10, 14, 13, 6}
	var Cols []interface{}
	for url := range urlsPollingsIntervals { //TODO: make sure we preserve urls order
		var sls []*widgets.Sparkline
		for _, stat := range statsNames {
			sl := widgets.NewSparkline()
			sl.Title = stat
			sl.Data = data
			sl.LineColor = ui.ColorCyan
			sls = append(sls, sl)
		}
		slg := widgets.NewSparklineGroup(sls...)
		slg.Title = url
		t.RowStats = append(t.RowStats, slg)
		Cols = append(Cols, ui.NewCol(1.0/float64(len(urlsPollingsIntervals)), slg))
	}

	grid := ui.NewGrid()
	grid.SetRect(0, 3, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.7, Cols...),
		ui.NewRow(0.3, t.Alerts),
	)

	ui.Render(grid)
	return nil
}

// Update updates UI widgets from UIData.
func (t *UI) Update(data *UIData, refreshInterval *time.Duration) {
	t.Title.Text = fmt.Sprintf("Monitoring %d websites, press q to quit", len(data.WebsitesStatsList))
	t.Status.Text = fmt.Sprintf("Last update: %v", data.LastTimestamp.Format(time.Stamp))

	// Update stats table with aggregated data
	t.StatsTable.Rows = t.StatsTable.Rows[:1]
	for _, stat := range data.WebsitesStatsList {
		// Update stat row in table
		t.StatsTable.Rows = append(t.StatsTable.Rows,
			[]string{stat.url,
				strings.Join(formatStatusCodeCount(stat.StatusCodesCount), ""),
				fmt.Sprintf("%.2f%%", stat.Availability*100),
				fmt.Sprintf("%dms (%dms)", stat.DNSLookup[0], stat.DNSLookup[1]),
				fmt.Sprintf("%dms (%dms)", stat.TCPConnection[0], stat.TCPConnection[1]),
				fmt.Sprintf("%dms (%dms)", stat.TLSHandshake[0], stat.TLSHandshake[1]),
				fmt.Sprintf("%dms (%dms)", stat.ServerProcessing[0], stat.ServerProcessing[1]),
				fmt.Sprintf("%dms (%dms)", stat.ContentTransfer[0], stat.ContentTransfer[1]),
				// fmt.Sprintf("%dms (%dms)", stat.NameLookup[0], stat.NameLookup[1]),
				fmt.Sprintf("%dms (%dms)", stat.Connect[0], stat.Connect[1]),
				fmt.Sprintf("%dms (%dms)", stat.PreTransfer[0], stat.PreTransfer[1]),
				fmt.Sprintf("%dms (%dms)", stat.StartTransfer[0], stat.StartTransfer[1]),
				fmt.Sprintf("%dms (%dms)", stat.Total[0], stat.Total[1]),
			})

		// Update alerts
		// Checks if website availability is below criticalAvailability for the past shortStatsHistoryInterval
		// Checks if website availability has recovered
		switch refreshInterval {
		case shortUIRefreshInterval:
			if stat.Availability < criticalAvailability {
				t.Alerts.Rows = append(t.Alerts.Rows, fmt.Sprintf("[Website %v is down. availability=%.2f, time=%v](fg:red)", stat.url, stat.Availability, time.Now().Format("2006-01-02 15:04:05")))
			}
			if stat.websiteHasRecovered == true {
				t.Alerts.Rows = append(t.Alerts.Rows, fmt.Sprintf("[Website %v has recovered. availability=%.2f, time=%v](fg:green)", stat.url, stat.Availability, time.Now().Format("2006-01-02 15:04:05")))
			}
		}
	}

	// Update stats rows with real time raw data
	for i, inspector := range data.inspectorsList { // iterate over urls
		reports := inspector.reports
		// keep only a number of reports depending on whether it's a long or short refresh
		reports = reports[len(reports)-numOfUsefulReports(inspector.url, refreshInterval):]

		sls := t.RowStats[i].Sparklines
		sls[0].Data = []float64{}
		sls[1].Data = []float64{}
		sls[2].Data = []float64{}
		sls[3].Data = []float64{}
		sls[4].Data = []float64{}
		var Availability int8
		for _, report := range reports {
			if report.StatusCode == 200 {
				Availability = 1
			}
			sls[0].Data = append(sls[0].Data, float64(Availability))
			sls[1].Data = append(sls[1].Data, float64(report.DNSLookup.Milliseconds()))
			sls[2].Data = append(sls[2].Data, float64(report.ContentTransfer.Milliseconds()))
			sls[3].Data = append(sls[3].Data, float64(report.Connect.Milliseconds()))
			sls[4].Data = append(sls[4].Data, float64(report.Total.Milliseconds()))
		}

		aggstat := data.WebsitesStatsList[i]

		sls[0].Title = "Availability: " + fmt.Sprintf("%.2f%%", aggstat.Availability*100)
		sls[1].Title = "DNSLookup: " + fmt.Sprintf("%dms (max:%dms)", aggstat.DNSLookup[0], aggstat.DNSLookup[1])
		sls[2].Title = "ContentTransfer: " + fmt.Sprintf("%dms (max:%dms)", aggstat.ContentTransfer[0], aggstat.ContentTransfer[1])
		sls[3].Title = "Connect: " + fmt.Sprintf("%dms (max:%dms)", aggstat.Connect[0], aggstat.Connect[1])
		sls[4].Title = "Total: " + fmt.Sprintf("%dms (max:%dms)", aggstat.Total[0], aggstat.Total[1])
	}

	// Colors table in different color depending on refreshInterval
	switch refreshInterval {
	case longUIRefreshInterval:
		t.Title.TextStyle = ui.NewStyle(ui.ColorCyan)
		t.StatsTable.TextStyle = ui.NewStyle(ui.ColorCyan)
	case shortUIRefreshInterval:
		t.Title.TextStyle = ui.Theme.Default
		t.StatsTable.TextStyle = ui.Theme.Table.Text
	}

	// Rerender widgets
	var widgets []ui.Drawable
	widgets = append(widgets, t.Title, t.Status, t.StatsTable, t.Alerts)

	for _, col := range t.RowStats {
		widgets = append(widgets, col)
	}
	ui.Render(widgets...)
}

func formatStatusCodeCount(statusCodesMap map[int]int) []string {
	// Format status code count
	var statusCodeCount []string
	for code, count := range statusCodesMap {
		statusCodeCount = append(statusCodeCount, fmt.Sprintf("(%v: %v)", code, count))
	}
	return statusCodeCount
}

// Close shuts down UI module.
func (t *UI) Close() {
	ui.Close()
}
