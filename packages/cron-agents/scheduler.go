package cron

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// ErrSchedulerStopped is returned when operations are attempted on a stopped scheduler.
var ErrSchedulerStopped = errors.New("scheduler is stopped")

// ErrJobNotFound is returned when a job cannot be found.
var ErrJobNotFound = errors.New("job not found")

// ErrJobAlreadyExists is returned when adding a job with an existing ID.
var ErrJobAlreadyExists = errors.New("job already exists")

// JobExecutor is the function signature for executing jobs.
type JobExecutor func(ctx context.Context, job *Job) (interface{}, error)

// Scheduler manages cron jobs with persistent history and thread-safe operations.
type Scheduler struct {
	config    Config
	store     *HistoryStore
	cron      *cron.Cron
	jobs      map[string]*Job
	running   map[string]bool
	executors map[string]JobExecutor

	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	wg      sync.WaitGroup
}

// SchedulerOption is a functional option for configuring the scheduler.
type SchedulerOption func(*Scheduler) error

// WithExecutor sets a default executor for all jobs.
func WithExecutor(executor JobExecutor) SchedulerOption {
	return func(s *Scheduler) error {
		s.executors["default"] = executor
		return nil
	}
}

// WithLogger sets the logger for the scheduler.
func WithLogger(logger Logger) SchedulerOption {
	return func(s *Scheduler) error {
		s.config.Logger = logger
		return nil
	}
}

// NewScheduler creates a new scheduler with the given configuration.
func NewScheduler(config Config, opts ...SchedulerOption) (*Scheduler, error) {
	// Set defaults
	if config.DBPath == "" {
		config.DBPath = ":memory:"
	}
	if config.DefaultTimezone == "" {
		config.DefaultTimezone = "UTC"
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.MissedRunBehavior == "" {
		config.MissedRunBehavior = "skip"
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize history store
	store, err := NewHistoryStore(config.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize history store: %w", err)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	// Create cron instance with timezone support and full parser
	loc := config.GetTimezone()
	cronInstance := cron.New(cron.WithLocation(loc))

	s := &Scheduler{
		config:    config,
		store:     store,
		cron:      cronInstance,
		jobs:      make(map[string]*Job),
		running:   make(map[string]bool),
		executors: make(map[string]JobExecutor),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(s); err != nil {
			store.Close()
			return nil, fmt.Errorf("option error: %w", err)
		}
	}

	// Load existing jobs from database
	if err := s.loadJobsFromDB(); err != nil {
		s.log("warn", "failed to load jobs from database", Field{Key: "error", Value: err.Error()})
	}

	return s, nil
}

// loadJobsFromDB loads persisted jobs from the database.
func (s *Scheduler) loadJobsFromDB() error {
	jobs, err := s.store.LoadAllJobs()
	if err != nil {
		return err
	}

	for i := range jobs {
		job := &jobs[i]
		s.jobs[job.ID] = job

		// Handle missed runs
		if job.Enabled && job.NextRun.After(time.Time{}) && job.NextRun.Before(time.Now()) {
			s.handleMissedRun(job)
		}
	}

	return nil
}

// handleMissedRun handles a missed run based on the configured behavior.
func (s *Scheduler) handleMissedRun(job *Job) {
	switch s.config.MissedRunBehavior {
	case "run":
		s.log("info", "running missed job", Field{Key: "job_id", Value: job.ID})
		go s.executeJob(job)
	case "skip", "ignore":
		// Recalculate next run
		nextRun, err := job.CalculateNextRun(time.Now())
		if err != nil {
			s.log("error", "failed to calculate next run", Field{Key: "job_id", Value: job.ID}, Field{Key: "error", Value: err.Error()})
			return
		}
		job.NextRun = nextRun
		s.store.SaveJob(job)
	}
}

// AddJob adds a new job to the scheduler.
func (s *Scheduler) AddJob(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return errors.New("cannot add job while scheduler is running")
	}

	// Validate job
	if err := job.Validate(); err != nil {
		return fmt.Errorf("invalid job: %w", err)
	}

	// Check for duplicate
	if _, exists := s.jobs[job.ID]; exists {
		return ErrJobAlreadyExists
	}

	// Set timestamps
	now := time.Now()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	job.UpdatedAt = now

	// Calculate next run
	nextRun, err := job.CalculateNextRun(now)
	if err != nil {
		return fmt.Errorf("failed to calculate next run: %w", err)
	}
	job.NextRun = nextRun

	// Save to database
	if err := s.store.SaveJob(&job); err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	// Add to memory
	s.jobs[job.ID] = &job

	s.log("info", "job added",
		Field{Key: "job_id", Value: job.ID},
		Field{Key: "schedule", Value: job.Schedule},
		Field{Key: "next_run", Value: job.NextRun},
	)

	return nil
}

// RemoveJob removes a job from the scheduler.
func (s *Scheduler) RemoveJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return ErrJobNotFound
	}

	// Wait for any running execution to complete
	if s.running[id] {
		s.mu.Unlock()
		// Wait with timeout
		for i := 0; i < 100; i++ {
			time.Sleep(100 * time.Millisecond)
			s.mu.Lock()
			if !s.running[id] {
				break
			}
			s.mu.Unlock()
		}
	}

	// Remove cron entry if scheduler is running
	if job.cronEntryID != 0 && s.started {
		s.cron.Remove(job.cronEntryID)
	}

	// Remove from memory
	delete(s.jobs, id)
	delete(s.running, id)

	// Remove from database
	if err := s.store.DeleteJob(id); err != nil {
		s.log("error", "failed to delete job from database",
			Field{Key: "job_id", Value: id},
			Field{Key: "error", Value: err.Error()},
		)
	}

	s.log("info", "job removed", Field{Key: "job_id", Value: id})

	return nil
}

