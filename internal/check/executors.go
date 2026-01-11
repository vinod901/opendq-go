package check

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/vinod901/opendq-go/internal/datasource"
)

// runRowCountCheck executes a row count check
func (m *Manager) runRowCountCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	count, err := connector.GetRowCount(ctx, check.Table)
	if err != nil {
		return nil, fmt.Errorf("failed to get row count: %w", err)
	}

	result := &CheckResult{
		ActualValue: count,
		Details:     make(map[string]interface{}),
	}

	// Evaluate against thresholds
	params := check.Parameters
	if params.MinRows > 0 && count < params.MinRows {
		result.Status = StatusFailed
		result.ExpectedValue = params.MinRows
		result.Message = fmt.Sprintf("row count %d is below minimum %d", count, params.MinRows)
	} else if params.MaxRows > 0 && count > params.MaxRows {
		result.Status = StatusFailed
		result.ExpectedValue = params.MaxRows
		result.Message = fmt.Sprintf("row count %d exceeds maximum %d", count, params.MaxRows)
	} else {
		result.Status = StatusPassed
		result.Message = fmt.Sprintf("row count %d is within acceptable range", count)
	}

	result.Details["row_count"] = count
	result.Details["min_rows"] = params.MinRows
	result.Details["max_rows"] = params.MaxRows

	return result, nil
}

// runNullCheck executes a null value check
func (m *Manager) runNullCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_count,
			SUM(CASE WHEN %s IS NULL THEN 1 ELSE 0 END) as null_count
		FROM %s`, check.Column, check.Table)

	queryResult, err := connector.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute null check query: %w", err)
	}

	if len(queryResult.Rows) == 0 {
		return nil, fmt.Errorf("null check query returned no results")
	}

	row := queryResult.Rows[0]
	totalCount := toInt64(row["total_count"])
	nullCount := toInt64(row["null_count"])

	var nullPercentage float64
	if totalCount > 0 {
		nullPercentage = float64(nullCount) / float64(totalCount) * 100
	}

	result := &CheckResult{
		ActualValue: nullPercentage,
		Details: map[string]interface{}{
			"total_count":     totalCount,
			"null_count":      nullCount,
			"null_percentage": nullPercentage,
		},
	}

	params := check.Parameters
	if params.MaxNullPercentage > 0 && nullPercentage > params.MaxNullPercentage {
		result.Status = StatusFailed
		result.ExpectedValue = params.MaxNullPercentage
		result.Message = fmt.Sprintf("null percentage %.2f%% exceeds maximum %.2f%%", nullPercentage, params.MaxNullPercentage)
	} else if params.MaxNullCount > 0 && nullCount > params.MaxNullCount {
		result.Status = StatusFailed
		result.ExpectedValue = params.MaxNullCount
		result.Message = fmt.Sprintf("null count %d exceeds maximum %d", nullCount, params.MaxNullCount)
	} else {
		result.Status = StatusPassed
		result.Message = fmt.Sprintf("null percentage %.2f%% is acceptable", nullPercentage)
	}

	return result, nil
}

// runUniquenessCheck executes a uniqueness check
func (m *Manager) runUniquenessCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	columns := check.Column
	if len(check.Parameters.UniqueColumns) > 0 {
		columns = ""
		for i, col := range check.Parameters.UniqueColumns {
			if i > 0 {
				columns += ", "
			}
			columns += col
		}
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_count,
			COUNT(DISTINCT %s) as unique_count
		FROM %s`, columns, check.Table)

	queryResult, err := connector.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute uniqueness check query: %w", err)
	}

	if len(queryResult.Rows) == 0 {
		return nil, fmt.Errorf("uniqueness check query returned no results")
	}

	row := queryResult.Rows[0]
	totalCount := toInt64(row["total_count"])
	uniqueCount := toInt64(row["unique_count"])
	duplicateCount := totalCount - uniqueCount

	var uniquenessPercentage float64
	if totalCount > 0 {
		uniquenessPercentage = float64(uniqueCount) / float64(totalCount) * 100
	}

	result := &CheckResult{
		ActualValue: uniquenessPercentage,
		Details: map[string]interface{}{
			"total_count":          totalCount,
			"unique_count":         uniqueCount,
			"duplicate_count":      duplicateCount,
			"uniqueness_percentage": uniquenessPercentage,
			"columns":              columns,
		},
	}

	// By default, expect 100% uniqueness
	expectedUniqueness := 100.0
	if check.Threshold.Value > 0 {
		expectedUniqueness = check.Threshold.Value
	}

	if uniquenessPercentage < expectedUniqueness {
		result.Status = StatusFailed
		result.ExpectedValue = expectedUniqueness
		result.Message = fmt.Sprintf("uniqueness %.2f%% is below expected %.2f%% (%d duplicates)", 
			uniquenessPercentage, expectedUniqueness, duplicateCount)
	} else {
		result.Status = StatusPassed
		result.Message = fmt.Sprintf("uniqueness %.2f%% meets expectation", uniquenessPercentage)
	}

	return result, nil
}

