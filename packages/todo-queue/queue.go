package todoqueue

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Queue manages a priority queue of TODO items with SQLite persistence.
type Queue struct {
	store     *Store
	config    Config
	mu        sync.RWMutex
	updatedAt time.Time
}

// New creates a new Queue with the given configuration.
func New(config Config) (*Queue, error) {
	if config.DBPath == "" {
		config.DBPath = DefaultConfig().DBPath
	}

	if config.RetryDelay == 0 {
		config.RetryDelay = DefaultConfig().RetryDelay
	}

	store, err := NewStore(config.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	// Run migrations if AutoMigrate is true (default)
	if config.AutoMigrate {
		if err := store.RunMigrations(); err != nil {
			store.Close()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return &Queue{
		store:  store,
		config: config,
	}, nil
}

// Close closes the queue and its underlying database connection.
func (q *Queue) Close() error {
	return q.store.Close()
}

// Add creates a new TODO item and adds it to the queue.
func (q *Queue) Add(description string, opts ...AddOption) (*Item, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := &Item{
		ID:          generateID(),
		Description: description,
		Status:      StatusPending,
		Priority:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Apply options
	for _, opt := range opts {
		opt(item)
	}

	// Validate status
	if !item.Status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", item.Status)
	}

	if err := q.store.Save(item); err != nil {
		return nil, fmt.Errorf("failed to add item: %w", err)
	}

	return item.clone(), nil
}

// AddOption is a functional option for adding items.
type AddOption func(*Item)

// WithPriority sets the priority of the item.
func WithPriority(priority int) AddOption {
	return func(item *Item) {
		item.Priority = priority
	}
}

// WithID sets a custom ID for the item.
func WithID(id string) AddOption {
	return func(item *Item) {
		item.ID = id
	}
}

// WithStatus sets the initial status of the item.
func WithStatus(status Status) AddOption {
	return func(item *Item) {
		item.Status = status
	}
}

// WithDependencies sets the dependencies of the item.
func WithDependencies(depIDs ...string) AddOption {
	return func(item *Item) {
		item.Dependencies = depIDs
	}
}

// WithSkills sets the required skills for the item.
func WithSkills(skills ...string) AddOption {
	return func(item *Item) {
		item.SkillsRequired = skills
	}
}

// WithMetadata sets custom metadata for the item.
func WithMetadata(metadata map[string]interface{}) AddOption {
	return func(item *Item) {
		item.Metadata = metadata
	}
}

// Next returns the highest priority pending item that is ready to be processed.
// Returns nil if no items are available.
func (q *Queue) Next() (*Item, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	items, err := q.store.Query(Filter{Status: StatusPending, Limit: 100})
	if err != nil {
		return nil, fmt.Errorf("failed to get next item: %w", err)
	}

	// Find the first item whose dependencies are met
	for _, item := range items {
		ready, err := q.checkDependenciesReady(item.ID)
		if err != nil {
			continue // Skip items with dependency check errors
		}
		if ready {
			return item.clone(), nil
		}
	}

	return nil, nil
}

// Start marks an item as in-progress.
func (q *Queue) Start(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, err := q.store.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item not found: %s", id)
	}

	if !item.Status.CanTransitionTo(StatusInProgress) {
		return fmt.Errorf("cannot start item in status %s", item.Status)
	}

	// Check dependencies before starting
	ready, err := q.checkDependenciesReady(id)
	if err != nil {
		return fmt.Errorf("failed to check dependencies: %w", err)
	}
	if !ready {
		return fmt.Errorf("dependencies not met for item %s", id)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":     StatusInProgress,
		"started_at": now,
	}

	if err := q.store.Update(id, updates); err != nil {
		return fmt.Errorf("failed to start item: %w", err)
	}

	return nil
}

// Complete marks an item as completed with the given result.
func (q *Queue) Complete(id string, result string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, err := q.store.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item not found: %s", id)
	}

	if !item.Status.CanTransitionTo(StatusCompleted) {
		return fmt.Errorf("cannot complete item in status %s", item.Status)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       StatusCompleted,
		"result":       result,
		"completed_at": now,
	}

	if err := q.store.Update(id, updates); err != nil {
		return fmt.Errorf("failed to complete item: %w", err)
	}

	// Update any blocked items that depend on this one
	go q.updateBlockedDependents(id)

	return nil
}

// Fail marks an item as failed with the given error message.
// If auto-retry is enabled and retries are available, the item is marked for retry.
func (q *Queue) Fail(id string, errMsg string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, err := q.store.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item not found: %s", id)
	}

	if !item.Status.CanTransitionTo(StatusFailed) {
		return fmt.Errorf("cannot fail item in status %s", item.Status)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"error":        errMsg,
		"completed_at": now,
	}

	// Check if we should retry
	if q.config.MaxRetries > 0 && item.RetryCount < q.config.MaxRetries {
		updates["status"] = StatusPending
		updates["retry_count"] = item.RetryCount + 1
		updates["completed_at"] = nil
	} else {
		updates["status"] = StatusFailed
	}

	if err := q.store.Update(id, updates); err != nil {
		return fmt.Errorf("failed to fail item: %w", err)
	}

	return nil
}

