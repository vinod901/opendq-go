// Package check provides data quality check definitions and execution capabilities.
// Supports various check types including row count, null checks, uniqueness,
// freshness, and custom SQL checks.
package check

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vinod901/opendq-go/internal/datasource"
)

// Type represents the type of data quality check
type Type string

const (
	// Basic checks
	TypeRowCount    Type = "row_count"
	TypeNullCheck   Type = "null_check"
	TypeUniqueness  Type = "uniqueness"
	TypeFreshness   Type = "freshness"
	TypeCustomSQL   Type = "custom_sql"
	
	// Value checks
	TypeMinValue    Type = "min_value"
	TypeMaxValue    Type = "max_value"
	TypeMeanValue   Type = "mean_value"
	TypeSumValue    Type = "sum_value"
	TypeStdDev      Type = "std_dev"
	
	// Pattern checks
	TypeRegex       Type = "regex"
	TypeFormat      Type = "format"
	TypeRange       Type = "range"
	TypeSetMembership Type = "set_membership"
	
	// Referential checks
	TypeReferentialIntegrity Type = "referential_integrity"
	TypeVolume               Type = "volume"
	TypeDistribution         Type = "distribution"
	
	// Schema checks
	TypeSchemaMatch Type = "schema_match"
	TypeColumnCount Type = "column_count"
	TypeColumnType  Type = "column_type"
)

// Status represents the status of a check execution
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusPassed    Status = "passed"
	StatusFailed    Status = "failed"
	StatusWarning   Status = "warning"
	StatusError     Status = "error"
	StatusSkipped   Status = "skipped"
)

// Severity represents the severity of a check failure
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Check represents a data quality check configuration
type Check struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	DatasourceID    string                 `json:"datasource_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Type            Type                   `json:"type"`
	Table           string                 `json:"table"`
	Column          string                 `json:"column,omitempty"`
	Parameters      CheckParameters        `json:"parameters"`
	Threshold       Threshold              `json:"threshold"`
	Severity        Severity               `json:"severity"`
	Tags            []string               `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	Active          bool                   `json:"active"`
	ScheduleID      string                 `json:"schedule_id,omitempty"`
	ViewID          string                 `json:"view_id,omitempty"` // For logical view checks
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	LastRunAt       *time.Time             `json:"last_run_at,omitempty"`
	LastStatus      Status                 `json:"last_status,omitempty"`
}

// CheckParameters contains type-specific parameters for checks
type CheckParameters struct {
	// Row count parameters
	MinRows int64 `json:"min_rows,omitempty"`
	MaxRows int64 `json:"max_rows,omitempty"`
	
	// Null check parameters
	MaxNullPercentage float64 `json:"max_null_percentage,omitempty"`
	MaxNullCount      int64   `json:"max_null_count,omitempty"`
	
	// Uniqueness parameters
	UniqueColumns []string `json:"unique_columns,omitempty"`
	
	// Freshness parameters
	MaxAgeHours      float64 `json:"max_age_hours,omitempty"`
	TimestampColumn  string  `json:"timestamp_column,omitempty"`
	
	// Custom SQL parameters
	CustomSQL        string `json:"custom_sql,omitempty"`
	ExpectedValue    string `json:"expected_value,omitempty"`
	
	// Value check parameters
	ExpectedMin      float64  `json:"expected_min,omitempty"`
	ExpectedMax      float64  `json:"expected_max,omitempty"`
	ExpectedMean     float64  `json:"expected_mean,omitempty"`
	Tolerance        float64  `json:"tolerance,omitempty"`
	
	// Pattern check parameters
	Pattern          string   `json:"pattern,omitempty"`
	AllowedValues    []string `json:"allowed_values,omitempty"`
	
	// Referential check parameters
	ReferenceTable   string `json:"reference_table,omitempty"`
	ReferenceColumn  string `json:"reference_column,omitempty"`
	
	// Volume check parameters
	ExpectedVolume    int64   `json:"expected_volume,omitempty"`
	VolumeTolerance   float64 `json:"volume_tolerance,omitempty"`
	
	// Schema check parameters
	ExpectedSchema   []datasource.ColumnInfo `json:"expected_schema,omitempty"`
	ExpectedColumns  int                     `json:"expected_columns,omitempty"`
}