// runFreshnessCheck executes a data freshness check
func (m *Manager) runFreshnessCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	timestampCol := check.Parameters.TimestampColumn
	if timestampCol == "" {
		return nil, fmt.Errorf("timestamp column not specified for freshness check")
	}

	query := fmt.Sprintf("SELECT MAX(%s) as latest_timestamp FROM %s", timestampCol, check.Table)

	queryResult, err := connector.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute freshness check query: %w", err)
	}

	if len(queryResult.Rows) == 0 {
		return nil, fmt.Errorf("freshness check query returned no results")
	}

	row := queryResult.Rows[0]
	latestTimestamp, ok := row["latest_timestamp"].(time.Time)
	if !ok {
		return nil, fmt.Errorf("failed to parse latest timestamp")
	}

	ageHours := time.Since(latestTimestamp).Hours()
	maxAgeHours := check.Parameters.MaxAgeHours

	result := &CheckResult{
		ActualValue: ageHours,
		Details: map[string]interface{}{
			"latest_timestamp":  latestTimestamp,
			"age_hours":         ageHours,
			"max_age_hours":     maxAgeHours,
			"timestamp_column":  timestampCol,
		},
	}

	if maxAgeHours > 0 && ageHours > maxAgeHours {
		result.Status = StatusFailed
		result.ExpectedValue = maxAgeHours
		result.Message = fmt.Sprintf("data age %.2f hours exceeds maximum %.2f hours", ageHours, maxAgeHours)
	} else {
		result.Status = StatusPassed
		result.Message = fmt.Sprintf("data age %.2f hours is acceptable", ageHours)
	}

	return result, nil
}

// runCustomSQLCheck executes a custom SQL check
func (m *Manager) runCustomSQLCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	query := check.Parameters.CustomSQL
	if query == "" {
		return nil, fmt.Errorf("custom SQL not specified")
	}

	queryResult, err := connector.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute custom SQL: %w", err)
	}

	result := &CheckResult{
		ActualValue: queryResult,
		Details: map[string]interface{}{
			"query":     query,
			"row_count": queryResult.RowCount,
			"columns":   queryResult.Columns,
		},
	}

	// Evaluate result based on expected value or row count
	if check.Parameters.ExpectedValue != "" {
		// Compare first row's first column to expected value
		if len(queryResult.Rows) > 0 && len(queryResult.Columns) > 0 {
			actualValue := fmt.Sprintf("%v", queryResult.Rows[0][queryResult.Columns[0]])
			if actualValue == check.Parameters.ExpectedValue {
				result.Status = StatusPassed
				result.Message = "custom SQL check passed"
			} else {
				result.Status = StatusFailed
				result.ExpectedValue = check.Parameters.ExpectedValue
				result.Message = fmt.Sprintf("expected '%s' but got '%s'", check.Parameters.ExpectedValue, actualValue)
			}
		}
	} else {
		// Check based on whether query returns results
		if queryResult.RowCount > 0 {
			result.Status = StatusPassed
			result.Message = fmt.Sprintf("custom SQL returned %d rows", queryResult.RowCount)
		} else {
			result.Status = StatusFailed
			result.Message = "custom SQL returned no results"
		}
	}

	return result, nil
}