// PauseJob pauses a job temporarily.
func (s *Scheduler) PauseJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return ErrJobNotFound
	}

	if !job.Enabled {
		return nil // Already paused/disabled
	}

	job.Enabled = false
	job.UpdatedAt = time.Now()

	// Remove from cron if running
	if job.cronEntryID != 0 && s.started {
		s.cron.Remove(job.cronEntryID)
		job.cronEntryID = 0
	}

	// Save to database
	if err := s.store.SaveJob(job); err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	s.log("info", "job paused", Field{Key: "job_id", Value: id})

	return nil
}

// ResumeJob resumes a paused job.
func (s *Scheduler) ResumeJob(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return ErrJobNotFound
	}

	if job.Enabled {
		return nil // Already enabled
	}

	job.Enabled = true
	job.UpdatedAt = time.Now()

	// Recalculate next run
	nextRun, err := job.CalculateNextRun(time.Now())
	if err != nil {
		return fmt.Errorf("failed to calculate next run: %w", err)
	}
	job.NextRun = nextRun

	// Add to cron if scheduler is running
	if s.started {
		if err := s.scheduleJob(job); err != nil {
			return fmt.Errorf("failed to schedule job: %w", err)
		}
	}

	// Save to database
	if err := s.store.SaveJob(job); err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	s.log("info", "job resumed",
		Field{Key: "job_id", Value: id},
		Field{Key: "next_run", Value: job.NextRun},
	)

	return nil
}

// Start starts the scheduler in the background.
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return errors.New("scheduler already started")
	}

	// Schedule all enabled jobs
	for _, job := range s.jobs {
		if !job.Enabled {
			continue
		}
		if err := s.scheduleJob(job); err != nil {
			s.log("error", "failed to schedule job",
				Field{Key: "job_id", Value: job.ID},
				Field{Key: "error", Value: err.Error()},
			)
		}
	}

	s.cron.Start()
	s.started = true

	s.log("info", "scheduler started")

	return nil
}

// scheduleJob adds a job to the cron scheduler.
func (s *Scheduler) scheduleJob(job *Job) error {
	// Validate and parse schedule
	parser := NewParser()
	parsed, err := parser.Parse(job.Schedule)
	if err != nil {
		return err
	}
	job.parsedSchedule = parsed

	// Get job location
	loc := job.GetLocation()

	// Create a cron.Job wrapper for our job
	cronJob := &jobWrapper{job: job, executor: s}

	// Add to cron using Schedule (supports @every syntax)
	entryID := s.cron.Schedule(parsed.Schedule, cronJob)

	job.cronEntryID = entryID
	job.location = loc

	// Update next run time
	job.NextRun = parsed.Next(time.Now().In(loc))

	s.log("debug", "job scheduled",
		Field{Key: "job_id", Value: job.ID},
		Field{Key: "entry_id", Value: entryID},
		Field{Key: "next_run", Value: job.NextRun},
	)

	return nil
}