// Threshold defines pass/fail criteria for a check
type Threshold struct {
	Type        ThresholdType `json:"type"`
	Value       float64       `json:"value"`
	MinValue    float64       `json:"min_value,omitempty"`
	MaxValue    float64       `json:"max_value,omitempty"`
	Operator    string        `json:"operator,omitempty"` // eq, ne, lt, lte, gt, gte, between
}

// ThresholdType represents the type of threshold
type ThresholdType string

const (
	ThresholdAbsolute   ThresholdType = "absolute"
	ThresholdPercentage ThresholdType = "percentage"
	ThresholdRange      ThresholdType = "range"
)

// CheckResult represents the result of a check execution
type CheckResult struct {
	ID           string                 `json:"id"`
	CheckID      string                 `json:"check_id"`
	DatasourceID string                 `json:"datasource_id"`
	Status       Status                 `json:"status"`
	ActualValue  interface{}            `json:"actual_value"`
	ExpectedValue interface{}           `json:"expected_value,omitempty"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details"`
	Duration     time.Duration          `json:"duration"`
	Timestamp    time.Time              `json:"timestamp"`
	Error        string                 `json:"error,omitempty"`
}

// Manager handles check operations
type Manager struct {
	checks           map[string]*Check
	results          map[string][]*CheckResult
	datasourceManager *datasource.Manager
}

// NewManager creates a new check manager
func NewManager(dsManager *datasource.Manager) *Manager {
	return &Manager{
		checks:           make(map[string]*Check),
		results:          make(map[string][]*CheckResult),
		datasourceManager: dsManager,
	}
}

// CreateCheck creates a new data quality check
func (m *Manager) CreateCheck(ctx context.Context, check *Check) error {
	if check.ID == "" {
		check.ID = uuid.New().String()
	}
	check.CreatedAt = time.Now()
	check.UpdatedAt = time.Now()
	check.Active = true
	check.LastStatus = StatusPending

	m.checks[check.ID] = check
	return nil
}

// GetCheck retrieves a check by ID
func (m *Manager) GetCheck(ctx context.Context, id string) (*Check, error) {
	check, exists := m.checks[id]
	if !exists {
		return nil, fmt.Errorf("check not found: %s", id)
	}
	return check, nil
}

// UpdateCheck updates a check
func (m *Manager) UpdateCheck(ctx context.Context, id string, updates map[string]interface{}) error {
	check, exists := m.checks[id]
	if !exists {
		return fmt.Errorf("check not found: %s", id)
	}

	if name, ok := updates["name"].(string); ok {
		check.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		check.Description = description
	}
	if active, ok := updates["active"].(bool); ok {
		check.Active = active
	}
	if params, ok := updates["parameters"].(CheckParameters); ok {
		check.Parameters = params
	}
	if threshold, ok := updates["threshold"].(Threshold); ok {
		check.Threshold = threshold
	}
	if severity, ok := updates["severity"].(Severity); ok {
		check.Severity = severity
	}
	if tags, ok := updates["tags"].([]string); ok {
		check.Tags = tags
	}

	check.UpdatedAt = time.Now()
	return nil
}

// DeleteCheck deletes a check
func (m *Manager) DeleteCheck(ctx context.Context, id string) error {
	if _, exists := m.checks[id]; !exists {
		return fmt.Errorf("check not found: %s", id)
	}

	delete(m.checks, id)
	delete(m.results, id)
	return nil
}