// runValueCheck executes value-based checks (min, max, mean, sum)
func (m *Manager) runValueCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	var aggFunc string
	switch check.Type {
	case TypeMinValue:
		aggFunc = "MIN"
	case TypeMaxValue:
		aggFunc = "MAX"
	case TypeMeanValue:
		aggFunc = "AVG"
	case TypeSumValue:
		aggFunc = "SUM"
	default:
		return nil, fmt.Errorf("unsupported value check type: %s", check.Type)
	}

	query := fmt.Sprintf("SELECT %s(%s) as value FROM %s", aggFunc, check.Column, check.Table)

	queryResult, err := connector.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute value check query: %w", err)
	}

	if len(queryResult.Rows) == 0 {
		return nil, fmt.Errorf("value check query returned no results")
	}

	actualValue := toFloat64(queryResult.Rows[0]["value"])

	result := &CheckResult{
		ActualValue: actualValue,
		Details: map[string]interface{}{
			"function": aggFunc,
			"column":   check.Column,
			"table":    check.Table,
		},
	}

	// Evaluate against expected range
	params := check.Parameters
	tolerance := params.Tolerance
	if tolerance == 0 {
		tolerance = check.Threshold.Value
	}

	switch check.Type {
	case TypeMinValue:
		if params.ExpectedMin > 0 {
			diff := actualValue - params.ExpectedMin
			if diff < -tolerance || diff > tolerance {
				result.Status = StatusFailed
				result.ExpectedValue = params.ExpectedMin
				result.Message = fmt.Sprintf("min value %v differs from expected %v", actualValue, params.ExpectedMin)
			} else {
				result.Status = StatusPassed
				result.Message = fmt.Sprintf("min value %v is within tolerance", actualValue)
			}
		} else {
			result.Status = StatusPassed
			result.Message = fmt.Sprintf("min value is %v", actualValue)
		}
	case TypeMaxValue:
		if params.ExpectedMax > 0 {
			diff := actualValue - params.ExpectedMax
			if diff < -tolerance || diff > tolerance {
				result.Status = StatusFailed
				result.ExpectedValue = params.ExpectedMax
				result.Message = fmt.Sprintf("max value %v differs from expected %v", actualValue, params.ExpectedMax)
			} else {
				result.Status = StatusPassed
				result.Message = fmt.Sprintf("max value %v is within tolerance", actualValue)
			}
		} else {
			result.Status = StatusPassed
			result.Message = fmt.Sprintf("max value is %v", actualValue)
		}
	case TypeMeanValue:
		if params.ExpectedMean > 0 {
			diff := actualValue - params.ExpectedMean
			if diff < -tolerance || diff > tolerance {
				result.Status = StatusFailed
				result.ExpectedValue = params.ExpectedMean
				result.Message = fmt.Sprintf("mean value %v differs from expected %v", actualValue, params.ExpectedMean)
			} else {
				result.Status = StatusPassed
				result.Message = fmt.Sprintf("mean value %v is within tolerance", actualValue)
			}
		} else {
			result.Status = StatusPassed
			result.Message = fmt.Sprintf("mean value is %v", actualValue)
		}
	}

	return result, nil
}

