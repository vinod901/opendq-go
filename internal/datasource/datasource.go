// Package datasource provides abstractions for connecting to various data sources
// including databases, data warehouses, and file-based storage systems.
package datasource

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Type represents the type of datasource
type Type string

const (
	// Database types
	TypePostgres   Type = "postgres"
	TypeMySQL      Type = "mysql"
	TypeSQLServer  Type = "sqlserver"
	TypeOracle     Type = "oracle"
	TypeSnowflake  Type = "snowflake"
	TypeDatabricks Type = "databricks"
	TypeBigQuery   Type = "bigquery"
	TypeTrino      Type = "trino"
	TypeDuckDB     Type = "duckdb"
	TypeClickHouse Type = "clickhouse"
	// Lakehouse types
	TypeHDFS      Type = "hdfs"
	TypeDeltaLake Type = "deltalake"
	TypeIceberg   Type = "iceberg"
	TypeHudi      Type = "hudi"
	// File types (for file observability)
	TypeS3           Type = "s3"
	TypeGCS          Type = "gcs"
	TypeAzureBlob    Type = "azure_blob"
	TypeLocalStorage Type = "local"
	// Virtual types
	TypeView Type = "view"
)

// Datasource represents a data source configuration
type Datasource struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        Type                   `json:"type"`
	Connection  ConnectionConfig       `json:"connection"`
	Metadata    map[string]interface{} `json:"metadata"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ConnectionConfig holds connection configuration for a datasource
type ConnectionConfig struct {
	// Common database fields
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	SSLMode  string `json:"ssl_mode,omitempty"`

	// Cloud-specific fields
	Account       string `json:"account,omitempty"`       // Snowflake, Databricks
	Warehouse     string `json:"warehouse,omitempty"`     // Snowflake
	Schema        string `json:"schema,omitempty"`        // Database schema
	Catalog       string `json:"catalog,omitempty"`       // Trino, Databricks
	ProjectID     string `json:"project_id,omitempty"`    // BigQuery
	Dataset       string `json:"dataset,omitempty"`       // BigQuery
	HTTPPath      string `json:"http_path,omitempty"`     // Databricks
	Token         string `json:"token,omitempty"`         // Auth token
	PrivateKey    string `json:"private_key,omitempty"`   // Key-based auth
	KeyFile       string `json:"key_file,omitempty"`      // Service account key file
	ConnectionURL string `json:"connection_url,omitempty"` // Direct connection URL

	// Storage-specific fields
	Bucket    string `json:"bucket,omitempty"`
	Region    string `json:"region,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"` // Custom endpoint (MinIO, etc.)

	// Additional options
	Options map[string]string `json:"options,omitempty"`
}

// Connector interface defines the contract for connecting to data sources
type Connector interface {
	// Connect establishes a connection to the datasource
	Connect(ctx context.Context) error

	// Close closes the connection
	Close() error

	// Ping checks if the connection is alive
	Ping(ctx context.Context) error

	// Query executes a query and returns results
	Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error)

	// GetTables returns a list of tables/datasets in the datasource
	GetTables(ctx context.Context) ([]TableInfo, error)

	// GetColumns returns column information for a table
	GetColumns(ctx context.Context, table string) ([]ColumnInfo, error)

	// GetRowCount returns the row count for a table
	GetRowCount(ctx context.Context, table string) (int64, error)

	// Type returns the datasource type
	Type() Type
}

// QueryResult holds the result of a query
type QueryResult struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	RowCount int64                   `json:"row_count"`
}

// TableInfo contains information about a table
type TableInfo struct {
	Schema    string `json:"schema"`
	Name      string `json:"name"`
	Type      string `json:"type"` // table, view, materialized_view
	RowCount  int64  `json:"row_count,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

