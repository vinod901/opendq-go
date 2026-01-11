package check

import (
	"context"
	"testing"

	"github.com/vinod901/opendq-go/internal/datasource"
)

func TestNewManager(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.checks == nil {
		t.Fatal("checks map is nil")
	}
	if m.results == nil {
		t.Fatal("results map is nil")
	}
}

func TestCheckType_Values(t *testing.T) {
	types := []Type{
		TypeRowCount,
		TypeNullCheck,
		TypeUniqueness,
		TypeFreshness,
		TypeCustomSQL,
		TypeMinValue,
		TypeMaxValue,
		TypeMeanValue,
		TypeRegex,
		TypeRange,
		TypeSetMembership,
		TypeReferentialIntegrity,
		TypeSchemaMatch,
	}

	for _, checkType := range types {
		if checkType == "" {
			t.Error("check type should not be empty")
		}
	}
}

func TestCheckStatus_Values(t *testing.T) {
	statuses := []Status{
		StatusPending,
		StatusRunning,
		StatusPassed,
		StatusFailed,
		StatusWarning,
		StatusError,
		StatusSkipped,
	}

	for _, status := range statuses {
		if status == "" {
			t.Error("status should not be empty")
		}
	}
}

func TestCheckSeverity_Values(t *testing.T) {
	severities := []Severity{
		SeverityCritical,
		SeverityHigh,
		SeverityMedium,
		SeverityLow,
		SeverityInfo,
	}

	for _, severity := range severities {
		if severity == "" {
			t.Error("severity should not be empty")
		}
	}
}

func TestManager_CreateCheck(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	check := &Check{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Test Check",
		Type:         TypeRowCount,
		Table:        "users",
	}

	err := m.CreateCheck(ctx, check)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if check.ID == "" {
		t.Error("check ID should be generated")
	}
	if !check.Active {
		t.Error("check should be active by default")
	}
	if check.LastStatus != StatusPending {
		t.Errorf("expected status %s, got %s", StatusPending, check.LastStatus)
	}
}

func TestManager_GetCheck(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	check := &Check{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Test Check",
		Type:         TypeRowCount,
		Table:        "users",
	}
	m.CreateCheck(ctx, check)

	retrieved, err := m.GetCheck(ctx, check.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.Name != check.Name {
		t.Errorf("expected name %s, got %s", check.Name, retrieved.Name)
	}
}

func TestManager_GetCheck_NotFound(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	_, err := m.GetCheck(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent check")
	}
}

func TestManager_UpdateCheck(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	check := &Check{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Test Check",
		Type:         TypeRowCount,
		Table:        "users",
	}
	m.CreateCheck(ctx, check)

	err := m.UpdateCheck(ctx, check.ID, map[string]interface{}{
		"name": "Updated Check",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := m.GetCheck(ctx, check.ID)
	if updated.Name != "Updated Check" {
		t.Errorf("expected name 'Updated Check', got %s", updated.Name)
	}
}

func TestManager_DeleteCheck(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	check := &Check{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Test Check",
		Type:         TypeRowCount,
		Table:        "users",
	}
	m.CreateCheck(ctx, check)

	err := m.DeleteCheck(ctx, check.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.GetCheck(ctx, check.ID)
	if err == nil {
		t.Fatal("expected error for deleted check")
	}
}

func TestManager_ListChecks(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	m.CreateCheck(ctx, &Check{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Check 1",
		Type:         TypeRowCount,
		Table:        "users",
	})
	m.CreateCheck(ctx, &Check{
		TenantID:     "tenant-1",
		DatasourceID: "ds-2",
		Name:         "Check 2",
		Type:         TypeNullCheck,
		Table:        "orders",
	})

	// List all
	checks, err := m.ListChecks(ctx, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(checks) != 2 {
		t.Errorf("expected 2 checks, got %d", len(checks))
	}

	// Filter by datasource
	checks, err = m.ListChecks(ctx, "", "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(checks) != 1 {
		t.Errorf("expected 1 check for ds-1, got %d", len(checks))
	}
}

func TestThresholdType_Values(t *testing.T) {
	types := []ThresholdType{
		ThresholdAbsolute,
		ThresholdPercentage,
		ThresholdRange,
	}

	for _, thresholdType := range types {
		if thresholdType == "" {
			t.Error("threshold type should not be empty")
		}
	}
}
