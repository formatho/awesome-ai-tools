package cron

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

// testConfig creates a test configuration.
func testConfig() Config {
	return Config{
		DBPath:            ":memory:",
		DefaultTimezone:   "UTC",
		RetryOnFailure:    false,
		MaxRetries:        3,
		MissedRunBehavior: "skip",
	}
}

// testJob creates a test job.
func testJob(id string) Job {
	return Job{
		ID:       id,
		Name:     "Test Job",
		Schedule: "0 9 * * *",
		Timezone: "UTC",
		AgentID:  "test-agent",
		Enabled:  true,
		TODO: map[string]interface{}{
			"description": "Test task",
		},
	}
}

// TestParser tests cron expression parsing.
func TestParser(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		expr     string
		wantErr  bool
		wantType string
	}{
		{"standard cron", "0 9 * * *", false, "standard"},
		{"hourly alias", "@hourly", false, "alias"},
		{"daily alias", "@daily", false, "alias"},
		{"weekly alias", "@weekly", false, "alias"},
		{"monthly alias", "@monthly", false, "alias"},
		{"every minute", "* * * * *", false, "standard"},
		{"specific time", "30 14 * * 1-5", false, "standard"},
		{"empty", "", true, ""},
		{"invalid", "invalid", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && parsed.Type != tt.wantType {
				t.Errorf("Parse() type = %v, want %v", parsed.Type, tt.wantType)
			}
		})
	}
}

// TestParserNextRun tests next run calculation.
func TestParserNextRun(t *testing.T) {
	parser := NewParser()

	// Test @daily
	parsed, err := parser.Parse("@daily")
	if err != nil {
		t.Fatalf("Failed to parse @daily: %v", err)
	}

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	next := parsed.Next(now)

	// Next run should be tomorrow at midnight
	expected := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next() = %v, want %v", next, expected)
	}
}

// TestParserWithTimezone tests timezone handling.
func TestParserWithTimezone(t *testing.T) {
	parser := NewParser()

	parsed, err := parser.Parse("0 9 * * *")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Create a time in New York timezone
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Date(2024, 1, 15, 8, 0, 0, 0, loc)

	next := parsed.NextInLocation(now, loc)

	// Next run should be 9 AM New York time
	if next.Hour() != 9 {
		t.Errorf("Next() hour = %d, want 9", next.Hour())
	}
}

