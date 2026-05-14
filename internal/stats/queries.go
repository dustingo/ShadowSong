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
	Total      int                `json:"total"`
	Firing     int                `json:"firing"`
	Acked      int                `json:"acked"`
	Silenced   int                `json:"silenced"`
	BySeverity map[string]int     `json:"by_severity"`
	Trend      []TrendPoint       `json:"trend"`
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

// getHourlyTrend retrieves alert counts grouped by hour for the last 24 hours.
func getHourlyTrend(db *gorm.DB, stats *AlertStats) error {
	var results []hourCount

	// Calculate time range: last 24 hours
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)

	// Use a database-agnostic approach for hour truncation
	// SQLite: strftime('%Y-%m-%d %H:00:00', trigger_time)
	// PostgreSQL: date_trunc('hour', trigger_time)
	// We detect the dialect and use appropriate syntax
	dialect := db.Dialector.Name()

	var query *gorm.DB
	if dialect == "sqlite" {
		query = db.Model(&models.Alert{}).
			Select("strftime('%Y-%m-%d %H:00:00', trigger_time) as hour, count(*) as count").
			Where("trigger_time >= ?", startTime).
			Group("hour").
			Order("hour asc")
	} else {
		// PostgreSQL and others that support date_trunc
		query = db.Model(&models.Alert{}).
			Select("date_trunc('hour', trigger_time) as hour, count(*) as count").
			Where("trigger_time >= ?", startTime).
			Group("hour").
			Order("hour asc")
	}

	if err := query.Scan(&results).Error; err != nil {
		return err
	}

	// Build a map of hour -> count for easy lookup
	countMap := make(map[string]int)
	for _, r := range results {
		// r.Hour is already a string in format "2006-01-02 15:00:00"
		countMap[r.Hour] = r.Count
	}

	// Generate all 24 hours, filling in zeros for missing hours
	stats.Trend = make([]TrendPoint, 24)
	for i := 0; i < 24; i++ {
		hourTime := startTime.Add(time.Duration(i) * time.Hour)
		// Truncate to hour boundary
		hourTime = time.Date(hourTime.Year(), hourTime.Month(), hourTime.Day(), hourTime.Hour(), 0, 0, 0, hourTime.Location())
		hourKey := hourTime.Format("2006-01-02 15:00:00")

		stats.Trend[i] = TrendPoint{
			Time:  hourTime,
			Count: countMap[hourKey],
		}
	}

	return nil
}
