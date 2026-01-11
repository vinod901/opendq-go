// Package view provides logical view abstractions for datasources.
// Allows defining virtual views that don't exist in the database but can be
// used for data quality checks.
package view

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vinod901/opendq-go/internal/datasource"
)

// View represents a logical view definition that can be used for checks
type View struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	DatasourceID string                 `json:"datasource_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Definition   ViewDefinition         `json:"definition"`
	Schema       []datasource.ColumnInfo `json:"schema,omitempty"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	Active       bool                   `json:"active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	ValidatedAt  *time.Time             `json:"validated_at,omitempty"`
}

// ViewDefinition defines how the logical view is constructed
type ViewDefinition struct {
	// SQL-based view definition
	SQL string `json:"sql,omitempty"`
	
	// Table-based view with transformations
	BaseTable    string            `json:"base_table,omitempty"`
	Columns      []ColumnDef       `json:"columns,omitempty"`
	Filters      []FilterDef       `json:"filters,omitempty"`
	Joins        []JoinDef         `json:"joins,omitempty"`
	GroupBy      []string          `json:"group_by,omitempty"`
	OrderBy      []OrderByDef      `json:"order_by,omitempty"`
	Limit        int               `json:"limit,omitempty"`
	
	// Union of multiple tables/views
	UnionTables  []string          `json:"union_tables,omitempty"`
	UnionAll     bool              `json:"union_all,omitempty"`
}

// ColumnDef defines a column in the view
type ColumnDef struct {
	Name        string `json:"name"`
	Expression  string `json:"expression,omitempty"` // SQL expression or column reference
	SourceColumn string `json:"source_column,omitempty"`
	Alias       string `json:"alias,omitempty"`
	DataType    string `json:"data_type,omitempty"`
}

// FilterDef defines a filter condition
type FilterDef struct {
	Column   string      `json:"column"`
	Operator string      `json:"operator"` // eq, ne, lt, lte, gt, gte, in, not_in, like, is_null, is_not_null
	Value    interface{} `json:"value,omitempty"`
	Values   []interface{} `json:"values,omitempty"` // For in/not_in operators
	LogicalOp string     `json:"logical_op,omitempty"` // AND, OR (for combining with previous filter)
}

// JoinDef defines a join with another table
type JoinDef struct {
	Table     string   `json:"table"`
	Type      string   `json:"type"` // inner, left, right, full, cross
	OnColumns []string `json:"on_columns,omitempty"` // Pairs of columns [left1, right1, left2, right2, ...]
	OnCondition string `json:"on_condition,omitempty"` // Custom join condition
}

// OrderByDef defines ordering
type OrderByDef struct {
	Column string `json:"column"`
	Direction string `json:"direction"` // asc, desc
}

// Manager handles view operations
type Manager struct {
	views             map[string]*View
	datasourceManager *datasource.Manager
}

// NewManager creates a new view manager
func NewManager(dsManager *datasource.Manager) *Manager {
	return &Manager{
		views:             make(map[string]*View),
		datasourceManager: dsManager,
	}
}

// CreateView creates a new logical view
func (m *Manager) CreateView(ctx context.Context, view *View) error {
	if view.ID == "" {
		view.ID = uuid.New().String()
	}
	view.CreatedAt = time.Now()
	view.UpdatedAt = time.Now()
	view.Active = true

	// Validate view definition
	if err := m.validateViewDefinition(ctx, view); err != nil {
		return fmt.Errorf("invalid view definition: %w", err)
	}

	// Infer schema if not provided
	if len(view.Schema) == 0 {
		schema, err := m.inferSchema(ctx, view)
		if err != nil {
			// Schema inference is optional, log warning but continue
			fmt.Printf("Warning: could not infer schema for view %s: %v\n", view.Name, err)
		} else {
			view.Schema = schema
		}
	}

	m.views[view.ID] = view
	return nil
}

// GetView retrieves a view by ID
func (m *Manager) GetView(ctx context.Context, id string) (*View, error) {
	view, exists := m.views[id]
	if !exists {
		return nil, fmt.Errorf("view not found: %s", id)
	}
	return view, nil
}

// UpdateView updates a view
func (m *Manager) UpdateView(ctx context.Context, id string, updates map[string]interface{}) error {
	view, exists := m.views[id]
	if !exists {
		return fmt.Errorf("view not found: %s", id)
	}

	if name, ok := updates["name"].(string); ok {
		view.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		view.Description = description
	}
	if active, ok := updates["active"].(bool); ok {
		view.Active = active
	}
	if definition, ok := updates["definition"].(ViewDefinition); ok {
		view.Definition = definition
		// Re-validate and re-infer schema
		if err := m.validateViewDefinition(ctx, view); err != nil {
			return fmt.Errorf("invalid view definition: %w", err)
		}
		schema, _ := m.inferSchema(ctx, view)
		view.Schema = schema
	}
	if tags, ok := updates["tags"].([]string); ok {
		view.Tags = tags
	}

	view.UpdatedAt = time.Now()
	return nil
}

// DeleteView deletes a view
func (m *Manager) DeleteView(ctx context.Context, id string) error {
	if _, exists := m.views[id]; !exists {
		return fmt.Errorf("view not found: %s", id)
	}
	delete(m.views, id)
	return nil
}

// ListViews lists views with optional filters
func (m *Manager) ListViews(ctx context.Context, tenantID, datasourceID string) ([]*View, error) {
	var result []*View
	for _, view := range m.views {
		if tenantID != "" && view.TenantID != tenantID {
			continue
		}
		if datasourceID != "" && view.DatasourceID != datasourceID {
			continue
		}
		result = append(result, view)
	}
	return result, nil
}

// GetViewSQL returns the SQL representation of a view
func (m *Manager) GetViewSQL(ctx context.Context, id string) (string, error) {
	view, err := m.GetView(ctx, id)
	if err != nil {
		return "", err
	}

	return m.buildViewSQL(view)
}

// QueryView executes the view and returns results
func (m *Manager) QueryView(ctx context.Context, id string, limit int) (*datasource.QueryResult, error) {
	view, err := m.GetView(ctx, id)
	if err != nil {
		return nil, err
	}

	if !view.Active {
		return nil, fmt.Errorf("view is inactive")
	}

	connector, err := m.datasourceManager.GetConnector(ctx, view.DatasourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get datasource connector: %w", err)
	}

	sql, err := m.buildViewSQL(view)
	if err != nil {
		return nil, fmt.Errorf("failed to build view SQL: %w", err)
	}

	if limit > 0 {
		sql = fmt.Sprintf("SELECT * FROM (%s) _view LIMIT %d", sql, limit)
	}

	return connector.Query(ctx, sql)
}

// GetViewRowCount returns the row count for a view
func (m *Manager) GetViewRowCount(ctx context.Context, id string) (int64, error) {
	view, err := m.GetView(ctx, id)
	if err != nil {
		return 0, err
	}

	connector, err := m.datasourceManager.GetConnector(ctx, view.DatasourceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get datasource connector: %w", err)
	}

	sql, err := m.buildViewSQL(view)
	if err != nil {
		return 0, fmt.Errorf("failed to build view SQL: %w", err)
	}

	countSQL := fmt.Sprintf("SELECT COUNT(*) as count FROM (%s) _view", sql)
	result, err := connector.Query(ctx, countSQL)
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	if len(result.Rows) > 0 {
		if count, ok := result.Rows[0]["count"].(int64); ok {
			return count, nil
		}
	}

	return 0, nil
}

// ValidateView validates a view definition
func (m *Manager) ValidateView(ctx context.Context, id string) error {
	view, err := m.GetView(ctx, id)
	if err != nil {
		return err
	}

	if err := m.validateViewDefinition(ctx, view); err != nil {
		return err
	}

	// Try to execute with limit 0 to validate SQL
	connector, err := m.datasourceManager.GetConnector(ctx, view.DatasourceID)
	if err != nil {
		return fmt.Errorf("failed to get datasource connector: %w", err)
	}

	sql, err := m.buildViewSQL(view)
	if err != nil {
		return fmt.Errorf("failed to build view SQL: %w", err)
	}

	// Execute with LIMIT 0 to validate without returning data
	validateSQL := fmt.Sprintf("SELECT * FROM (%s) _view LIMIT 0", sql)
	if _, err := connector.Query(ctx, validateSQL); err != nil {
		return fmt.Errorf("view validation failed: %w", err)
	}

	now := time.Now()
	view.ValidatedAt = &now
	view.UpdatedAt = now

	return nil
}

// validateViewDefinition validates the view definition structure
func (m *Manager) validateViewDefinition(ctx context.Context, view *View) error {
	def := view.Definition

	// Must have either SQL or base table
	if def.SQL == "" && def.BaseTable == "" && len(def.UnionTables) == 0 {
		return fmt.Errorf("view must have SQL, base table, or union tables defined")
	}

	// Validate joins
	for i, join := range def.Joins {
		if join.Table == "" {
			return fmt.Errorf("join %d: table is required", i)
		}
		if join.Type == "" {
			return fmt.Errorf("join %d: type is required", i)
		}
		validJoinTypes := map[string]bool{
			"inner": true, "left": true, "right": true, "full": true, "cross": true,
		}
		if !validJoinTypes[join.Type] {
			return fmt.Errorf("join %d: invalid type '%s'", i, join.Type)
		}
		if len(join.OnColumns) == 0 && join.OnCondition == "" && join.Type != "cross" {
			return fmt.Errorf("join %d: on condition is required for non-cross joins", i)
		}
	}

	// Validate filters
	for i, filter := range def.Filters {
		if filter.Column == "" {
			return fmt.Errorf("filter %d: column is required", i)
		}
		if filter.Operator == "" {
			return fmt.Errorf("filter %d: operator is required", i)
		}
		validOperators := map[string]bool{
			"eq": true, "ne": true, "lt": true, "lte": true, "gt": true, "gte": true,
			"in": true, "not_in": true, "like": true, "is_null": true, "is_not_null": true,
		}
		if !validOperators[filter.Operator] {
			return fmt.Errorf("filter %d: invalid operator '%s'", i, filter.Operator)
		}
	}

	return nil
}

// inferSchema infers the schema of a view
func (m *Manager) inferSchema(ctx context.Context, view *View) ([]datasource.ColumnInfo, error) {
	connector, err := m.datasourceManager.GetConnector(ctx, view.DatasourceID)
	if err != nil {
		return nil, err
	}

	sql, err := m.buildViewSQL(view)
	if err != nil {
		return nil, err
	}

	// Execute with LIMIT 0 to get column info
	result, err := connector.Query(ctx, fmt.Sprintf("SELECT * FROM (%s) _view LIMIT 0", sql))
	if err != nil {
		return nil, err
	}

	var schema []datasource.ColumnInfo
	for _, col := range result.Columns {
		schema = append(schema, datasource.ColumnInfo{
			Name: col,
		})
	}

	return schema, nil
}

// buildViewSQL builds the SQL query for a view
func (m *Manager) buildViewSQL(view *View) (string, error) {
	def := view.Definition

	// If raw SQL is provided, use it directly
	if def.SQL != "" {
		return def.SQL, nil
	}

	// Build SQL from definition
	if len(def.UnionTables) > 0 {
		return m.buildUnionSQL(def)
	}

	return m.buildSelectSQL(def)
}

// buildSelectSQL builds a SELECT statement from definition
func (m *Manager) buildSelectSQL(def ViewDefinition) (string, error) {
	sql := "SELECT "

	// Columns
	if len(def.Columns) == 0 {
		sql += "*"
	} else {
		for i, col := range def.Columns {
			if i > 0 {
				sql += ", "
			}
			if col.Expression != "" {
				sql += col.Expression
			} else if col.SourceColumn != "" {
				sql += col.SourceColumn
			} else {
				sql += col.Name
			}
			if col.Alias != "" {
				sql += " AS " + col.Alias
			} else if col.Name != "" && col.Expression != "" {
				sql += " AS " + col.Name
			}
		}
	}

	// FROM
	sql += " FROM " + def.BaseTable

	// JOINs
	for _, join := range def.Joins {
		sql += fmt.Sprintf(" %s JOIN %s", join.Type, join.Table)
		if join.OnCondition != "" {
			sql += " ON " + join.OnCondition
		} else if len(join.OnColumns) >= 2 {
			sql += " ON "
			for i := 0; i < len(join.OnColumns); i += 2 {
				if i > 0 {
					sql += " AND "
				}
				sql += fmt.Sprintf("%s = %s", join.OnColumns[i], join.OnColumns[i+1])
			}
		}
	}

	// WHERE
	if len(def.Filters) > 0 {
		sql += " WHERE "
		for i, filter := range def.Filters {
			if i > 0 {
				logicalOp := filter.LogicalOp
				if logicalOp == "" {
					logicalOp = "AND"
				}
				sql += fmt.Sprintf(" %s ", logicalOp)
			}
			sql += buildFilterCondition(filter)
		}
	}

	// GROUP BY
	if len(def.GroupBy) > 0 {
		sql += " GROUP BY "
		for i, col := range def.GroupBy {
			if i > 0 {
				sql += ", "
			}
			sql += col
		}
	}

	// ORDER BY
	if len(def.OrderBy) > 0 {
		sql += " ORDER BY "
		for i, order := range def.OrderBy {
			if i > 0 {
				sql += ", "
			}
			sql += order.Column
			if order.Direction != "" {
				sql += " " + order.Direction
			}
		}
	}

	// LIMIT
	if def.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", def.Limit)
	}

	return sql, nil
}

// buildUnionSQL builds a UNION query
func (m *Manager) buildUnionSQL(def ViewDefinition) (string, error) {
	if len(def.UnionTables) == 0 {
		return "", fmt.Errorf("no tables specified for union")
	}

	unionType := "UNION"
	if def.UnionAll {
		unionType = "UNION ALL"
	}

	sql := ""
	for i, table := range def.UnionTables {
		if i > 0 {
			sql += fmt.Sprintf(" %s ", unionType)
		}
		sql += fmt.Sprintf("SELECT * FROM %s", table)
	}

	return sql, nil
}

// buildFilterCondition builds a SQL condition from a filter
func buildFilterCondition(filter FilterDef) string {
	switch filter.Operator {
	case "eq":
		return fmt.Sprintf("%s = %v", filter.Column, formatValue(filter.Value))
	case "ne":
		return fmt.Sprintf("%s <> %v", filter.Column, formatValue(filter.Value))
	case "lt":
		return fmt.Sprintf("%s < %v", filter.Column, formatValue(filter.Value))
	case "lte":
		return fmt.Sprintf("%s <= %v", filter.Column, formatValue(filter.Value))
	case "gt":
		return fmt.Sprintf("%s > %v", filter.Column, formatValue(filter.Value))
	case "gte":
		return fmt.Sprintf("%s >= %v", filter.Column, formatValue(filter.Value))
	case "in":
		return fmt.Sprintf("%s IN (%s)", filter.Column, formatValues(filter.Values))
	case "not_in":
		return fmt.Sprintf("%s NOT IN (%s)", filter.Column, formatValues(filter.Values))
	case "like":
		return fmt.Sprintf("%s LIKE %v", filter.Column, formatValue(filter.Value))
	case "is_null":
		return fmt.Sprintf("%s IS NULL", filter.Column)
	case "is_not_null":
		return fmt.Sprintf("%s IS NOT NULL", filter.Column)
	default:
		return fmt.Sprintf("%s = %v", filter.Column, formatValue(filter.Value))
	}
}

// formatValue formats a value for SQL
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatValues formats multiple values for SQL IN clause
func formatValues(values []interface{}) string {
	result := ""
	for i, v := range values {
		if i > 0 {
			result += ", "
		}
		result += formatValue(v)
	}
	return result
}

// ViewConnector wraps a view to implement the Connector interface
// This allows running checks against views using the same check infrastructure
type ViewConnector struct {
	view    *View
	manager *Manager
}

// NewViewConnector creates a connector for a view
func NewViewConnector(view *View, manager *Manager) *ViewConnector {
	return &ViewConnector{
		view:    view,
		manager: manager,
	}
}

// Connect is a no-op for views
func (c *ViewConnector) Connect(ctx context.Context) error {
	return nil
}

// Close is a no-op for views
func (c *ViewConnector) Close() error {
	return nil
}

// Ping validates the view
func (c *ViewConnector) Ping(ctx context.Context) error {
	return c.manager.ValidateView(ctx, c.view.ID)
}

// Query executes a query against the view
func (c *ViewConnector) Query(ctx context.Context, query string, args ...interface{}) (*datasource.QueryResult, error) {
	// Replace table references with the view subquery
	viewSQL, err := c.manager.GetViewSQL(ctx, c.view.ID)
	if err != nil {
		return nil, err
	}

	// Wrap the view SQL as a subquery
	// This is a simplified approach - actual implementation would need SQL parsing
	wrappedQuery := fmt.Sprintf("WITH _view AS (%s) %s", viewSQL, query)

	connector, err := c.manager.datasourceManager.GetConnector(ctx, c.view.DatasourceID)
	if err != nil {
		return nil, err
	}

	return connector.Query(ctx, wrappedQuery, args...)
}

// GetTables returns the view as a single "table"
func (c *ViewConnector) GetTables(ctx context.Context) ([]datasource.TableInfo, error) {
	return []datasource.TableInfo{
		{
			Name: c.view.Name,
			Type: "view",
		},
	}, nil
}

// GetColumns returns the view schema
func (c *ViewConnector) GetColumns(ctx context.Context, table string) ([]datasource.ColumnInfo, error) {
	return c.view.Schema, nil
}

// GetRowCount returns the view row count
func (c *ViewConnector) GetRowCount(ctx context.Context, table string) (int64, error) {
	return c.manager.GetViewRowCount(ctx, c.view.ID)
}

// Type returns the datasource type
func (c *ViewConnector) Type() datasource.Type {
	return "view"
}