// jobWrapper implements cron.Job interface
type jobWrapper struct {
	job      *Job
	executor *Scheduler
}

func (j *jobWrapper) Run() {
	j.executor.executeJob(j.job)
}

// executeJob executes a job and records the result.
func (s *Scheduler) executeJob(job *Job) {
	s.mu.Lock()

	// Check if already running (skip concurrent execution)
	if s.running[job.ID] {
		s.mu.Unlock()
		s.log("warn", "job already running, skipping", Field{Key: "job_id", Value: job.ID})
		return
	}

	// Check if job is still enabled
	if !job.Enabled {
		s.mu.Unlock()
		return
	}

	s.running[job.ID] = true
	s.mu.Unlock()

	// Create run history entry
	history := &RunHistory{
		JobID:        job.ID,
		ScheduledFor: job.NextRun,
		StartedAt:    time.Now(),
		Status:       StatusRunning,
		RetryCount:   0,
	}

	// Record start
	s.store.AddRun(history)

	// Execute with retry logic
	var result interface{}
	var execErr error
	maxAttempts := 1
	if s.config.RetryOnFailure {
		maxAttempts = s.config.MaxRetries + 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			history.RetryCount = attempt
			s.log("info", "retrying job",
				Field{Key: "job_id", Value: job.ID},
				Field{Key: "attempt", Value: attempt + 1},
			)
		}

		// Get executor
		executor := s.getExecutor(job)
		if executor == nil {
			execErr = errors.New("no executor configured for job")
			break
		}

		// Execute with timeout context
		ctx, cancel := context.WithTimeout(s.ctx, 30*time.Minute)
		result, execErr = executor(ctx, job)
		cancel()

		if execErr == nil {
			break
		}

		s.log("warn", "job execution failed",
			Field{Key: "job_id", Value: job.ID},
			Field{Key: "attempt", Value: attempt + 1},
			Field{Key: "error", Value: execErr.Error()},
		)

		// Wait before retry (exponential backoff)
		if attempt < maxAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	// Update history
	history.CompletedAt = time.Now()
	history.Duration = history.CompletedAt.Sub(history.StartedAt).Milliseconds()
	history.Result = result

	if execErr != nil {
		history.Status = StatusFailed
		history.Error = execErr.Error()
	} else {
		history.Status = StatusSuccess
	}

	s.store.UpdateRun(history)

	// Update job
	s.mu.Lock()
	job.LastRun = history.StartedAt
	if job.parsedSchedule != nil {
		job.NextRun = job.parsedSchedule.Next(time.Now().In(job.GetLocation()))
	}
	job.UpdatedAt = time.Now()
	s.store.SaveJob(job)
	delete(s.running, job.ID)
	s.mu.Unlock()

	// Log result
	if history.Status == StatusSuccess {
		s.log("info", "job completed successfully",
			Field{Key: "job_id", Value: job.ID},
			Field{Key: "duration_ms", Value: history.Duration},
		)
	} else {
		s.log("error", "job failed",
			Field{Key: "job_id", Value: job.ID},
			Field{Key: "duration_ms", Value: history.Duration},
			Field{Key: "error", Value: history.Error},
			Field{Key: "retries", Value: history.RetryCount},
		)
	}
}

// getExecutor returns the executor for a job.
func (s *Scheduler) getExecutor(job *Job) JobExecutor {
	// Check for job-specific executor
	if exec, ok := s.executors[job.ID]; ok {
		return exec
	}
	// Fall back to default
	return s.executors["default"]
}

// SetExecutor sets an executor for a specific job or as default.
func (s *Scheduler) SetExecutor(jobID string, executor JobExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executors[jobID] = executor
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	// Stop cron
	ctx := s.cron.Stop()
	<-ctx.Done()

	s.started = false
	s.cancel()

	s.log("info", "scheduler stopped")

	return nil
}

