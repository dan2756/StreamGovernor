package main

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// parseRate parses COUNT or COUNT/PERIOD, e.g. "30", "5/s" or "100/10m".
// period is zero when s does not name one.
func parseRate(s string) (count int64, period time.Duration, err error) {
	c, p, hasPeriod := strings.Cut(s, "/")
	count, err = strconv.ParseInt(c, 10, 64)
	if err != nil || count < 1 {
		return 0, 0, fmt.Errorf("invalid rate %q: count must be a positive integer", s)
	}
	if hasPeriod {
		period, err = parsePeriod(p)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid rate %q: %v", s, err)
		}
	}
	return count, period, nil
}

var periodUnits = map[string]time.Duration{
	"ms":      time.Millisecond,
	"s":       time.Second,
	"sec":     time.Second,
	"secs":    time.Second,
	"second":  time.Second,
	"seconds": time.Second,
	"m":       time.Minute,
	"min":     time.Minute,
	"mins":    time.Minute,
	"minute":  time.Minute,
	"minutes": time.Minute,
	"h":       time.Hour,
	"hr":      time.Hour,
	"hrs":     time.Hour,
	"hour":    time.Hour,
	"hours":   time.Hour,
	"d":       24 * time.Hour,
	"day":     24 * time.Hour,
	"days":    24 * time.Hour,
}

// parsePeriod parses a period such as "s", "10m", "1.5s", "30sec" or
// "1h30m". A bare unit means one of it.
func parsePeriod(s string) (time.Duration, error) {
	num, unit := s, ""
	if i := strings.IndexFunc(s, func(r rune) bool { return (r < '0' || r > '9') && r != '.' }); i >= 0 {
		num, unit = s[:i], s[i:]
	}
	if u, ok := periodUnits[strings.ToLower(unit)]; ok {
		if num == "" {
			return u, nil
		}
		d, err := scaleDecimal(num, u)
		if err != nil {
			switch {
			case errors.Is(err, errBadDecimal):
				return 0, fmt.Errorf("invalid period %q", s)
			case errors.Is(err, errNonPositive):
				return 0, fmt.Errorf("period %q must be positive", s)
			default:
				return 0, fmt.Errorf("period %q out of range", s)
			}
		}
		return d, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid period %q", s)
	}
	if d <= 0 {
		return 0, fmt.Errorf("period %q must be positive", s)
	}
	return d, nil
}

var (
	errBadDecimal  = errors.New("bad decimal")
	errNonPositive = errors.New("non-positive")
	errOutOfRange  = errors.New("out of range")
)

// scaleDecimal returns the nearest nanosecond of num*unit. num is a decimal
// such as "2.05" or ".5". Float multiplication is not used: 32.02 days is
// not exact in float64, and the extra nanosecond shows up in summaries.
func scaleDecimal(num string, unit time.Duration) (time.Duration, error) {
	r, ok := new(big.Rat).SetString(num)
	if !ok {
		return 0, errBadDecimal
	}
	if r.Sign() <= 0 {
		return 0, errNonPositive
	}
	r.Mul(r, new(big.Rat).SetInt64(int64(unit)))
	// Round half away from zero: (2n + d) / (2d).
	n := new(big.Int).Set(r.Num())
	d := new(big.Int).Set(r.Denom())
	n.Lsh(n, 1)
	n.Add(n, d)
	d.Lsh(d, 1)
	n.Quo(n, d)
	if !n.IsInt64() {
		return 0, errOutOfRange
	}
	v := n.Int64()
	if v <= 0 {
		return 0, errOutOfRange
	}
	return time.Duration(v), nil
}
