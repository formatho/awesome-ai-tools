package models

import "time"

// ABTestConversion represents a conversion event in an A/B test
type ABTestConversion struct {
	ID          string    `json:"id"`
	TestID      string    `json:"test_id"`
	VariantID   string    `json:"variant_id"`
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	UserID      string    `json:"user_id,omitempty"`
	SessionID   string    `json:"session_id,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	UserAgent   string    `json:"user_agent,omitempty"`
	PageURL     string    `json:"page_url,omitempty"`
}

// ABTestResults represents aggregated results for a test
type ABTestResults struct {
	TestID     string                   `json:"test_id"`
	TestName   string                   `json:"test_name"`
	StartDate  time.Time                `json:"start_date"`
	EndDate    *time.Time               `json:"end_date,omitempty"`
	Status     string                   `json:"status"`
	Variants   []ABTestVariantResult    `json:"variants"`
	Winner     *string                  `json:"winner,omitempty"`
	Confidence float64                  `json:"confidence"`
	TotalUsers int                      `json:"total_users"`
	PrimaryMetric string                `json:"primary_metric"`
}

// ABTestVariantResult represents results for a single variant
type ABTestVariantResult struct {
	VariantID      string             `json:"variant_id"`
	VariantName    string             `json:"variant_name"`
	Visitors       int                `json:"visitors"`
	Conversions    int                `json:"conversions"`
	ConversionRate float64            `json:"conversion_rate"`
	Metrics        map[string]float64 `json:"metrics"`
	Confidence     float64            `json:"confidence"`
	IsWinner       bool               `json:"is_winner"`
}