// ListChecks lists checks with optional filters
func (m *Manager) ListChecks(ctx context.Context, tenantID, datasourceID string) ([]*Check, error) {
	var result []*Check
	for _, check := range m.checks {
		if tenantID != "" && check.TenantID != tenantID {
			continue
		}
		if datasourceID != "" && check.DatasourceID != datasourceID {
			continue
		}
		result = append(result, check)
	}
	return result, nil
}

// RunCheck executes a data quality check
func (m *Manager) RunCheck(ctx context.Context, id string) (*CheckResult, error) {
	check, err := m.GetCheck(ctx, id)
	if err != nil {
		return nil, err
	}

	if !check.Active {
		return &CheckResult{
			ID:        uuid.New().String(),
			CheckID:   id,
			Status:    StatusSkipped,
			Message:   "check is inactive",
			Timestamp: time.Now(),
		}, nil
	}

	// Get datasource connector
	connector, err := m.datasourceManager.GetConnector(ctx, check.DatasourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource connector: %w", err)
	}

	startTime := time.Now()

	// Execute check based on type
	result, err := m.executeCheck(ctx, check, connector)
	if err != nil {
		result = &CheckResult{
			ID:        uuid.New().String(),
			CheckID:   id,
			DatasourceID: check.DatasourceID,
			Status:    StatusError,
			Message:   fmt.Sprintf("check execution failed: %v", err),
			Error:     err.Error(),
			Timestamp: time.Now(),
			Duration:  time.Since(startTime),
		}
	} else {
		result.ID = uuid.New().String()
		result.CheckID = id
		result.DatasourceID = check.DatasourceID
		result.Duration = time.Since(startTime)
		result.Timestamp = time.Now()
	}

	// Store result
	m.results[id] = append(m.results[id], result)

	// Update check status
	now := time.Now()
	check.LastRunAt = &now
	check.LastStatus = result.Status
	check.UpdatedAt = now

	return result, nil
}

// executeCheck executes the appropriate check based on type
func (m *Manager) executeCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	switch check.Type {
	case TypeRowCount:
		return m.runRowCountCheck(ctx, check, connector)
	case TypeNullCheck:
		return m.runNullCheck(ctx, check, connector)
	case TypeUniqueness:
		return m.runUniquenessCheck(ctx, check, connector)
	case TypeFreshness:
		return m.runFreshnessCheck(ctx, check, connector)
	case TypeCustomSQL:
		return m.runCustomSQLCheck(ctx, check, connector)
	case TypeMinValue, TypeMaxValue, TypeMeanValue, TypeSumValue:
		return m.runValueCheck(ctx, check, connector)
	case TypeRegex:
		return m.runRegexCheck(ctx, check, connector)
	case TypeRange:
		return m.runRangeCheck(ctx, check, connector)
	case TypeSetMembership:
		return m.runSetMembershipCheck(ctx, check, connector)
	case TypeReferentialIntegrity:
		return m.runReferentialCheck(ctx, check, connector)
	case TypeSchemaMatch:
		return m.runSchemaCheck(ctx, check, connector)
	default:
		return nil, fmt.Errorf("unsupported check type: %s", check.Type)
	}
}

// GetCheckResults returns results for a check
func (m *Manager) GetCheckResults(ctx context.Context, checkID string, limit int) ([]*CheckResult, error) {
	results, exists := m.results[checkID]
	if !exists {
		return []*CheckResult{}, nil
	}
	
	if limit > 0 && len(results) > limit {
		return results[len(results)-limit:], nil
	}
	return results, nil
}

// RunChecksForDatasource runs all active checks for a datasource
func (m *Manager) RunChecksForDatasource(ctx context.Context, datasourceID string) ([]*CheckResult, error) {
	checks, err := m.ListChecks(ctx, "", datasourceID)
	if err != nil {
		return nil, err
	}

	var results []*CheckResult
	for _, check := range checks {
		if !check.Active {
			continue
		}
		result, err := m.RunCheck(ctx, check.ID)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}
