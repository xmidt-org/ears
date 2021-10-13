package sharder

import (
	"strings"
	"time"
)

// Timestamp formats into the one true time format
func Timestamp(t time.Time) string {
	ts := t.UTC().Format(time.RFC3339Nano)
	//Format library cuts off trailing 0s. See: https://github.com/golang/go/issues/19635
	//This cause probably if we want to compare natural ordering of the string.
	//Artificially add the precision back
	if !strings.Contains(ts, ".") {
		ts = ts[0:len(ts)-1] + ".000000000Z"
	}
	return ts
}

func ParseTime(timestamp string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, timestamp)
}