// TestScheduleBuilder tests the fluent schedule builder.
func TestScheduleBuilder(t *testing.T) {
	tests := []struct {
		name     string
		builder  func(*ScheduleBuilder) *ScheduleBuilder
		expected string
	}{
		{
			name:     "daily at 9am",
			builder:  func(b *ScheduleBuilder) *ScheduleBuilder { return b.DailyAt(9, 0) },
			expected: "0 9 * * *",
		},
		{
			name:     "weekly on Monday at 10:30",
			builder:  func(b *ScheduleBuilder) *ScheduleBuilder { return b.WeeklyOn(1, 10, 30) },
			expected: "30 10 * * 1",
		},
		{
			name:     "monthly on 15th at noon",
			builder:  func(b *ScheduleBuilder) *ScheduleBuilder { return b.MonthlyOn(15, 12, 0) },
			expected: "0 12 15 * *",
		},
		{
			name:     "every hour",
			builder:  func(b *ScheduleBuilder) *ScheduleBuilder { return b.EveryHour() },
			expected: "0 * * * *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewScheduleBuilder()
			result := tt.builder(builder).Build()
			if result != tt.expected {
				t.Errorf("Build() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestJobValidation tests job validation.
func TestJobValidation(t *testing.T) {
	tests := []struct {
		name    string
		job     Job
		wantErr bool
	}{
		{"valid job", testJob("test"), false},
		{"empty ID", Job{Schedule: "@daily", AgentID: "agent"}, true},
		{"empty schedule", Job{ID: "test", AgentID: "agent"}, true},
		{"empty agent", Job{ID: "test", Schedule: "@daily"}, true},
		{"invalid schedule", Job{ID: "test", Schedule: "invalid", AgentID: "agent"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfigValidation tests configuration validation.
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{"valid config", testConfig(), false},
		{"negative retries", Config{DBPath: ":memory:", MaxRetries: -1}, true},
		{"invalid missed run", Config{DBPath: ":memory:", MissedRunBehavior: "invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHistoryStore tests the history storage.
func TestHistoryStore(t *testing.T) {
	store, err := NewHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create history store: %v", err)
	}
	defer store.Close()

	// Test adding a run
	history := &RunHistory{
		JobID:        "test-job",
		ScheduledFor: time.Now(),
		StartedAt:    time.Now(),
		Status:       StatusSuccess,
		Result:       map[string]string{"output": "done"},
	}

	err = store.AddRun(history)
	if err != nil {
		t.Fatalf("Failed to add run: %v", err)
	}

	if history.ID == 0 {
		t.Error("Expected ID to be set after AddRun")
	}

	// Test getting history
	histories, err := store.GetHistory("test-job", 10)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(histories) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(histories))
	}

	// Test stats
	stats, err := store.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats["total_runs"].(int64) != 1 {
		t.Error("Expected total_runs to be 1")
	}
}

// TestHistoryStoreJobPersistence tests job persistence.
func TestHistoryStoreJobPersistence(t *testing.T) {
	store, err := NewHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create history store: %v", err)
	}
	defer store.Close()

	// Create and save job
	job := testJob("persist-test")
	err = store.SaveJob(&job)
	if err != nil {
		t.Fatalf("Failed to save job: %v", err)
	}

	// Load job
	loaded, err := store.LoadJob("persist-test")
	if err != nil {
		t.Fatalf("Failed to load job: %v", err)
	}

	if loaded.ID != job.ID {
		t.Errorf("Loaded job ID = %v, want %v", loaded.ID, job.ID)
	}

	// Load all jobs
	jobs, err := store.LoadAllJobs()
	if err != nil {
		t.Fatalf("Failed to load all jobs: %v", err)
	}

	if len(jobs) != 1 {
		t.Errorf("Expected 1 job, got %d", len(jobs))
	}

	// Delete job
	err = store.DeleteJob("persist-test")
	if err != nil {
		t.Fatalf("Failed to delete job: %v", err)
	}

	// Verify deleted
	_, err = store.LoadJob("persist-test")
	if err == nil {
		t.Error("Expected error when loading deleted job")
	}
}

// TestScheduler tests basic scheduler operations.
func TestScheduler(t *testing.T) {
	config := testConfig()
	scheduler, err := NewScheduler(config)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}
	defer scheduler.Close()

	// Add job
	job := testJob("scheduler-test")
	err = scheduler.AddJob(job)
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Get job
	retrieved, err := scheduler.GetJob("scheduler-test")
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrieved.ID != job.ID {
		t.Errorf("Retrieved job ID = %v, want %v", retrieved.ID, job.ID)
	}

	// Get all jobs
	jobs := scheduler.GetJobs()
	if len(jobs) != 1 {
		t.Errorf("Expected 1 job, got %d", len(jobs))
	}

	// Remove job
	err = scheduler.RemoveJob("scheduler-test")
	if err != nil {
		t.Fatalf("Failed to remove job: %v", err)
	}

	// Verify removed
	jobs = scheduler.GetJobs()
	if len(jobs) != 0 {
		t.Errorf("Expected 0 jobs after removal, got %d", len(jobs))
	}
}

// TestSchedulerStartStop tests starting and stopping the scheduler.
func TestSchedulerStartStop(t *testing.T) {
	config := testConfig()
	scheduler, err := NewScheduler(config)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Add a job before starting
	job := testJob("start-stop-test")
	scheduler.AddJob(job)

	// Start
	err = scheduler.Start()
	if err != nil {
		t.Fatalf("Failed to start scheduler: %v", err)
	}

	if !scheduler.IsRunning() {
		t.Error("Expected scheduler to be running")
	}

	// Try to start again
	err = scheduler.Start()
	if err == nil {
		t.Error("Expected error when starting already running scheduler")
	}

	// Stop
	err = scheduler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}

	if scheduler.IsRunning() {
		t.Error("Expected scheduler to be stopped")
	}

	scheduler.Close()
}

// TestSchedulerPauseResume tests pausing and resuming jobs.
func TestSchedulerPauseResume(t *testing.T) {
	config := testConfig()
	scheduler, err := NewScheduler(config)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}
	defer scheduler.Close()

	job := testJob("pause-resume-test")
	scheduler.AddJob(job)

	// Pause
	err = scheduler.PauseJob("pause-resume-test")
	if err != nil {
		t.Fatalf("Failed to pause job: %v", err)
	}

	retrieved, _ := scheduler.GetJob("pause-resume-test")
	if retrieved.Enabled {
		t.Error("Expected job to be disabled after pause")
	}

	// Resume
	err = scheduler.ResumeJob("pause-resume-test")
	if err != nil {
		t.Fatalf("Failed to resume job: %v", err)
	}

	retrieved, _ = scheduler.GetJob("pause-resume-test")
	if !retrieved.Enabled {
		t.Error("Expected job to be enabled after resume")
	}
}

// TestSchedulerExecution tests job execution.
func TestSchedulerExecution(t *testing.T) {
	config := testConfig()

	var execCount int64

	scheduler, err := NewScheduler(config, WithExecutor(func(ctx context.Context, job *Job) (interface{}, error) {
		atomic.AddInt64(&execCount, 1)
		return "executed", nil
	}))
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}
	defer scheduler.Close()

	// Add a job that runs every second
	job := Job{
		ID:       "exec-test",
		Name:     "Execution Test",
		Schedule: "@every 1s",
		AgentID:  "test-agent",
		Enabled:  true,
	}
	scheduler.AddJob(job)

	// Start scheduler
	scheduler.Start()
	defer scheduler.Stop()

	// Wait for execution
	time.Sleep(2 * time.Second)

	// Check execution count
	if atomic.LoadInt64(&execCount) < 1 {
		t.Errorf("Expected at least 1 execution, got %d", execCount)
	}

	// Check history
	history, err := scheduler.GetHistory("exec-test", 10)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) < 1 {
		t.Error("Expected at least 1 history entry")
	}
}

// TestSchedulerRetry tests retry on failure.
func TestSchedulerRetry(t *testing.T) {
	config := testConfig()
	config.RetryOnFailure = true
	config.MaxRetries = 2

	var attempts int64

	scheduler, err := NewScheduler(config, WithExecutor(func(ctx context.Context, job *Job) (interface{}, error) {
		attempt := atomic.AddInt64(&attempts, 1)
		if attempt < 3 {
			return nil, errors.New("simulated failure")
		}
		return "success", nil
	}))
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}
	defer scheduler.Close()

	// Add job
	job := Job{
		ID:       "retry-test",
		Name:     "Retry Test",
		Schedule: "@every 1h", // Use hourly to avoid auto-triggering
		AgentID:  "test-agent",
		Enabled:  true,
	}
	scheduler.AddJob(job)

	// Trigger immediate execution
	scheduler.RunJob("retry-test")

	// Wait for retries (1s + 2s backoff = 3s + execution time)
	time.Sleep(3500 * time.Millisecond)

	// Should have attempted 3 times
	if atomic.LoadInt64(&attempts) != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

// TestSchedulerRunJob tests manual job triggering.
func TestSchedulerRunJob(t *testing.T) {
	config := testConfig()

	var executed int64

	scheduler, err := NewScheduler(config, WithExecutor(func(ctx context.Context, job *Job) (interface{}, error) {
		atomic.StoreInt64(&executed, 1)
		return "done", nil
	}))
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}
	defer scheduler.Close()

	// Add job
	job := testJob("manual-run-test")
	scheduler.AddJob(job)

	// Trigger manual run
	err = scheduler.RunJob("manual-run-test")
	if err != nil {
		t.Fatalf("Failed to run job: %v", err)
	}

	// Wait for execution
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&executed) != 1 {
		t.Error("Expected job to be executed")
	}
}

