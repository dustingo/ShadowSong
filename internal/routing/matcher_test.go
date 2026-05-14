package routing

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/datatypes"
)

func TestMatchLabels(t *testing.T) {
	tests := []struct {
		name       string
		labelsJSON []byte
		matchers   []models.LabelMatcher
		want       bool
	}{
		{
			name:       "empty matchers returns true",
			labelsJSON: []byte(`{"env":"prod","team":"backend"}`),
			matchers:   []models.LabelMatcher{},
			want:       true,
		},
		{
			name:       "exact match returns true",
			labelsJSON: []byte(`{"env":"prod","team":"backend"}`),
			matchers: []models.LabelMatcher{
				{Key: "env", Pattern: "prod"},
			},
			want: true,
		},
		{
			name:       "exact match returns false when not matching",
			labelsJSON: []byte(`{"env":"staging","team":"backend"}`),
			matchers: []models.LabelMatcher{
				{Key: "env", Pattern: "prod"},
			},
			want: false,
		},
		{
			name:       "regex match returns true",
			labelsJSON: []byte(`{"env":"production","team":"backend"}`),
			matchers: []models.LabelMatcher{
				{Key: "env", Pattern: "^prod.*$"},
			},
			want: true,
		},
		{
			name:       "regex match returns false when not matching",
			labelsJSON: []byte(`{"env":"staging","team":"backend"}`),
			matchers: []models.LabelMatcher{
				{Key: "env", Pattern: "^prod.*$"},
			},
			want: false,
		},
		{
			name:       "multiple matchers all must match",
			labelsJSON: []byte(`{"env":"prod","team":"backend","region":"us-east"}`),
			matchers: []models.LabelMatcher{
				{Key: "env", Pattern: "prod"},
				{Key: "team", Pattern: "backend"},
			},
			want: true,
		},
		{
			name:       "multiple matchers fails if one does not match",
			labelsJSON: []byte(`{"env":"prod","team":"frontend","region":"us-east"}`),
			matchers: []models.LabelMatcher{
				{Key: "env", Pattern: "prod"},
				{Key: "team", Pattern: "backend"},
			},
			want: false,
		},
		{
			name:       "missing label key returns false",
			labelsJSON: []byte(`{"env":"prod"}`),
			matchers: []models.LabelMatcher{
				{Key: "team", Pattern: "backend"},
			},
			want: false,
		},
		{
			name:       "invalid labels JSON returns false",
			labelsJSON: []byte(`{invalid json}`),
			matchers: []models.LabelMatcher{
				{Key: "env", Pattern: "prod"},
			},
			want: false,
		},
		{
			name:       "empty pattern matches any value",
			labelsJSON: []byte(`{"env":"prod"}`),
			matchers: []models.LabelMatcher{
				{Key: "env", Pattern: ""},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewMatcher(nil)
			got := matcher.MatchLabels(tt.labelsJSON, tt.matchers)
			if got != tt.want {
				t.Errorf("MatchLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsInTimeRange(t *testing.T) {
	// Helper to create time ranges JSON
	timeRangesJSON := func(trs []models.TimeRange) []byte {
		if trs == nil {
			return nil
		}
		b, _ := json.Marshal(trs)
		return b
	}

	tests := []struct {
		name          string
		timeRangesJSON []byte
		currentTime   time.Time
		want          bool
	}{
		{
			name:          "empty time ranges returns true",
			timeRangesJSON: nil,
			currentTime:   time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
			want:          true,
		},
		{
			name:          "empty array returns true",
			timeRangesJSON: []byte(`[]`),
			currentTime:   time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
			want:          true,
		},
		{
			name: "same day within range returns true",
			timeRangesJSON: timeRangesJSON([]models.TimeRange{
				{StartTime: "09:00", EndTime: "17:00"},
			}),
			currentTime: time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
			want:        true,
		},
		{
			name: "same day outside range returns false",
			timeRangesJSON: timeRangesJSON([]models.TimeRange{
				{StartTime: "09:00", EndTime: "17:00"},
			}),
			currentTime: time.Date(2024, 1, 1, 18, 30, 0, 0, time.UTC),
			want:        false,
		},
		{
			name: "cross-day within range (after midnight) returns true",
			timeRangesJSON: timeRangesJSON([]models.TimeRange{
				{StartTime: "22:00", EndTime: "06:00"},
			}),
			currentTime: time.Date(2024, 1, 1, 23, 30, 0, 0, time.UTC),
			want:        true,
		},
		{
			name: "cross-day within range (before midnight) returns true",
			timeRangesJSON: timeRangesJSON([]models.TimeRange{
				{StartTime: "22:00", EndTime: "06:00"},
			}),
			currentTime: time.Date(2024, 1, 1, 5, 30, 0, 0, time.UTC),
			want:        true,
		},
		{
			name: "cross-day outside range returns false",
			timeRangesJSON: timeRangesJSON([]models.TimeRange{
				{StartTime: "22:00", EndTime: "06:00"},
			}),
			currentTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			want:        false,
		},
		{
			name: "multiple ranges matches any returns true",
			timeRangesJSON: timeRangesJSON([]models.TimeRange{
				{StartTime: "09:00", EndTime: "12:00"},
				{StartTime: "14:00", EndTime: "18:00"},
			}),
			currentTime: time.Date(2024, 1, 1, 15, 30, 0, 0, time.UTC),
			want:        true,
		},
		{
			name: "boundary at start time returns true",
			timeRangesJSON: timeRangesJSON([]models.TimeRange{
				{StartTime: "09:00", EndTime: "17:00"},
			}),
			currentTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			want:        true,
		},
		{
			name: "boundary at end time returns true",
			timeRangesJSON: timeRangesJSON([]models.TimeRange{
				{StartTime: "09:00", EndTime: "17:00"},
			}),
			currentTime: time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC),
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewMatcher(nil)
			matcher.now = func() time.Time { return tt.currentTime }
			got := matcher.IsInTimeRange(tt.timeRangesJSON)
			if got != tt.want {
				t.Errorf("IsInTimeRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindMatchedTargets(t *testing.T) {
	// This test verifies batch loading of channels
	// We use a mock database to verify that channels are loaded in a single query

	t.Run("batch loads channels for matched rules", func(t *testing.T) {
		alert := &models.Alert{
			AlertID:   "test-alert-1",
			Source:    "prometheus",
			Severity:  "P1",
			AlertName: "HighCPU",
			Labels:    datatypes.JSON(`{"env":"prod","team":"backend"}`),
		}

		rules := []models.RouteRule{
			{
				ID:            1,
				Name:          "prod-backend",
				Priority:      0,
				Enabled:       true,
				Sources:       datatypes.JSON(`["prometheus"]`),
				Severities:    datatypes.JSON(`["P1","P2"]`),
				LabelMatchers: datatypes.JSON(`[{"key":"env","pattern":"prod"},{"key":"team","pattern":"backend"}]`),
				ChannelIDs:    datatypes.JSON(`[1,2,3]`),
			},
		}

		// Create mock channels
		channel1 := models.Channel{ID: 1, Name: "channel-1", Type: "feishu", Enabled: true}
		channel2 := models.Channel{ID: 2, Name: "channel-2", Type: "dingtalk", Enabled: true}
		channel3 := models.Channel{ID: 3, Name: "channel-3", Type: "webhook", Enabled: false} // disabled

		mockDB := &mockDB{
			channels: []models.Channel{channel1, channel2, channel3},
		}

		matcher := NewMatcher(mockDB)
		targets, err := matcher.FindMatchedTargets(alert, rules)
		if err != nil {
			t.Fatalf("FindMatchedTargets() error = %v", err)
		}

		// Should return 2 enabled channels
		if len(targets) != 2 {
			t.Errorf("FindMatchedTargets() returned %d targets, want 2", len(targets))
		}

		// Verify batch query was used (not N+1)
		if mockDB.queryCount != 1 {
			t.Errorf("Expected 1 batch query, got %d queries (N+1 problem)", mockDB.queryCount)
		}

		// Verify IN clause was used
		if !mockDB.usedINClause {
			t.Error("Expected batch query with IN clause, but it was not used")
		}
	})

	t.Run("returns empty when no rules match", func(t *testing.T) {
		alert := &models.Alert{
			AlertID:   "test-alert-2",
			Source:    "grafana",
			Severity:  "P3",
			AlertName: "LowDisk",
			Labels:    datatypes.JSON(`{"env":"dev"}`),
		}

		rules := []models.RouteRule{
			{
				ID:         1,
				Name:       "prod-only",
				Priority:   0,
				Enabled:    true,
				Sources:    datatypes.JSON(`["prometheus"]`),
				Severities: datatypes.JSON(`["P1","P2"]`),
				ChannelIDs: datatypes.JSON(`[1]`),
			},
		}

		mockDB := &mockDB{channels: []models.Channel{{ID: 1, Name: "ch1", Type: "feishu", Enabled: true}}}
		matcher := NewMatcher(mockDB)
		targets, err := matcher.FindMatchedTargets(alert, rules)
		if err != nil {
			t.Fatalf("FindMatchedTargets() error = %v", err)
		}

		if len(targets) != 0 {
			t.Errorf("FindMatchedTargets() returned %d targets, want 0", len(targets))
		}
	})

	t.Run("stops at first matching rule", func(t *testing.T) {
		alert := &models.Alert{
			AlertID:   "test-alert-3",
			Source:    "prometheus",
			Severity:  "P1",
			AlertName: "HighCPU",
			Labels:    datatypes.JSON(`{}`),
		}

		rules := []models.RouteRule{
			{
				ID:         1,
				Name:       "first-rule",
				Priority:   0,
				Enabled:    true,
				Sources:    datatypes.JSON(`["prometheus"]`),
				Severities: datatypes.JSON(`["P1"]`),
				ChannelIDs: datatypes.JSON(`[1]`),
			},
			{
				ID:         2,
				Name:       "second-rule",
				Priority:   1,
				Enabled:    true,
				Sources:    datatypes.JSON(`["prometheus"]`),
				Severities: datatypes.JSON(`["P1"]`),
				ChannelIDs: datatypes.JSON(`[2]`),
			},
		}

		mockDB := &mockDB{
			channels: []models.Channel{
				{ID: 1, Name: "ch1", Type: "feishu", Enabled: true},
				{ID: 2, Name: "ch2", Type: "feishu", Enabled: true},
			},
		}
		matcher := NewMatcher(mockDB)
		targets, err := matcher.FindMatchedTargets(alert, rules)
		if err != nil {
			t.Fatalf("FindMatchedTargets() error = %v", err)
		}

		// Should only return channels from first matching rule
		if len(targets) != 1 {
			t.Errorf("FindMatchedTargets() returned %d targets, want 1", len(targets))
		}
		if len(targets) > 0 && targets[0].Channel.ID != 1 {
			t.Errorf("Expected channel 1, got channel %d", targets[0].Channel.ID)
		}
	})
}

func TestRegexCaching(t *testing.T) {
	matcher := NewMatcher(nil)

	// Compile same pattern multiple times
	pattern := "^prod.*$"
	re1, err := matcher.getCompiledRegex(pattern)
	if err != nil {
		t.Fatalf("getCompiledRegex() error = %v", err)
	}
	re2, _ := matcher.getCompiledRegex(pattern)

	// Should return same cached instance
	if re1 != re2 {
		t.Error("Expected cached regex to be reused")
	}

	// Different pattern should return different instance
	re3, _ := matcher.getCompiledRegex("^staging.*$")
	if re1 == re3 {
		t.Error("Expected different regex instances for different patterns")
	}
}

// mockDB implements dbLoader interface for testing
type mockDB struct {
	channels     []models.Channel
	queryCount   int
	usedINClause bool
}

func (m *mockDB) Find(dst interface{}, conds ...interface{}) error {
	m.queryCount++

	// Check if this is a channel query with IN clause
	if channels, ok := dst.(*[]models.Channel); ok {
		if len(conds) > 0 {
			// Simulate IN clause query
			m.usedINClause = true
			// Filter channels by the IDs in the query
			// The conds format is: "id IN ?", []uint{1, 2, 3}
			if len(conds) >= 2 {
				if ids, ok := conds[1].([]uint); ok {
					var filtered []models.Channel
					for _, ch := range m.channels {
						for _, id := range ids {
							if ch.ID == id {
								filtered = append(filtered, ch)
								break
							}
						}
					}
					*channels = filtered
					return nil
				}
			}
			*channels = m.channels
		}
	}
	return nil
}
