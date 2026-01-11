package scheduler

import (
	"context"
	"testing"

	"github.com/vinod901/opendq-go/internal/alerting"
	"github.com/vinod901/opendq-go/internal/check"
	"github.com/vinod901/opendq-go/internal/datasource"
)

func TestNewManager(t *testing.T) {
	dsManager := datasource.NewManager()
	checkManager := check.NewManager(dsManager)
	alertManager := alerting.NewManager()
	
	m := NewManager(checkManager, alertManager)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.schedules == nil {
		t.Fatal("schedules map is nil")
	}
	if m.executions == nil {
		t.Fatal("executions map is nil")
	}
}

func TestCronConstants(t *testing.T) {
	cronExpressions := []string{
		CronEveryMinute,
		CronEvery5Minutes,
		CronEvery15Minutes,
		CronEvery30Minutes,
		CronHourly,
		CronDaily,
		CronWeekly,
		CronMonthly,
	}

	for _, expr := range cronExpressions {
		if expr == "" {
			t.Error("cron expression should not be empty")
		}
	}
}

func TestValidateCronExpression(t *testing.T) {
	testCases := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"valid minute", "* * * * *", false},
		{"valid 5 fields", "0 0 * * 0", false},
		{"empty", "", true},
		{"too few fields", "* * *", true},
		{"too many fields", "* * * * * *", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCronExpression(tc.expr)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateCronExpression(%s) error = %v, wantErr %v", tc.expr, err, tc.wantErr)
			}
		})
	}
}

func TestManager_CreateSchedule(t *testing.T) {
	dsManager := datasource.NewManager()
	checkManager := check.NewManager(dsManager)
	alertManager := alerting.NewManager()
	m := NewManager(checkManager, alertManager)
	ctx := context.Background()

	schedule := &Schedule{
		TenantID:       "tenant-1",
		Name:           "Daily Check",
		CronExpression: CronDaily,
		CheckIDs:       []string{"check-1"},
	}

	err := m.CreateSchedule(ctx, schedule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schedule.ID == "" {
		t.Error("schedule ID should be generated")
	}
	if !schedule.Active {
		t.Error("schedule should be active by default")
	}
}

func TestManager_CreateSchedule_InvalidCron(t *testing.T) {
	dsManager := datasource.NewManager()
	checkManager := check.NewManager(dsManager)
	alertManager := alerting.NewManager()
	m := NewManager(checkManager, alertManager)
	ctx := context.Background()

	schedule := &Schedule{
		TenantID:       "tenant-1",
		Name:           "Invalid Schedule",
		CronExpression: "invalid",
	}

	err := m.CreateSchedule(ctx, schedule)
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestManager_GetSchedule(t *testing.T) {
	dsManager := datasource.NewManager()
	checkManager := check.NewManager(dsManager)
	alertManager := alerting.NewManager()
	m := NewManager(checkManager, alertManager)
	ctx := context.Background()

	schedule := &Schedule{
		TenantID:       "tenant-1",
		Name:           "Daily Check",
		CronExpression: CronDaily,
	}
	m.CreateSchedule(ctx, schedule)

	retrieved, err := m.GetSchedule(ctx, schedule.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.Name != schedule.Name {
		t.Errorf("expected name %s, got %s", schedule.Name, retrieved.Name)
	}
}

func TestManager_GetSchedule_NotFound(t *testing.T) {
	dsManager := datasource.NewManager()
	checkManager := check.NewManager(dsManager)
	alertManager := alerting.NewManager()
	m := NewManager(checkManager, alertManager)
	ctx := context.Background()

	_, err := m.GetSchedule(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent schedule")
	}
}

func TestManager_DeleteSchedule(t *testing.T) {
	dsManager := datasource.NewManager()
	checkManager := check.NewManager(dsManager)
	alertManager := alerting.NewManager()
	m := NewManager(checkManager, alertManager)
	ctx := context.Background()

	schedule := &Schedule{
		TenantID:       "tenant-1",
		Name:           "Daily Check",
		CronExpression: CronDaily,
	}
	m.CreateSchedule(ctx, schedule)

	err := m.DeleteSchedule(ctx, schedule.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.GetSchedule(ctx, schedule.ID)
	if err == nil {
		t.Fatal("expected error for deleted schedule")
	}
}

func TestManager_ListSchedules(t *testing.T) {
	dsManager := datasource.NewManager()
	checkManager := check.NewManager(dsManager)
	alertManager := alerting.NewManager()
	m := NewManager(checkManager, alertManager)
	ctx := context.Background()

	m.CreateSchedule(ctx, &Schedule{
		TenantID:       "tenant-1",
		Name:           "Schedule 1",
		CronExpression: CronDaily,
	})
	m.CreateSchedule(ctx, &Schedule{
		TenantID:       "tenant-2",
		Name:           "Schedule 2",
		CronExpression: CronHourly,
	})

	// List all
	schedules, err := m.ListSchedules(ctx, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schedules) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(schedules))
	}

	// Filter by tenant
	schedules, err = m.ListSchedules(ctx, "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(schedules) != 1 {
		t.Errorf("expected 1 schedule for tenant-1, got %d", len(schedules))
	}
}

func TestGetNextRunTime(t *testing.T) {
	testCases := []struct {
		name string
		expr string
	}{
		{"every minute", CronEveryMinute},
		{"every 5 minutes", CronEvery5Minutes},
		{"hourly", CronHourly},
		{"daily", CronDaily},
		{"weekly", CronWeekly},
		{"monthly", CronMonthly},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nextRun, err := getNextRunTime(tc.expr, "UTC")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if nextRun.IsZero() {
				t.Error("next run time should not be zero")
			}
		})
	}
}
