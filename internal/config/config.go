package config

import "time"

var (
	ShortUIRefreshInterval    time.Duration            // Short refreshing UI interval (in seconds)
	LongUIRefreshInterval     time.Duration            // Long refreshing UI interval (in seconds)
	ShortStatsHistoryInterval time.Duration            // Short history interval (in minutes)
	LongStatsHistoryInterval  time.Duration            // Long history interval (in minutes)
	UrlsPollingsIntervals     map[string]time.Duration // maps urls to their corresponding polling interval
	MaxHistoryPerURL          time.Duration            // max stats history duration
	CriticalAvailability      float64                  // availability of websites below which we show an alert
)
