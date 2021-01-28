package main

import (
	"testing"
	"time"
)

func TestWebsiteStats_updateAlerting(t *testing.T) {
	type args struct {
		refreshInterval *time.Duration
	}
	type alert struct {
		isDown       bool
		hasRecovered bool
	}
	type want []alert

	tests := []struct {
		name           string
		availabilities []float64
		args           args
		want           want
	}{
		{"Website goes down and recovers", []float64{1.0, 0.8, 0.5, 0.2, 0.6, 0.8, 1.0}, args{refreshInterval: shortUIRefreshInterval}, want{
			{isDown: false, hasRecovered: false},
			{isDown: false, hasRecovered: false},
			{isDown: true, hasRecovered: false},
			{isDown: true, hasRecovered: false},
			{isDown: true, hasRecovered: false},
			{isDown: false, hasRecovered: true},
			{isDown: false, hasRecovered: false},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stat := &WebsiteStats{}
			var got want
			for _, av := range tt.availabilities {
				stat.Availability = av
				stat.updateAlerting(tt.args.refreshInterval)
				got = append(got, alert{isDown: stat.Availability < criticalAvailability, hasRecovered: stat.websiteHasRecovered})
			}

			for i := 0; i < len(tt.want); i++ {
				if got[i] != tt.want[i] {
					t.Errorf("{isDown, hasRecovered} = %v, want %v", got[i], tt.want[i])
				}
			}
		})
	}
}
