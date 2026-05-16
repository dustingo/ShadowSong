package stats

import (
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/gorm"
)

// TrendPoint represents a single point in the alert trend.
type TrendPoint struct {
	Time  time.Time `json:"time"`
	Count int       `json:"count"`
}

// AlertStats contains aggregated statistics about alerts.
type AlertStats struct {
	Total      int            `json:"total"`
	Firing     int            `json:"firing"`
	Acked      int            `json:"acked"`
	Silenced   int            `json:"silenced"`
	BySeverity map[string]int `json:"by_severity"`
	Trend      []TrendPoint   `json:"trend"`
}

// GetAlertStats retrieves aggregated alert statistics using optimized GROUP BY queries.
// This reduces the query count from 32 individual queries to just 3 aggregated queries.
func GetAlertStats(db *gorm.DB) (*AlertStats, error) {
	stats := &AlertStats{
		BySeverity: make(map[string]int),
		Trend:      make([]TrendPoint, 0, 24),
	}

	// Initialize all severity levels to ensure they're present even if 0
	for _, severity := range models.ValidSeverities {
		stats.BySeverity[severity] = 0
	}

	// Query 1: Get counts by status using GROUP BY
	if err := getStatusCounts(db, stats); err != nil {
		return nil, err
	}

	// Query 2: Get counts by severity (firing only) using GROUP BY
	if err := getSeverityCounts(db, stats); err != nil {
		return nil, err
	}

	// Query 3: Get hourly trend for last 24 hours using GROUP BY
	if err := getHourlyTrend(db, stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// statusCount is a helper struct for status count queries.
type statusCount struct {
	Status string
	Count  int
}

// severityCount is a helper struct for severity count queries.
type severityCount struct {
	Severity string
	Count    int
}

// hourCount is a helper struct for hourly trend queries.
type hourCount struct {
	Hour  string // Scan as string for SQLite compatibility
	Count int
}

// getStatusCounts retrieves alert counts grouped by status.
func getStatusCounts(db *gorm.DB, stats *AlertStats) error {
	var results []statusCount

	// Query all status counts in a single GROUP BY query
	if err := db.Model(&models.Alert{}).
		Select("status as status, count(*) as count").
		Group("status").
		Scan(&results).Error; err != nil {
		return err
	}

	// Process results
	for _, r := range results {
		stats.Total += r.Count
		switch r.Status {
		case "firing":
			stats.Firing = r.Count
		case "acked":
			stats.Acked = r.Count
		case "silenced":
			stats.Silenced = r.Count
		}
	}

	return nil
}

// getSeverityCounts retrieves alert counts grouped by severity (firing alerts only).
func getSeverityCounts(db *gorm.DB, stats *AlertStats) error {
	var results []severityCount

	// Query severity counts for firing alerts in a single GROUP BY query
	if err := db.Model(&models.Alert{}).
		Select("severity as severity, count(*) as count").
		Where("status = ?", "firing").
		Group("severity").
		Scan(&results).Error; err != nil {
		return err
	}

	// Process results - update the pre-initialized map
	for _, r := range results {
		stats.BySeverity[r.Severity] = r.Count
	}

	return nil
}

// parseHourString parses a timestamp string returned by PostgreSQL into a time.Time.
// pgx may return timestamps in different formats depending on the protocol mode:
//   - Simple protocol with text format: "2006-01-02 15:04:05" (space separator)
//   - pgx codec scanning into string:   "2006-01-02T15:04:05Z" (RFC3339)
//   - With timezone suffix:              "2006-01-02 15:04:05-07:00"
func parseHourString(s string) (time.Time, bool) {
	for _, layout := range []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 UTC
		"2006-01-02 15:04:05",       // PostgreSQL text format
		"2006-01-02 15:04:05-07:00", // PostgreSQL text with tz
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// getHourlyTrend retrieves alert counts grouped by hour for the last 24 hours.
// All internal calculations use UTC to ensure consistent key matching regardless
// of server timezone or how trigger_time was originally stored.
func getHourlyTrend(db *gorm.DB, stats *AlertStats) error {
	// Calculate time range: last 24 hours in UTC
	// Start from 23 hours ago (aligned to hour boundary) so the 24th slot is the current hour.
	nowUTC := time.Now().UTC()
	startTime := nowUTC.Add(-23 * time.Hour)
	// Align to hour boundary
	startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), startTime.Hour(), 0, 0, 0, time.UTC)

	dialect := db.Dialector.Name()

	// Build a map of UTC-hour -> count
	countMap := make(map[time.Time]int)

	if dialect == "sqlite" {
		// SQLite: fetch raw trigger_time values and group in Go.
		var alerts []struct {
			TriggerTime time.Time
		}
		if err := db.Model(&models.Alert{}).
			Select("trigger_time").
			Where("trigger_time >= ?", startTime).
			Find(&alerts).Error; err != nil {
			return err
		}
		for _, a := range alerts {
			utcHour := time.Date(a.TriggerTime.Year(), a.TriggerTime.Month(), a.TriggerTime.Day(), a.TriggerTime.Hour(), 0, 0, 0, time.UTC)
			countMap[utcHour]++
		}
	} else {
		// PostgreSQL: use AT TIME ZONE 'UTC' to force UTC truncation regardless of
		// session timezone.
		var results []hourCount
		query := db.Model(&models.Alert{}).
			Select("date_trunc('hour', trigger_time AT TIME ZONE 'UTC') as hour, count(*) as count").
			Where("trigger_time >= ?", startTime).
			Group("hour").
			Order("hour asc")
		if err := query.Scan(&results).Error; err != nil {
			return err
		}
		for _, r := range results {
			parsed, ok := parseHourString(r.Hour)
			if !ok {
				continue
			}
			key := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), parsed.Hour(), 0, 0, 0, time.UTC)
			countMap[key] = r.Count
		}
	}

	// Generate all 24 hours in UTC, filling in zeros for missing hours
	stats.Trend = make([]TrendPoint, 24)
	for i := 0; i < 24; i++ {
		hourTime := startTime.Add(time.Duration(i) * time.Hour)
		stats.Trend[i] = TrendPoint{
			Time:  hourTime,
			Count: countMap[hourTime],
		}
	}

	return nil
}