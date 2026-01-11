// Package scheduler provides scheduling capabilities for data quality checks.
// Supports cron-based scheduling with alerting integration.
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vinod901/opendq-go/internal/alerting"
	"github.com/vinod901/opendq-go/internal/check"
)

// Schedule represents a schedule for running checks
type Schedule struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	CronExpression  string                 `json:"cron_expression"`
	Timezone        string                 `json:"timezone"`
	CheckIDs        []string               `json:"check_ids"`
	DatasourceID    string                 `json:"datasource_id,omitempty"` // Run all checks for datasource
	AlertChannelIDs []string               `json:"alert_channel_ids"`
	Active          bool                   `json:"active"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	LastRunAt       *time.Time             `json:"last_run_at,omitempty"`
	NextRunAt       *time.Time             `json:"next_run_at,omitempty"`
}

// ScheduleExecution represents a single execution of a schedule
type ScheduleExecution struct {
	ID          string                 `json:"id"`
	ScheduleID  string                 `json:"schedule_id"`
	Status      ExecutionStatus        `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Results     []*check.CheckResult   `json:"results"`
	Summary     ExecutionSummary       `json:"summary"`
	Error       string                 `json:"error,omitempty"`
}

// ExecutionStatus represents the status of a schedule execution
type ExecutionStatus string

const (
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusPartial   ExecutionStatus = "partial"
)

// ExecutionSummary contains a summary of the execution results
type ExecutionSummary struct {
	TotalChecks   int `json:"total_checks"`
	PassedChecks  int `json:"passed_checks"`
	FailedChecks  int `json:"failed_checks"`
	WarningChecks int `json:"warning_checks"`
	ErrorChecks   int `json:"error_checks"`
	SkippedChecks int `json:"skipped_checks"`
}

// Common cron expressions
const (
	CronEveryMinute    = "* * * * *"
	CronEvery5Minutes  = "*/5 * * * *"
	CronEvery15Minutes = "*/15 * * * *"
	CronEvery30Minutes = "*/30 * * * *"
	CronHourly         = "0 * * * *"
	CronDaily          = "0 0 * * *"
	CronWeekly         = "0 0 * * 0"
	CronMonthly        = "0 0 1 * *"
)

// Manager handles schedule operations
type Manager struct {
	schedules     map[string]*Schedule
	executions    map[string][]*ScheduleExecution
	checkManager  *check.Manager
	alertManager  *alerting.Manager
	running       map[string]context.CancelFunc
	mu            sync.RWMutex
	stopChan      chan struct{}
}

// NewManager creates a new scheduler manager
func NewManager(checkManager *check.Manager, alertManager *alerting.Manager) *Manager {
	return &Manager{
		schedules:    make(map[string]*Schedule),
		executions:   make(map[string][]*ScheduleExecution),
		checkManager: checkManager,
		alertManager: alertManager,
		running:      make(map[string]context.CancelFunc),
		stopChan:     make(chan struct{}),
	}
}

// CreateSchedule creates a new schedule
func (m *Manager) CreateSchedule(ctx context.Context, schedule *Schedule) error {
	if schedule.ID == "" {
		schedule.ID = uuid.New().String()
	}
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()
	schedule.Active = true

	// Validate cron expression
	if err := validateCronExpression(schedule.CronExpression); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Calculate next run time
	nextRun, err := getNextRunTime(schedule.CronExpression, schedule.Timezone)
	if err == nil {
		schedule.NextRunAt = &nextRun
	}

	m.mu.Lock()
	m.schedules[schedule.ID] = schedule
	m.mu.Unlock()

	// Start the schedule if active
	if schedule.Active {
		m.startSchedule(schedule)
	}

	return nil
}

// GetSchedule retrieves a schedule by ID
func (m *Manager) GetSchedule(ctx context.Context, id string) (*Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schedule, exists := m.schedules[id]
	if !exists {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}
	return schedule, nil
}

// UpdateSchedule updates a schedule
func (m *Manager) UpdateSchedule(ctx context.Context, id string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	schedule, exists := m.schedules[id]
	if !exists {
		return fmt.Errorf("schedule not found: %s", id)
	}

	wasActive := schedule.Active

	if name, ok := updates["name"].(string); ok {
		schedule.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		schedule.Description = description
	}
	if cronExpr, ok := updates["cron_expression"].(string); ok {
		if err := validateCronExpression(cronExpr); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		schedule.CronExpression = cronExpr
		nextRun, err := getNextRunTime(cronExpr, schedule.Timezone)
		if err == nil {
			schedule.NextRunAt = &nextRun
		}
	}
	if active, ok := updates["active"].(bool); ok {
		schedule.Active = active
	}
	if checkIDs, ok := updates["check_ids"].([]string); ok {
		schedule.CheckIDs = checkIDs
	}
	if alertChannelIDs, ok := updates["alert_channel_ids"].([]string); ok {
		schedule.AlertChannelIDs = alertChannelIDs
	}

	schedule.UpdatedAt = time.Now()

	// Handle activation/deactivation
	if wasActive && !schedule.Active {
		m.stopScheduleInternal(id)
	} else if !wasActive && schedule.Active {
		m.startSchedule(schedule)
	}

	return nil
}

