package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseRate(t *testing.T) {
	tests := []struct {
		in     string
		count  int64
		period time.Duration
	}{
		{"30", 30, 0},
		{"1", 1, 0},
		{"5/s", 5, time.Second},
		{"5/sec", 5, time.Second},
		{"120/m", 120, time.Minute},
		{"120/min", 120, time.Minute},
		{"100/10m", 100, 10 * time.Minute},
		{"100/10min", 100, 10 * time.Minute},
		{"3/1.5s", 3, 1500 * time.Millisecond},
		{"1/2.05s", 1, 2050 * time.Millisecond},
		{"1/2.05m", 1, 2*time.Minute + 3*time.Second},
		{"1/2.05min", 1, 2*time.Minute + 3*time.Second},
		{"1/4.1ms", 1, 4100 * time.Microsecond},
		{"1/8.2ms", 1, 8200 * time.Microsecond},
		{"1/0.57d", 1, 57 * 24 * time.Hour / 100},
		{"1/32.02d", 1, 768*time.Hour + 28*time.Minute + 48*time.Second},
		{"10/500ms", 10, 500 * time.Millisecond},
		{"1000/h", 1000, time.Hour},
		{"1000/Hour", 1000, time.Hour},
		{"50/1h30m", 50, 90 * time.Minute},
		{"7/d", 7, 24 * time.Hour},
		{"1/9223372036.854775807s", 1, time.Duration(1<<63 - 1)},
	}
	for _, tt := range tests {
		count, period, err := parseRate(tt.in)
		if err != nil {
			t.Errorf("parseRate(%q): %v", tt.in, err)
			continue
		}
		if count != tt.count || period != tt.period {
			t.Errorf("parseRate(%q) = %d, %v; want %d, %v", tt.in, count, period, tt.count, tt.period)
		}
	}
}

func TestParseRateErrors(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "count must be a positive integer"},
		{"0", "count must be a positive integer"},
		{"-5", "count must be a positive integer"},
		{"2.5", "count must be a positive integer"},
		{"abc", "count must be a positive integer"},
		{"/s", "count must be a positive integer"},
		{"5/", "invalid period"},
		{"5/fortnight", "invalid period"},
		{"5/0s", "must be positive"},
		{"5/0", "must be positive"},
		{"5/-1m", "must be positive"},
		{"5/1..5s", "invalid period"},
		{"5/106752d", "out of range"},
		{"5/106751.99116730065d", "out of range"},
		{"5/9223372036.854775808s", "out of range"},
	}
	for _, tt := range tests {
		_, _, err := parseRate(tt.in)
		if err == nil {
			t.Errorf("parseRate(%q): expected an error", tt.in)
			continue
		}
		if !strings.Contains(err.Error(), tt.want) {
			t.Errorf("parseRate(%q) error %q, want it to mention %q", tt.in, err, tt.want)
		}
	}
}
