# Data Quality Checks

OpenDQ provides a comprehensive set of data quality checks to validate data integrity, completeness, and accuracy.

## Overview

```
┌────────────────────────────────────────────────────────────────────────────┐
│                        Check Execution Flow                                 │
│                                                                            │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                 │
│  │ Check Config │───▶│ Check Manager│───▶│  Connector   │                 │
│  │              │    │              │    │  (Query DB)  │                 │
│  └──────────────┘    └──────┬───────┘    └──────────────┘                 │
│                             │                                              │
│                             ▼                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                        Check Executors                                │ │
│  │                                                                       │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐     │ │
│  │  │ Row Count  │  │ Null Check │  │ Uniqueness │  │ Freshness  │     │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘     │ │
│  │                                                                       │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐     │ │
│  │  │ Value Chks │  │ Regex/Fmt  │  │ Ref Integ  │  │ Schema     │     │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘     │ │
│  │                                                                       │ │
│  │  ┌────────────┐                                                      │ │
│  │  │ Custom SQL │                                                      │ │
│  │  └────────────┘                                                      │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                             │                                              │
│                             ▼                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                 │
│  │ Check Result │───▶│ Threshold    │───▶│ Alert        │                 │
│  │              │    │ Evaluation   │    │ (if failed)  │                 │
│  └──────────────┘    └──────────────┘    └──────────────┘                 │
└────────────────────────────────────────────────────────────────────────────┘
```

## Check Types

### Basic Checks

| Type | Purpose | Example Use Case |
|------|---------|------------------|
| `row_count` | Validate table has expected row count | Ensure ETL loaded data |
| `null_check` | Check for null values in columns | Data completeness |
| `uniqueness` | Validate unique values | Primary key validity |
| `freshness` | Check data recency | Data pipeline latency |
| `custom_sql` | Custom SQL validation | Business logic validation |

### Value Checks

| Type | Purpose | Example Use Case |
|------|---------|------------------|
| `min_value` | Validate minimum value | Price >= 0 |
| `max_value` | Validate maximum value | Age <= 150 |
| `mean_value` | Validate average | Average order value |
| `sum_value` | Validate sum | Total revenue |
| `std_dev` | Validate standard deviation | Outlier detection |

### Pattern Checks

| Type | Purpose | Example Use Case |
|------|---------|------------------|
| `regex` | Match regex pattern | Email format |
| `format` | Match expected format | Date format |
| `range` | Value within range | Score between 0-100 |
| `set_membership` | Value in allowed set | Status values |

### Referential Checks

| Type | Purpose | Example Use Case |
|------|---------|------------------|
| `referential_integrity` | Foreign key validation | Orders reference customers |
| `volume` | Expected data volume | Daily transaction volume |
| `distribution` | Value distribution | Category distribution |

### Schema Checks

| Type | Purpose | Example Use Case |
|------|---------|------------------|
| `schema_match` | Schema matches expected | Schema drift detection |
| `column_count` | Expected column count | Structure validation |
| `column_type` | Column type validation | Type compatibility |

## Check Model

```go
// internal/check/check.go

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
    ViewID          string                 `json:"view_id,omitempty"`
    CreatedAt       time.Time              `json:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at"`
    LastRunAt       *time.Time             `json:"last_run_at,omitempty"`
    LastStatus      Status                 `json:"last_status,omitempty"`
}

type CheckParameters struct {
    // Row count
    MinRows int64 `json:"min_rows,omitempty"`
    MaxRows int64 `json:"max_rows,omitempty"`
    
    // Null check
    MaxNullPercentage float64 `json:"max_null_percentage,omitempty"`
    MaxNullCount      int64   `json:"max_null_count,omitempty"`
    
    // Uniqueness
    UniqueColumns []string `json:"unique_columns,omitempty"`
    
    // Freshness
    MaxAgeHours     float64 `json:"max_age_hours,omitempty"`
    TimestampColumn string  `json:"timestamp_column,omitempty"`
    
    // Custom SQL
    CustomSQL     string `json:"custom_sql,omitempty"`
    ExpectedValue string `json:"expected_value,omitempty"`
    
    // Value checks
    ExpectedMin  float64 `json:"expected_min,omitempty"`
    ExpectedMax  float64 `json:"expected_max,omitempty"`
    ExpectedMean float64 `json:"expected_mean,omitempty"`
    Tolerance    float64 `json:"tolerance,omitempty"`
    
    // Pattern checks
    Pattern       string   `json:"pattern,omitempty"`
    AllowedValues []string `json:"allowed_values,omitempty"`
    
    // Referential
    ReferenceTable  string `json:"reference_table,omitempty"`
    ReferenceColumn string `json:"reference_column,omitempty"`
    
    // Schema
    ExpectedSchema  []ColumnInfo `json:"expected_schema,omitempty"`
    ExpectedColumns int          `json:"expected_columns,omitempty"`
}