// DeleteSchedule deletes a schedule
func (m *Manager) DeleteSchedule(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.schedules[id]; !exists {
		return fmt.Errorf("schedule not found: %s", id)
	}

	m.stopScheduleInternal(id)
	delete(m.schedules, id)
	delete(m.executions, id)

	return nil
}

// ListSchedules lists schedules with optional filters
func (m *Manager) ListSchedules(ctx context.Context, tenantID string) ([]*Schedule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Schedule
	for _, schedule := range m.schedules {
		if tenantID == "" || schedule.TenantID == tenantID {
			result = append(result, schedule)
		}
	}
	return result, nil
}

// RunScheduleNow triggers immediate execution of a schedule
func (m *Manager) RunScheduleNow(ctx context.Context, id string) (*ScheduleExecution, error) {
	m.mu.RLock()
	schedule, exists := m.schedules[id]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("schedule not found: %s", id)
	}

	return m.executeSchedule(ctx, schedule)
}

// GetScheduleExecutions returns execution history for a schedule
func (m *Manager) GetScheduleExecutions(ctx context.Context, scheduleID string, limit int) ([]*ScheduleExecution, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	executions, exists := m.executions[scheduleID]
	if !exists {
		return []*ScheduleExecution{}, nil
	}

	if limit > 0 && len(executions) > limit {
		return executions[len(executions)-limit:], nil
	}
	return executions, nil
}

// Start starts the scheduler
func (m *Manager) Start(ctx context.Context) error {
	m.mu.RLock()
	schedules := make([]*Schedule, 0, len(m.schedules))
	for _, s := range m.schedules {
		if s.Active {
			schedules = append(schedules, s)
		}
	}
	m.mu.RUnlock()

	for _, schedule := range schedules {
		m.startSchedule(schedule)
	}

	return nil
}

// Stop stops the scheduler
func (m *Manager) Stop() {
	close(m.stopChan)

	m.mu.Lock()
	for id, cancel := range m.running {
		cancel()
		delete(m.running, id)
	}
	m.mu.Unlock()
}

// startSchedule starts a schedule's timer
func (m *Manager) startSchedule(schedule *Schedule) {
	ctx, cancel := context.WithCancel(context.Background())

	m.mu.Lock()
	m.running[schedule.ID] = cancel
	m.mu.Unlock()

	go m.scheduleLoop(ctx, schedule)
}

// stopScheduleInternal stops a schedule's timer (must hold lock)
func (m *Manager) stopScheduleInternal(id string) {
	if cancel, exists := m.running[id]; exists {
		cancel()
		delete(m.running, id)
	}
}

// scheduleLoop runs the schedule loop
func (m *Manager) scheduleLoop(ctx context.Context, schedule *Schedule) {
	for {
		nextRun, err := getNextRunTime(schedule.CronExpression, schedule.Timezone)
		if err != nil {
			return
		}

		sleepDuration := time.Until(nextRun)
		if sleepDuration < 0 {
			sleepDuration = time.Minute // Minimum sleep
		}

		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-time.After(sleepDuration):
			m.executeSchedule(ctx, schedule)
		}
	}
}