// TestSchedulerGetStats tests getting scheduler statistics.
func TestSchedulerGetStats(t *testing.T) {
	config := testConfig()
	scheduler, err := NewScheduler(config)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}
	defer scheduler.Close()

	// Add some jobs
	scheduler.AddJob(testJob("stats-test-1"))
	scheduler.AddJob(testJob("stats-test-2"))

	// Get stats
	stats, err := scheduler.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats["total_jobs"].(int) != 2 {
		t.Errorf("Expected 2 total jobs, got %v", stats["total_jobs"])
	}
}

// TestSchedulerUpdateJob tests updating a job.
func TestSchedulerUpdateJob(t *testing.T) {
	config := testConfig()
	scheduler, err := NewScheduler(config)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}
	defer scheduler.Close()

	// Add job
	job := testJob("update-test")
	scheduler.AddJob(job)

	// Update job
	updatedJob := Job{
		ID:       "update-test",
		Name:     "Updated Job",
		Schedule: "@hourly",
		AgentID:  "test-agent",
		Enabled:  true,
	}

	err = scheduler.UpdateJob(updatedJob)
	if err != nil {
		t.Fatalf("Failed to update job: %v", err)
	}

	// Verify update
	retrieved, _ := scheduler.GetJob("update-test")
	if retrieved.Name != "Updated Job" {
		t.Errorf("Expected name 'Updated Job', got %v", retrieved.Name)
	}
}

// TestHistoryRange tests getting history in a time range.
func TestHistoryRange(t *testing.T) {
	store, err := NewHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create history store: %v", err)
	}
	defer store.Close()

	now := time.Now()

	// Add runs at different times
	for i := 0; i < 5; i++ {
		history := &RunHistory{
			JobID:        "range-test",
			ScheduledFor: now.Add(-time.Duration(i) * time.Hour),
			StartedAt:    now.Add(-time.Duration(i) * time.Hour),
			Status:       StatusSuccess,
		}
		store.AddRun(history)
	}

	// Get history for last 2.5 hours (excludes the 3h ago entry)
	start := now.Add(-150 * time.Minute)
	end := now

	histories, err := store.GetHistoryInRange(start, end)
	if err != nil {
		t.Fatalf("Failed to get history in range: %v", err)
	}

	// Should get 3 entries (0, 1, 2 hours ago - within 2.5 hours)
	if len(histories) != 3 {
		t.Errorf("Expected 3 history entries, got %d", len(histories))
	}
}