// Close stops the scheduler and closes the database.
func (s *Scheduler) Close() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.store.Close()
}

// GetJob retrieves a job by ID.
func (s *Scheduler) GetJob(id string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, ErrJobNotFound
	}

	return job, nil
}

// GetJobs retrieves all jobs.
func (s *Scheduler) GetJobs() []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, *job)
	}

	return jobs
}

// GetEnabledJobs retrieves all enabled jobs.
func (s *Scheduler) GetEnabledJobs() []Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []Job
	for _, job := range s.jobs {
		if job.Enabled {
			jobs = append(jobs, *job)
		}
	}

	return jobs
}

// GetNextRun returns the next scheduled run time for a job.
func (s *Scheduler) GetNextRun(id string) (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return time.Time{}, ErrJobNotFound
	}

	return job.NextRun, nil
}

// GetHistory retrieves the run history for a job.
func (s *Scheduler) GetHistory(id string, limit int) ([]RunHistory, error) {
	return s.store.GetHistory(id, limit)
}

// GetAllHistory retrieves all run history.
func (s *Scheduler) GetAllHistory(limit int) ([]RunHistory, error) {
	return s.store.GetAllHistory(limit)
}

// GetJobStatus returns the current status of a job.
func (s *Scheduler) GetJobStatus(id string) (*JobStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, ErrJobNotFound
	}

	return &JobStatus{
		ID:        job.ID,
		Name:      job.Name,
		Enabled:   job.Enabled,
		LastRun:   job.LastRun,
		NextRun:   job.NextRun,
		IsRunning: s.running[id],
	}, nil
}

// RunJob triggers an immediate execution of a job.
func (s *Scheduler) RunJob(id string) error {
	s.mu.RLock()
	job, exists := s.jobs[id]
	s.mu.RUnlock()

	if !exists {
		return ErrJobNotFound
	}

	go s.executeJob(job)

	return nil
}

// UpdateJob updates an existing job.
func (s *Scheduler) UpdateJob(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.jobs[job.ID]
	if !exists {
		return ErrJobNotFound
	}

	// Validate new schedule
	parser := NewParser()
	parsed, err := parser.Parse(job.Schedule)
	if err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}
	job.parsedSchedule = parsed

	// Preserve creation time
	job.CreatedAt = existing.CreatedAt
	job.UpdatedAt = time.Now()

	// Recalculate next run
	job.NextRun = parsed.Next(time.Now().In(job.GetLocation()))

	// If scheduler is running and job was scheduled, reschedule
	if s.started && existing.cronEntryID != 0 {
		s.cron.Remove(existing.cronEntryID)
		if job.Enabled {
			if err := s.scheduleJob(&job); err != nil {
				return fmt.Errorf("failed to reschedule job: %w", err)
			}
		}
	}

	// Save to database
	if err := s.store.SaveJob(&job); err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	// Update memory
	s.jobs[job.ID] = &job

	s.log("info", "job updated", Field{Key: "job_id", Value: job.ID})

	return nil
}

// IsRunning returns whether the scheduler is running.
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

// GetStats returns statistics about the scheduler.
func (s *Scheduler) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	s.mu.RLock()
	stats["total_jobs"] = len(s.jobs)
	enabled := 0
	running := 0
	for _, job := range s.jobs {
		if job.Enabled {
			enabled++
		}
		if s.running[job.ID] {
			running++
		}
	}
	s.mu.RUnlock()

	stats["enabled_jobs"] = enabled
	stats["running_jobs"] = running
	stats["is_running"] = s.started

	// Get history stats
	historyStats, err := s.store.GetStats()
	if err != nil {
		return nil, err
	}
	stats["history"] = historyStats

	return stats, nil
}

// log is a helper for logging.
func (s *Scheduler) log(level, msg string, fields ...Field) {
	if s.config.Logger == nil {
		return
	}

	switch level {
	case "debug":
		s.config.Logger.Debug(msg, fields...)
	case "info":
		s.config.Logger.Info(msg, fields...)
	case "warn":
		s.config.Logger.Warn(msg, fields...)
	case "error":
		s.config.Logger.Error(msg, fields...)
	}
}