// executeSchedule executes all checks in a schedule
func (m *Manager) executeSchedule(ctx context.Context, schedule *Schedule) (*ScheduleExecution, error) {
	execution := &ScheduleExecution{
		ID:         uuid.New().String(),
		ScheduleID: schedule.ID,
		Status:     ExecutionStatusRunning,
		StartedAt:  time.Now(),
		Results:    make([]*check.CheckResult, 0),
	}

	// Get checks to run
	var checkIDs []string
	if schedule.DatasourceID != "" {
		// Run all checks for the datasource
		checks, err := m.checkManager.ListChecks(ctx, schedule.TenantID, schedule.DatasourceID)
		if err != nil {
			execution.Status = ExecutionStatusFailed
			execution.Error = err.Error()
			return execution, err
		}
		for _, c := range checks {
			checkIDs = append(checkIDs, c.ID)
		}
	} else {
		checkIDs = schedule.CheckIDs
	}

	// Execute checks
	for _, checkID := range checkIDs {
		result, err := m.checkManager.RunCheck(ctx, checkID)
		if err != nil {
			execution.Summary.ErrorChecks++
			continue
		}
		execution.Results = append(execution.Results, result)

		// Update summary
		switch result.Status {
		case check.StatusPassed:
			execution.Summary.PassedChecks++
		case check.StatusFailed:
			execution.Summary.FailedChecks++
		case check.StatusWarning:
			execution.Summary.WarningChecks++
		case check.StatusError:
			execution.Summary.ErrorChecks++
		case check.StatusSkipped:
			execution.Summary.SkippedChecks++
		}
	}

	execution.Summary.TotalChecks = len(checkIDs)

	// Complete execution
	now := time.Now()
	execution.CompletedAt = &now
	execution.Duration = now.Sub(execution.StartedAt)

	// Determine final status
	if execution.Summary.ErrorChecks > 0 {
		execution.Status = ExecutionStatusPartial
	} else if execution.Summary.FailedChecks > 0 {
		execution.Status = ExecutionStatusCompleted
	} else {
		execution.Status = ExecutionStatusCompleted
	}

	// Store execution
	m.mu.Lock()
	m.executions[schedule.ID] = append(m.executions[schedule.ID], execution)
	schedule.LastRunAt = &now
	nextRun, _ := getNextRunTime(schedule.CronExpression, schedule.Timezone)
	schedule.NextRunAt = &nextRun
	m.mu.Unlock()

	// Send alerts if there are failures
	if execution.Summary.FailedChecks > 0 && m.alertManager != nil {
		m.sendAlerts(ctx, schedule, execution)
	}

	return execution, nil
}

// sendAlerts sends alerts for failed checks
func (m *Manager) sendAlerts(ctx context.Context, schedule *Schedule, execution *ScheduleExecution) {
	for _, channelID := range schedule.AlertChannelIDs {
		alert := &alerting.Alert{
			Title:       fmt.Sprintf("Data Quality Check Failures - %s", schedule.Name),
			Message:     fmt.Sprintf("%d of %d checks failed", execution.Summary.FailedChecks, execution.Summary.TotalChecks),
			Severity:    alerting.SeverityHigh,
			ScheduleID:  schedule.ID,
			ExecutionID: execution.ID,
			Details: map[string]interface{}{
				"summary":     execution.Summary,
				"schedule":    schedule.Name,
				"executed_at": execution.StartedAt,
			},
		}

		m.alertManager.SendAlert(ctx, channelID, alert)
	}
}

// validateCronExpression validates a cron expression
func validateCronExpression(expr string) error {
	// Basic validation - in production use robfig/cron parser
	if expr == "" {
		return fmt.Errorf("cron expression cannot be empty")
	}
	// Simple check for 5 space-separated fields
	fields := 0
	prevSpace := true
	for _, c := range expr {
		if c == ' ' {
			prevSpace = true
		} else if prevSpace {
			fields++
			prevSpace = false
		}
	}
	if fields != 5 {
		return fmt.Errorf("cron expression must have 5 fields")
	}
	return nil
}

// getNextRunTime calculates the next run time based on cron expression
func getNextRunTime(cronExpr, timezone string) (time.Time, error) {
	// In production: use robfig/cron to parse and calculate next run
	// For now, return a simple approximation
	loc := time.UTC
	if timezone != "" {
		var err error
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			loc = time.UTC
		}
	}

	now := time.Now().In(loc)
	
	// Simple parsing for common patterns
	switch cronExpr {
	case CronEveryMinute:
		return now.Add(time.Minute).Truncate(time.Minute), nil
	case CronEvery5Minutes:
		nextMinute := (now.Minute()/5 + 1) * 5
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), nextMinute%60, 0, 0, loc), nil
	case CronEvery15Minutes:
		nextMinute := (now.Minute()/15 + 1) * 15
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), nextMinute%60, 0, 0, loc), nil
	case CronEvery30Minutes:
		nextMinute := (now.Minute()/30 + 1) * 30
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), nextMinute%60, 0, 0, loc), nil
	case CronHourly:
		return now.Add(time.Hour).Truncate(time.Hour), nil
	case CronDaily:
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc), nil
	case CronWeekly:
		daysUntilSunday := (7 - int(now.Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7
		}
		return time.Date(now.Year(), now.Month(), now.Day()+daysUntilSunday, 0, 0, 0, 0, loc), nil
	case CronMonthly:
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, loc), nil
	default:
		// Default to next minute for unknown patterns
		return now.Add(time.Minute).Truncate(time.Minute), nil
	}
}