type Threshold struct {
    Type     ThresholdType `json:"type"`      // absolute, percentage, range
    Value    float64       `json:"value"`
    MinValue float64       `json:"min_value,omitempty"`
    MaxValue float64       `json:"max_value,omitempty"`
    Operator string        `json:"operator,omitempty"` // eq, ne, lt, lte, gt, gte, between
}
```

## Check Results

```go
type CheckResult struct {
    ID            string                 `json:"id"`
    CheckID       string                 `json:"check_id"`
    DatasourceID  string                 `json:"datasource_id"`
    Status        Status                 `json:"status"`
    ActualValue   interface{}            `json:"actual_value"`
    ExpectedValue interface{}            `json:"expected_value,omitempty"`
    Message       string                 `json:"message"`
    Details       map[string]interface{} `json:"details"`
    Duration      time.Duration          `json:"duration"`
    Timestamp     time.Time              `json:"timestamp"`
    Error         string                 `json:"error,omitempty"`
}

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

type Severity string

const (
    SeverityCritical Severity = "critical"
    SeverityHigh     Severity = "high"
    SeverityMedium   Severity = "medium"
    SeverityLow      Severity = "low"
    SeverityInfo     Severity = "info"
)
```

## Check Executors

### Row Count Check

```go
func (m *Manager) runRowCountCheck(ctx context.Context, check *Check, conn Connector) (*CheckResult, error) {
    rowCount, err := conn.GetRowCount(ctx, check.Table)
    if err != nil {
        return nil, err
    }
    
    result := &CheckResult{
        ActualValue: rowCount,
        Details:     map[string]interface{}{"table": check.Table},
    }
    
    // Evaluate thresholds
    if check.Parameters.MinRows > 0 && rowCount < check.Parameters.MinRows {
        result.Status = StatusFailed
        result.Message = fmt.Sprintf("Row count %d is below minimum %d", 
            rowCount, check.Parameters.MinRows)
    } else if check.Parameters.MaxRows > 0 && rowCount > check.Parameters.MaxRows {
        result.Status = StatusFailed
        result.Message = fmt.Sprintf("Row count %d exceeds maximum %d",
            rowCount, check.Parameters.MaxRows)
    } else {
        result.Status = StatusPassed
        result.Message = fmt.Sprintf("Row count %d is within expected range", rowCount)
    }
    
    return result, nil
}
```

### Null Check

```go
func (m *Manager) runNullCheck(ctx context.Context, check *Check, conn Connector) (*CheckResult, error) {
    query := fmt.Sprintf(`
        SELECT 
            COUNT(*) as total,
            COUNT(%s) as non_null,
            COUNT(*) - COUNT(%s) as null_count
        FROM %s
    `, check.Column, check.Column, check.Table)
    
    qr, err := conn.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    
    row := qr.Rows[0]
    total := row["total"].(int64)
    nullCount := row["null_count"].(int64)
    nullPct := float64(nullCount) / float64(total) * 100
    
    result := &CheckResult{
        ActualValue: nullPct,
        Details: map[string]interface{}{
            "total_rows":  total,
            "null_count":  nullCount,
            "null_percent": nullPct,
        },
    }
    
    // Evaluate threshold
    if check.Parameters.MaxNullPercentage > 0 && nullPct > check.Parameters.MaxNullPercentage {
        result.Status = StatusFailed
        result.Message = fmt.Sprintf("Null percentage %.2f%% exceeds maximum %.2f%%",
            nullPct, check.Parameters.MaxNullPercentage)
    } else {
        result.Status = StatusPassed
        result.Message = fmt.Sprintf("Null percentage %.2f%% is within threshold", nullPct)
    }
    
    return result, nil
}
```

### Freshness Check

```go
func (m *Manager) runFreshnessCheck(ctx context.Context, check *Check, conn Connector) (*CheckResult, error) {
    query := fmt.Sprintf(`
        SELECT MAX(%s) as max_timestamp FROM %s
    `, check.Parameters.TimestampColumn, check.Table)
    
    qr, err := conn.Query(ctx, query)
    if err != nil {
        return nil, err
    }
    
    maxTime := qr.Rows[0]["max_timestamp"].(time.Time)
    age := time.Since(maxTime)
    ageHours := age.Hours()
    
    result := &CheckResult{
        ActualValue: ageHours,
        Details: map[string]interface{}{
            "max_timestamp": maxTime,
            "age_hours":     ageHours,
        },
    }
    
    if ageHours > check.Parameters.MaxAgeHours {
        result.Status = StatusFailed
        result.Message = fmt.Sprintf("Data is %.2f hours old, exceeds %.2f hour limit",
            ageHours, check.Parameters.MaxAgeHours)
    } else {
        result.Status = StatusPassed
        result.Message = fmt.Sprintf("Data is %.2f hours old, within %.2f hour limit",
            ageHours, check.Parameters.MaxAgeHours)
    }
    
    return result, nil
}
```

### Custom SQL Check

```go
func (m *Manager) runCustomSQLCheck(ctx context.Context, check *Check, conn Connector) (*CheckResult, error) {
    qr, err := conn.Query(ctx, check.Parameters.CustomSQL)
    if err != nil {
        return nil, err
    }
    
    // Get first value from result
    var actualValue interface{}
    if len(qr.Rows) > 0 && len(qr.Columns) > 0 {
        actualValue = qr.Rows[0][qr.Columns[0]]
    }
    
    result := &CheckResult{
        ActualValue:   actualValue,
        ExpectedValue: check.Parameters.ExpectedValue,
        Details: map[string]interface{}{
            "query":    check.Parameters.CustomSQL,
            "row_count": len(qr.Rows),
        },
    }
    
    // Compare with expected value
    actualStr := fmt.Sprintf("%v", actualValue)
    if actualStr == check.Parameters.ExpectedValue {
        result.Status = StatusPassed
        result.Message = "Custom SQL check passed"
    } else {
        result.Status = StatusFailed
        result.Message = fmt.Sprintf("Expected %v but got %v",
            check.Parameters.ExpectedValue, actualValue)
    }
    
    return result, nil
}
```

## API Examples

### Create Row Count Check

```bash
curl -X POST http://localhost:8080/api/v1/checks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Users Table Minimum Rows",
    "datasource_id": "ds-123",
    "type": "row_count",
    "table": "users",
    "severity": "high",
    "parameters": {
      "min_rows": 1000
    }
  }'