// TestDeleteOldHistory tests deleting old history.
func TestDeleteOldHistory(t *testing.T) {
	store, err := NewHistoryStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create history store: %v", err)
	}
	defer store.Close()

	// Add old run
	oldHistory := &RunHistory{
		JobID:        "delete-test",
		ScheduledFor: time.Now().Add(-48 * time.Hour),
		StartedAt:    time.Now().Add(-48 * time.Hour),
		Status:       StatusSuccess,
	}
	store.AddRun(oldHistory)

	// Add new run
	newHistory := &RunHistory{
		JobID:        "delete-test",
		ScheduledFor: time.Now(),
		StartedAt:    time.Now(),
		Status:       StatusSuccess,
	}
	store.AddRun(newHistory)

	// Delete old history
	deleted, err := store.DeleteOldHistory(24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to delete old history: %v", err)
	}

	if deleted != 1 {
		t.Errorf("Expected 1 deleted record, got %d", deleted)
	}

	// Verify only new history remains
	histories, _ := store.GetHistory("delete-test", 10)
	if len(histories) != 1 {
		t.Errorf("Expected 1 remaining history entry, got %d", len(histories))
	}
}

// TestExpandAlias tests alias expansion.
func TestExpandAlias(t *testing.T) {
	tests := []struct {
		alias    string
		expected string
		ok       bool
	}{
		{"@daily", "0 0 * * *", true},
		{"@hourly", "0 * * * *", true},
		{"@weekly", "0 0 * * 0", true},
		{"@monthly", "0 0 1 * *", true},
		{"@invalid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			expanded, ok := ExpandAlias(tt.alias)
			if ok != tt.ok {
				t.Errorf("ExpandAlias() ok = %v, want %v", ok, tt.ok)
			}
			if ok && expanded != tt.expected {
				t.Errorf("ExpandAlias() = %v, want %v", expanded, tt.expected)
			}
		})
	}
}

// TestFileDatabase tests using a file-based database.
func TestFileDatabase(t *testing.T) {
	dbPath := "./test-cron.db"
	defer os.Remove(dbPath)

	// Create scheduler with file database
	config := Config{
		DBPath:          dbPath,
		DefaultTimezone: "UTC",
	}

	scheduler, err := NewScheduler(config)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Add job
	job := testJob("file-db-test")
	scheduler.AddJob(job)

	// Close and recreate
	scheduler.Close()

	// Verify persistence
	scheduler2, err := NewScheduler(config)
	if err != nil {
		t.Fatalf("Failed to create second scheduler: %v", err)
	}
	defer scheduler2.Close()

	// Job should still exist
	jobs := scheduler2.GetJobs()
	if len(jobs) != 1 {
		t.Errorf("Expected 1 persisted job, got %d", len(jobs))
	}
}

// TestJobGetLocation tests job timezone handling.
func TestJobGetLocation(t *testing.T) {
	job := Job{
		ID:       "tz-test",
		Schedule: "@daily",
		AgentID:  "test-agent",
		Timezone: "America/New_York",
	}

	loc := job.GetLocation()
	if loc == nil {
		t.Fatal("Expected non-nil location")
	}

	// Verify it's New York timezone
	nyLoc, _ := time.LoadLocation("America/New_York")
	if loc.String() != nyLoc.String() {
		t.Errorf("Location = %v, want %v", loc, nyLoc)
	}

	// Test empty timezone defaults to UTC
	utcJob := Job{
		ID:       "utc-test",
		Schedule: "@daily",
		AgentID:  "test-agent",
	}

	utcLoc := utcJob.GetLocation()
	if utcLoc != time.UTC {
		t.Errorf("Empty timezone should default to UTC, got %v", utcLoc)
	}
}

// BenchmarkParser benchmarks cron parsing.
func BenchmarkParser(b *testing.B) {
	parser := NewParser()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parser.Parse("0 9 * * *")
	}
}

// BenchmarkSchedulerAddJob benchmarks adding jobs.
func BenchmarkSchedulerAddJob(b *testing.B) {
	config := testConfig()
	scheduler, _ := NewScheduler(config)
	defer scheduler.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job := Job{
			ID:       string(rune(i)),
			Schedule: "@daily",
			AgentID:  "test-agent",
		}
		scheduler.AddJob(job)
	}
}