// Block marks an item as blocked.
func (q *Queue) Block(id string, reason string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, err := q.store.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item not found: %s", id)
	}

	if !item.Status.CanTransitionTo(StatusBlocked) {
		return fmt.Errorf("cannot block item in status %s", item.Status)
	}

	updates := map[string]interface{}{
		"status": StatusBlocked,
		"error":  reason,
	}

	if err := q.store.Update(id, updates); err != nil {
		return fmt.Errorf("failed to block item: %w", err)
	}

	return nil
}

// Unblock marks a blocked item as pending.
func (q *Queue) Unblock(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, err := q.store.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item not found: %s", id)
	}

	if item.Status != StatusBlocked {
		return fmt.Errorf("item is not blocked")
	}

	updates := map[string]interface{}{
		"status": StatusPending,
		"error":  "",
	}

	if err := q.store.Update(id, updates); err != nil {
		return fmt.Errorf("failed to unblock item: %w", err)
	}

	return nil
}

// Retry manually retries a failed item.
func (q *Queue) Retry(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, err := q.store.Get(id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item not found: %s", id)
	}

	if item.Status != StatusFailed {
		return fmt.Errorf("can only retry failed items")
	}

	updates := map[string]interface{}{
		"status":       StatusPending,
		"retry_count":  item.RetryCount + 1,
		"error":        "",
		"completed_at": nil,
		"started_at":   nil,
	}

	if err := q.store.Update(id, updates); err != nil {
		return fmt.Errorf("failed to retry item: %w", err)
	}

	return nil
}

// List returns items matching the filter.
func (q *Queue) List(opts ListOptions) ([]*Item, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	items, err := q.store.Query(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	// Return copies
	result := make([]*Item, len(items))
	for i, item := range items {
		result[i] = item.clone()
	}

	return result, nil
}

// Get retrieves a specific item by ID.
func (q *Queue) Get(id string) (*Item, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	item, err := q.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if item == nil {
		return nil, nil
	}

	return item.clone(), nil
}

// Delete removes an item from the queue.
func (q *Queue) Delete(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.store.Delete(id)
}

// CheckDependencies checks if all dependencies for an item are met.
func (q *Queue) CheckDependencies(id string) (bool, []string, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	item, err := q.store.Get(id)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return false, nil, fmt.Errorf("item not found: %s", id)
	}

	if len(item.Dependencies) == 0 {
		return true, nil, nil
	}

	unmet := make([]string, 0)
	for _, depID := range item.Dependencies {
		dep, err := q.store.Get(depID)
		if err != nil {
			return false, nil, err
		}
		if dep == nil || dep.Status != StatusCompleted {
			unmet = append(unmet, depID)
		}
	}

	return len(unmet) == 0, unmet, nil
}

// checkDependenciesReady is an internal method to check if dependencies are met.
func (q *Queue) checkDependenciesReady(id string) (bool, error) {
	ready, _, err := q.CheckDependencies(id)
	return ready, err
}

// updateBlockedDependents checks and updates items that depend on the completed item.
func (q *Queue) updateBlockedDependents(completedID string) {
	// Find all items with this dependency
	items, err := q.store.Query(Filter{Status: StatusBlocked})
	if err != nil {
		return
	}

	for _, item := range items {
		for _, depID := range item.Dependencies {
			if depID == completedID {
				// Check if all dependencies are now met
				ready, _ := q.checkDependenciesReady(item.ID)
				if ready {
					q.store.Update(item.ID, map[string]interface{}{
						"status": StatusPending,
						"error":  "",
					})
				}
			}
		}
	}
}

// Stats returns statistics about the queue.
func (q *Queue) Stats() (*QueueStats, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	stats := &QueueStats{}

	var err error
	stats.Total, err = q.store.Count(Filter{})
	if err != nil {
		return nil, err
	}

	stats.Pending, err = q.store.Count(Filter{Status: StatusPending})
	if err != nil {
		return nil, err
	}

	stats.InProgress, err = q.store.Count(Filter{Status: StatusInProgress})
	if err != nil {
		return nil, err
	}

	stats.Completed, err = q.store.Count(Filter{Status: StatusCompleted})
	if err != nil {
		return nil, err
	}

	stats.Failed, err = q.store.Count(Filter{Status: StatusFailed})
	if err != nil {
		return nil, err
	}

	stats.Blocked, err = q.store.Count(Filter{Status: StatusBlocked})
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// QueueStats holds statistics about the queue.
type QueueStats struct {
	Total      int
	Pending    int
	InProgress int
	Completed  int
	Failed     int
	Blocked    int
}

// Update modifies an existing item's fields.
func (q *Queue) Update(id string, updates map[string]interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Validate status if being updated
	if status, ok := updates["status"].(Status); ok {
		if !status.IsValid() {
			return fmt.Errorf("invalid status: %s", status)
		}
	}

	return q.store.Update(id, updates)
}

// generateID generates a unique ID for a new item.
func generateID() string {
	return uuid.New().String()
}
