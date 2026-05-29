package retention

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// CleanupResult reports how many records were deleted.
type CleanupResult struct {
	AlertsDeleted      int64
	DeliveriesDeleted  int64
	AttemptsDeleted    int64
	RecoveriesDeleted  int64
}

// Cleanup deletes records older than retentionDays. Returns immediately if retentionDays <= 0.
func Cleanup(db *gorm.DB, retentionDays int) CleanupResult {
	if retentionDays <= 0 {
		return CleanupResult{}
	}

	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	var result CleanupResult

	res := db.Exec("DELETE FROM notification_delivery_attempts WHERE created_at < ?", cutoff)
	result.AttemptsDeleted = res.RowsAffected

	res = db.Exec("DELETE FROM notification_deliveries WHERE created_at < ?", cutoff)
	result.DeliveriesDeleted = res.RowsAffected

	res = db.Exec("DELETE FROM notification_delivery_recoveries WHERE created_at < ?", cutoff)
	result.RecoveriesDeleted = res.RowsAffected

	res = db.Exec("DELETE FROM alerts WHERE created_at < ?", cutoff)
	result.AlertsDeleted = res.RowsAffected

	log.Printf("retention: cleaned alerts=%d deliveries=%d attempts=%d recoveries=%d (retention_days=%d)",
		result.AlertsDeleted, result.DeliveriesDeleted, result.AttemptsDeleted, result.RecoveriesDeleted, retentionDays)

	return result
}

// Run starts the retention cleanup loop. It blocks until stop is closed.
func Run(db *gorm.DB, retentionDays int, interval time.Duration, stop <-chan struct{}) {
	Cleanup(db, retentionDays)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			Cleanup(db, retentionDays)
		case <-stop:
			return
		}
	}
}