```

### Create Null Check

```bash
curl -X POST http://localhost:8080/api/v1/checks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Email Not Null",
    "datasource_id": "ds-123",
    "type": "null_check",
    "table": "users",
    "column": "email",
    "severity": "critical",
    "parameters": {
      "max_null_percentage": 0
    }
  }'
```

### Create Freshness Check

```bash
curl -X POST http://localhost:8080/api/v1/checks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Orders Data Freshness",
    "datasource_id": "ds-123",
    "type": "freshness",
    "table": "orders",
    "severity": "high",
    "parameters": {
      "timestamp_column": "created_at",
      "max_age_hours": 24
    }
  }'
```

### Create Custom SQL Check

```bash
curl -X POST http://localhost:8080/api/v1/checks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Active Users Count",
    "datasource_id": "ds-123",
    "type": "custom_sql",
    "table": "users",
    "severity": "medium",
    "parameters": {
      "custom_sql": "SELECT COUNT(*) FROM users WHERE status = '\''active'\''",
      "expected_value": "1000"
    },
    "threshold": {
      "type": "absolute",
      "operator": "gte",
      "value": 1000
    }
  }'
```

### Run Check

```bash
curl -X POST http://localhost:8080/api/v1/checks/check-123/run

Response:
{
    "id": "result-456",
    "check_id": "check-123",
    "status": "passed",
    "actual_value": 1500,
    "message": "Row count 1500 is within expected range",
    "duration": "125ms",
    "timestamp": "2024-01-15T10:30:00Z"
}
```

### Get Check Results

```bash
curl http://localhost:8080/api/v1/checks/check-123/results

Response:
[
    {
        "id": "result-456",
        "check_id": "check-123",
        "status": "passed",
        "actual_value": 1500,
        "timestamp": "2024-01-15T10:30:00Z"
    },
    {
        "id": "result-455",
        "check_id": "check-123",
        "status": "failed",
        "actual_value": 800,
        "timestamp": "2024-01-14T10:30:00Z"
    }
]
```

## Check Manager

```go
type Manager struct {
    checks            map[string]*Check
    results           map[string][]*CheckResult
    datasourceManager *datasource.Manager
}

func (m *Manager) RunCheck(ctx context.Context, id string) (*CheckResult, error) {
    check, err := m.GetCheck(ctx, id)
    if err != nil {
        return nil, err
    }
    
    if !check.Active {
        return &CheckResult{
            Status:  StatusSkipped,
            Message: "check is inactive",
        }, nil
    }
    
    // Get connector
    connector, err := m.datasourceManager.GetConnector(ctx, check.DatasourceID)
    if err != nil {
        return nil, err
    }
    
    startTime := time.Now()
    
    // Execute based on type
    result, err := m.executeCheck(ctx, check, connector)
    if err != nil {
        result = &CheckResult{
            Status:  StatusError,
            Error:   err.Error(),
            Message: "check execution failed",
        }
    }
    
    result.ID = uuid.New().String()
    result.CheckID = id
    result.Duration = time.Since(startTime)
    result.Timestamp = time.Now()
    
    // Store result
    m.results[id] = append(m.results[id], result)
    
    // Update check status
    check.LastRunAt = &result.Timestamp
    check.LastStatus = result.Status
    
    return result, nil
}
```

## Best Practices

1. **Start Simple**: Begin with basic checks (row count, nulls)
2. **Critical First**: Implement critical checks first
3. **Appropriate Thresholds**: Set realistic thresholds
4. **Clear Names**: Use descriptive check names
5. **Tagging**: Use tags for organization
6. **Documentation**: Document business context in description
7. **Scheduling**: Schedule checks during low-traffic periods
8. **Alerting**: Configure alerts for critical checks
