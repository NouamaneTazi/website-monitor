package config

import "time"

var (
	ShortUIRefreshInterval    time.Duration                    // Short refreshing UI interval (in seconds)
	LongUIRefreshInterval     time.Duration                    // Long refreshing UI interval (in seconds)
	ShortStatsHistoryInterval time.Duration                    // Short history interval (in minutes)
	LongStatsHistoryInterval  time.Duration                    // Long history interval (in minutes)
	UrlsPollingsIntervals     = make(map[string]time.Duration) // maps urls to their corresponding polling interval
	MaxHistoryPerURL          = 1 * time.Minute                // max stats history duration
	CriticalAvailability      = 0.8                            // availability of websites below which we show an alert
	EnableCUI                 = false                          // whether to enable CUI
)
