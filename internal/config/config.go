package config

import "time"

var (
	ShortUIRefreshInterval    time.Duration                    // Short refreshing UI interval (in seconds)
	LongUIRefreshInterval     time.Duration                    // Long refreshing UI interval (in seconds)
	ShortStatsHistoryInterval time.Duration                    // Short history interval (in minutes)
	LongStatsHistoryInterval  time.Duration                    // Long history interval (in minutes)
	WebsiteAlertInterval      time.Duration                    // Shows alert if website is down for `WebsiteAlertInterval` minutes
	UrlsPollingsIntervals     = make(map[string]time.Duration) // maps urls to their corresponding polling interval
	CriticalAvailability      = 0.8                            // availability of websites below which we show an alert
)