// ColumnInfo contains information about a column
type ColumnInfo struct {
	Name         string `json:"name"`
	DataType     string `json:"data_type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value,omitempty"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	Description  string `json:"description,omitempty"`
}

// Manager handles datasource operations
type Manager struct {
	datasources map[string]*Datasource
	connectors  map[string]Connector
}

// NewManager creates a new datasource manager
func NewManager() *Manager {
	return &Manager{
		datasources: make(map[string]*Datasource),
		connectors:  make(map[string]Connector),
	}
}

// CreateDatasource creates a new datasource
func (m *Manager) CreateDatasource(ctx context.Context, ds *Datasource) error {
	if ds.ID == "" {
		ds.ID = uuid.New().String()
	}
	ds.CreatedAt = time.Now()
	ds.UpdatedAt = time.Now()
	ds.Active = true

	// Validate connection before storing
	connector, err := m.createConnector(ds)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}

	if err := connector.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to datasource: %w", err)
	}

	if err := connector.Ping(ctx); err != nil {
		connector.Close()
		return fmt.Errorf("failed to ping datasource: %w", err)
	}

	m.datasources[ds.ID] = ds
	m.connectors[ds.ID] = connector
	return nil
}

// GetDatasource retrieves a datasource by ID
func (m *Manager) GetDatasource(ctx context.Context, id string) (*Datasource, error) {
	ds, exists := m.datasources[id]
	if !exists {
		return nil, fmt.Errorf("datasource not found: %s", id)
	}
	return ds, nil
}

// UpdateDatasource updates a datasource
func (m *Manager) UpdateDatasource(ctx context.Context, id string, updates map[string]interface{}) error {
	ds, exists := m.datasources[id]
	if !exists {
		return fmt.Errorf("datasource not found: %s", id)
	}

	if name, ok := updates["name"].(string); ok {
		ds.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		ds.Description = description
	}
	if metadata, ok := updates["metadata"].(map[string]interface{}); ok {
		ds.Metadata = metadata
	}
	if active, ok := updates["active"].(bool); ok {
		ds.Active = active
	}

	ds.UpdatedAt = time.Now()
	return nil
}

// DeleteDatasource deletes a datasource
func (m *Manager) DeleteDatasource(ctx context.Context, id string) error {
	if _, exists := m.datasources[id]; !exists {
		return fmt.Errorf("datasource not found: %s", id)
	}

	if connector, exists := m.connectors[id]; exists {
		connector.Close()
		delete(m.connectors, id)
	}

	delete(m.datasources, id)
	return nil
}

// ListDatasources lists datasources for a tenant
func (m *Manager) ListDatasources(ctx context.Context, tenantID string) ([]*Datasource, error) {
	var result []*Datasource
	for _, ds := range m.datasources {
		if tenantID == "" || ds.TenantID == tenantID {
			result = append(result, ds)
		}
	}
	return result, nil
}

// GetConnector returns the connector for a datasource
func (m *Manager) GetConnector(ctx context.Context, id string) (Connector, error) {
	connector, exists := m.connectors[id]
	if !exists {
		return nil, fmt.Errorf("connector not found for datasource: %s", id)
	}
	return connector, nil
}

// TestConnection tests a datasource connection without storing it
func (m *Manager) TestConnection(ctx context.Context, ds *Datasource) error {
	connector, err := m.createConnector(ds)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}
	defer connector.Close()

	if err := connector.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	return connector.Ping(ctx)
}

// createConnector creates the appropriate connector based on datasource type
func (m *Manager) createConnector(ds *Datasource) (Connector, error) {
	switch ds.Type {
	case TypePostgres:
		return NewPostgresConnector(ds.Connection), nil
	case TypeMySQL:
		return NewMySQLConnector(ds.Connection), nil
	case TypeSQLServer:
		return NewSQLServerConnector(ds.Connection), nil
	case TypeOracle:
		return NewOracleConnector(ds.Connection), nil
	case TypeSnowflake:
		return NewSnowflakeConnector(ds.Connection), nil
	case TypeDatabricks:
		return NewDatabricksConnector(ds.Connection), nil
	case TypeBigQuery:
		return NewBigQueryConnector(ds.Connection), nil
	case TypeTrino:
		return NewTrinoConnector(ds.Connection), nil
	case TypeDuckDB:
		return NewDuckDBConnector(ds.Connection), nil
	case TypeClickHouse:
		return NewClickHouseConnector(ds.Connection), nil
	case TypeHDFS, TypeDeltaLake, TypeIceberg, TypeHudi:
		return NewLakehouseConnector(ds.Type, ds.Connection), nil
	case TypeS3, TypeGCS, TypeAzureBlob, TypeLocalStorage:
		return NewStorageConnector(ds.Type, ds.Connection), nil
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", ds.Type)
	}
}

// BaseConnector provides common functionality for SQL-based connectors
type BaseConnector struct {
	config ConnectionConfig
	db     *sql.DB
	dsType Type
}

// Connect establishes a connection
func (c *BaseConnector) Connect(ctx context.Context) error {
	// Base implementation - overridden by specific connectors
	return nil
}

// Close closes the connection
func (c *BaseConnector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping checks the connection
func (c *BaseConnector) Ping(ctx context.Context) error {
	if c.db != nil {
		return c.db.PingContext(ctx)
	}
	return fmt.Errorf("database connection not established")
}

// Query executes a query
func (c *BaseConnector) Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	result := &QueryResult{
		Columns: columns,
		Rows:    make([]map[string]interface{}, 0),
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		result.Rows = append(result.Rows, row)
		result.RowCount++
	}

	return result, nil
}

// Type returns the datasource type
func (c *BaseConnector) Type() Type {
	return c.dsType
}