// runRegexCheck executes a regex pattern check
func (m *Manager) runRegexCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	pattern := check.Parameters.Pattern
	if pattern == "" {
		return nil, fmt.Errorf("regex pattern not specified")
	}

	_, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Count rows that don't match the pattern (depends on database regex support)
	// This is a simplified version - actual implementation depends on database
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_count,
			SUM(CASE WHEN %s ~ '%s' THEN 1 ELSE 0 END) as match_count
		FROM %s`, check.Column, pattern, check.Table)

	queryResult, err := connector.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute regex check query: %w", err)
	}

	if len(queryResult.Rows) == 0 {
		return nil, fmt.Errorf("regex check query returned no results")
	}

	row := queryResult.Rows[0]
	totalCount := toInt64(row["total_count"])
	matchCount := toInt64(row["match_count"])
	nonMatchCount := totalCount - matchCount

	var matchPercentage float64
	if totalCount > 0 {
		matchPercentage = float64(matchCount) / float64(totalCount) * 100
	}

	result := &CheckResult{
		ActualValue: matchPercentage,
		Details: map[string]interface{}{
			"total_count":      totalCount,
			"match_count":      matchCount,
			"non_match_count":  nonMatchCount,
			"match_percentage": matchPercentage,
			"pattern":          pattern,
		},
	}

	expectedMatch := 100.0
	if check.Threshold.Value > 0 {
		expectedMatch = check.Threshold.Value
	}

	if matchPercentage < expectedMatch {
		result.Status = StatusFailed
		result.ExpectedValue = expectedMatch
		result.Message = fmt.Sprintf("pattern match %.2f%% is below expected %.2f%%", matchPercentage, expectedMatch)
	} else {
		result.Status = StatusPassed
		result.Message = fmt.Sprintf("pattern match %.2f%% meets expectation", matchPercentage)
	}

	return result, nil
}

// runRangeCheck executes a value range check
func (m *Manager) runRangeCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	params := check.Parameters
	
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_count,
			SUM(CASE WHEN %s >= %f AND %s <= %f THEN 1 ELSE 0 END) as in_range_count
		FROM %s`, check.Column, params.ExpectedMin, check.Column, params.ExpectedMax, check.Table)

	queryResult, err := connector.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute range check query: %w", err)
	}

	if len(queryResult.Rows) == 0 {
		return nil, fmt.Errorf("range check query returned no results")
	}

	row := queryResult.Rows[0]
	totalCount := toInt64(row["total_count"])
	inRangeCount := toInt64(row["in_range_count"])
	outOfRangeCount := totalCount - inRangeCount

	var inRangePercentage float64
	if totalCount > 0 {
		inRangePercentage = float64(inRangeCount) / float64(totalCount) * 100
	}

	result := &CheckResult{
		ActualValue: inRangePercentage,
		Details: map[string]interface{}{
			"total_count":         totalCount,
			"in_range_count":      inRangeCount,
			"out_of_range_count":  outOfRangeCount,
			"in_range_percentage": inRangePercentage,
			"min_value":           params.ExpectedMin,
			"max_value":           params.ExpectedMax,
		},
	}

	expectedInRange := 100.0
	if check.Threshold.Value > 0 {
		expectedInRange = check.Threshold.Value
	}

	if inRangePercentage < expectedInRange {
		result.Status = StatusFailed
		result.ExpectedValue = expectedInRange
		result.Message = fmt.Sprintf("in-range percentage %.2f%% is below expected %.2f%%", inRangePercentage, expectedInRange)
	} else {
		result.Status = StatusPassed
		result.Message = fmt.Sprintf("in-range percentage %.2f%% meets expectation", inRangePercentage)
	}

	return result, nil
}

