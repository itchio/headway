package united

import (
	"fmt"
	"strings"
	"time"
)

// FormatBytes converts bytes to human readable string. Like 2 MiB, 64.2 KiB, 52 B
func FormatBytes(i int64) (result string) {
	switch {
	case i > (1024 * 1024 * 1024 * 1024):
		result = fmt.Sprintf("%.02f TiB", float64(i)/1024/1024/1024/1024)
	case i > (1024 * 1024 * 1024):
		result = fmt.Sprintf("%.02f GiB", float64(i)/1024/1024/1024)
	case i > (1024 * 1024):
		result = fmt.Sprintf("%.02f MiB", float64(i)/1024/1024)
	case i > 1024:
		result = fmt.Sprintf("%.02f KiB", float64(i)/1024)
	default:
		result = fmt.Sprintf("%d B", i)
	}
	result = strings.Trim(result, " ")
	return
}

// FormatBPSValue formats a bandwidth value, ie. a number of bytes per second
func FormatBPSValue(bps float64) string {
	return fmt.Sprintf("%s/s", FormatBytes(int64(bps)))
}

// FormatBPS formats a bandwidth value, given the number of bytes processed
// over a certain duration
func FormatBPS(size int64, duration time.Duration) string {
	return FormatBPSValue(float64(size) / duration.Seconds())
}

// FormatDuration formats a duration
func FormatDuration(d time.Duration) string {
	res := ""
	if d > time.Hour*24 {
		res = fmt.Sprintf("%dd", d/24/time.Hour)
		d -= (d / time.Hour / 24) * (time.Hour * 24)
	}
	return fmt.Sprintf("%s%v ", res, d)
}
