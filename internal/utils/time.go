package utils

import (
	"strconv"
	"strings"
)

// ParseTimeToMinutes parses a time string in "HH:mm" format to minutes since midnight.
func ParseTimeToMinutes(timeStr string) int {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}
	hour, _ := strconv.Atoi(parts[0])
	minute, _ := strconv.Atoi(parts[1])
	return hour*60 + minute
}