// runSetMembershipCheck executes a set membership check
func (m *Manager) runSetMembershipCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	allowedValues := check.Parameters.AllowedValues
	if len(allowedValues) == 0 {
		return nil, fmt.Errorf("allowed values not specified for set membership check")
	}

	// Build IN clause
	inClause := ""
	for i, v := range allowedValues {
		if i > 0 {
			inClause += ", "
		}
		inClause += fmt.Sprintf("'%s'", v)
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_count,
			SUM(CASE WHEN %s IN (%s) THEN 1 ELSE 0 END) as valid_count
		FROM %s`, check.Column, inClause, check.Table)

	queryResult, err := connector.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute set membership check query: %w", err)
	}

	if len(queryResult.Rows) == 0 {
		return nil, fmt.Errorf("set membership check query returned no results")
	}

	row := queryResult.Rows[0]
	totalCount := toInt64(row["total_count"])
	validCount := toInt64(row["valid_count"])
	invalidCount := totalCount - validCount

	var validPercentage float64
	if totalCount > 0 {
		validPercentage = float64(validCount) / float64(totalCount) * 100
	}

	result := &CheckResult{
		ActualValue: validPercentage,
		Details: map[string]interface{}{
			"total_count":      totalCount,
			"valid_count":      validCount,
			"invalid_count":    invalidCount,
			"valid_percentage": validPercentage,
			"allowed_values":   allowedValues,
		},
	}

	expectedValid := 100.0
	if check.Threshold.Value > 0 {
		expectedValid = check.Threshold.Value
	}

	if validPercentage < expectedValid {
		result.Status = StatusFailed
		result.ExpectedValue = expectedValid
		result.Message = fmt.Sprintf("valid percentage %.2f%% is below expected %.2f%%", validPercentage, expectedValid)
	} else {
		result.Status = StatusPassed
		result.Message = fmt.Sprintf("valid percentage %.2f%% meets expectation", validPercentage)
	}

	return result, nil
}

// runReferentialCheck executes a referential integrity check
func (m *Manager) runReferentialCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	params := check.Parameters
	if params.ReferenceTable == "" || params.ReferenceColumn == "" {
		return nil, fmt.Errorf("reference table/column not specified")
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_count,
			COUNT(r.%s) as matched_count
		FROM %s t
		LEFT JOIN %s r ON t.%s = r.%s`,
		params.ReferenceColumn,
		check.Table,
		params.ReferenceTable,
		check.Column,
		params.ReferenceColumn)

	queryResult, err := connector.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute referential check query: %w", err)
	}

	if len(queryResult.Rows) == 0 {
		return nil, fmt.Errorf("referential check query returned no results")
	}

	row := queryResult.Rows[0]
	totalCount := toInt64(row["total_count"])
	matchedCount := toInt64(row["matched_count"])
	orphanCount := totalCount - matchedCount

	var integrityPercentage float64
	if totalCount > 0 {
		integrityPercentage = float64(matchedCount) / float64(totalCount) * 100
	}

	result := &CheckResult{
		ActualValue: integrityPercentage,
		Details: map[string]interface{}{
			"total_count":          totalCount,
			"matched_count":        matchedCount,
			"orphan_count":         orphanCount,
			"integrity_percentage": integrityPercentage,
			"reference_table":      params.ReferenceTable,
			"reference_column":     params.ReferenceColumn,
		},
	}

	expectedIntegrity := 100.0
	if check.Threshold.Value > 0 {
		expectedIntegrity = check.Threshold.Value
	}

	if integrityPercentage < expectedIntegrity {
		result.Status = StatusFailed
		result.ExpectedValue = expectedIntegrity
		result.Message = fmt.Sprintf("referential integrity %.2f%% is below expected %.2f%% (%d orphans)", 
			integrityPercentage, expectedIntegrity, orphanCount)
	} else {
		result.Status = StatusPassed
		result.Message = fmt.Sprintf("referential integrity %.2f%% meets expectation", integrityPercentage)
	}

	return result, nil
}

// runSchemaCheck executes a schema validation check
func (m *Manager) runSchemaCheck(ctx context.Context, check *Check, connector datasource.Connector) (*CheckResult, error) {
	actualColumns, err := connector.GetColumns(ctx, check.Table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table columns: %w", err)
	}

	expectedSchema := check.Parameters.ExpectedSchema
	
	result := &CheckResult{
		ActualValue: len(actualColumns),
		Details: map[string]interface{}{
			"actual_columns":   actualColumns,
			"expected_columns": expectedSchema,
		},
	}

	if len(expectedSchema) == 0 {
		// Just checking if we can read schema
		result.Status = StatusPassed
		result.Message = fmt.Sprintf("schema has %d columns", len(actualColumns))
		return result, nil
	}

	// Compare schemas
	missingColumns := []string{}
	extraColumns := []string{}
	typeMismatches := []string{}

	expectedMap := make(map[string]datasource.ColumnInfo)
	for _, col := range expectedSchema {
		expectedMap[col.Name] = col
	}

	actualMap := make(map[string]datasource.ColumnInfo)
	for _, col := range actualColumns {
		actualMap[col.Name] = col
	}

	for name, expected := range expectedMap {
		actual, exists := actualMap[name]
		if !exists {
			missingColumns = append(missingColumns, name)
		} else if expected.DataType != "" && expected.DataType != actual.DataType {
			typeMismatches = append(typeMismatches, fmt.Sprintf("%s: expected %s, got %s", name, expected.DataType, actual.DataType))
		}
	}

	for name := range actualMap {
		if _, exists := expectedMap[name]; !exists {
			extraColumns = append(extraColumns, name)
		}
	}

	result.Details["missing_columns"] = missingColumns
	result.Details["extra_columns"] = extraColumns
	result.Details["type_mismatches"] = typeMismatches

	if len(missingColumns) > 0 || len(typeMismatches) > 0 {
		result.Status = StatusFailed
		result.Message = fmt.Sprintf("schema mismatch: %d missing columns, %d type mismatches", 
			len(missingColumns), len(typeMismatches))
	} else {
		result.Status = StatusPassed
		result.Message = "schema matches expected"
	}

	return result, nil
}

// Helper functions

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return 0
	}
}
