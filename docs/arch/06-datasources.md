# Datasources

OpenDQ supports connecting to a wide variety of data sources including traditional databases, cloud data warehouses, lakehouse platforms, and cloud storage systems.

## Overview

```
┌────────────────────────────────────────────────────────────────────────────┐
│                        Datasource Architecture                              │
│                                                                            │
│  ┌─────────────────────────────────────────────────────────────────────┐  │
│  │                      Datasource Manager                              │  │
│  │  - CreateDatasource()    - GetConnector()                           │  │
│  │  - GetDatasource()       - TestConnection()                         │  │
│  │  - UpdateDatasource()    - ListDatasources()                        │  │
│  │  - DeleteDatasource()                                               │  │
│  └────────────────────────────────────┬────────────────────────────────┘  │
│                                       │                                    │
│                          ┌────────────▼───────────────┐                   │
│                          │    Connector Interface     │                   │
│                          │  - Connect()               │                   │
│                          │  - Close()                 │                   │
│                          │  - Ping()                  │                   │
│                          │  - Query()                 │                   │
│                          │  - GetTables()             │                   │
│                          │  - GetColumns()            │                   │
│                          │  - GetRowCount()           │                   │
│                          └────────────┬───────────────┘                   │
│                                       │                                    │
│          ┌────────────────────────────┼────────────────────────────┐      │
│          │                            │                            │      │
│  ┌───────▼───────┐  ┌────────────────▼───────────────┐  ┌─────────▼────┐ │
│  │ SQL Connectors│  │   Lakehouse Connectors         │  │   Storage    │ │
│  │ - Postgres    │  │   - HDFS                       │  │ - S3         │ │
│  │ - MySQL       │  │   - Delta Lake                 │  │ - GCS        │ │
│  │ - SQL Server  │  │   - Iceberg                    │  │ - Azure Blob │ │
│  │ - Oracle      │  │   - Hudi                       │  │ - Local      │ │
│  │ - Snowflake   │  │                                │  │              │ │
│  │ - Databricks  │  │                                │  │              │ │
│  │ - BigQuery    │  │                                │  │              │ │
│  │ - Trino       │  │                                │  │              │ │
│  │ - DuckDB      │  │                                │  │              │ │
│  │ - ClickHouse  │  │                                │  │              │ │
│  └───────────────┘  └────────────────────────────────┘  └──────────────┘ │
└────────────────────────────────────────────────────────────────────────────┘
```

## Supported Datasource Types

### Traditional Databases

| Type | Constant | Driver | Description |
|------|----------|--------|-------------|
| PostgreSQL | `postgres` | `lib/pq` | Open-source relational database |
| MySQL | `mysql` | `mysql` | Popular open-source database |
| SQL Server | `sqlserver` | `sqlserver` | Microsoft SQL Server |
| Oracle | `oracle` | `godror` | Oracle Database |

### Cloud Data Warehouses

| Type | Constant | Description |
|------|----------|-------------|
| Snowflake | `snowflake` | Cloud data warehouse |
| Databricks | `databricks` | Lakehouse platform |
| BigQuery | `bigquery` | Google Cloud DWH |
| Trino | `trino` | Distributed SQL engine |

### Analytics Databases

| Type | Constant | Description |
|------|----------|-------------|
| DuckDB | `duckdb` | Embedded analytics |
| ClickHouse | `clickhouse` | Column-oriented OLAP |

### Lakehouse Formats

| Type | Constant | Description |
|------|----------|-------------|
| HDFS | `hdfs` | Hadoop Distributed File System |
| Delta Lake | `deltalake` | Databricks open table format |
| Iceberg | `iceberg` | Apache Iceberg tables |
| Hudi | `hudi` | Apache Hudi tables |

### Cloud Storage

| Type | Constant | Description |
|------|----------|-------------|
| S3 | `s3` | Amazon S3 |
| GCS | `gcs` | Google Cloud Storage |
| Azure Blob | `azure_blob` | Azure Blob Storage |
| Local | `local` | Local filesystem |

## Datasource Model

```go
// internal/datasource/datasource.go

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
    KeyFile       string `json:"key_file,omitempty"`      // Service account key
    ConnectionURL string `json:"connection_url,omitempty"` // Direct URL

    // Storage-specific fields
    Bucket    string `json:"bucket,omitempty"`
    Region    string `json:"region,omitempty"`
    AccessKey string `json:"access_key,omitempty"`
    SecretKey string `json:"secret_key,omitempty"`
    Endpoint  string `json:"endpoint,omitempty"` // Custom endpoint

    // Additional options
    Options map[string]string `json:"options,omitempty"`
}
```

## Connector Interface

All connectors implement the same interface:

```go
type Connector interface {
    // Connection management
    Connect(ctx context.Context) error
    Close() error
    Ping(ctx context.Context) error
    
    // Query operations
    Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error)
    
    // Metadata operations
    GetTables(ctx context.Context) ([]TableInfo, error)
    GetColumns(ctx context.Context, table string) ([]ColumnInfo, error)
    GetRowCount(ctx context.Context, table string) (int64, error)
    
    // Type identification
    Type() Type
}

type QueryResult struct {
    Columns  []string                 `json:"columns"`
    Rows     []map[string]interface{} `json:"rows"`
    RowCount int64                    `json:"row_count"`
}

type TableInfo struct {
    Schema    string `json:"schema"`
    Name      string `json:"name"`
    Type      string `json:"type"` // table, view, materialized_view
    RowCount  int64  `json:"row_count,omitempty"`
    SizeBytes int64  `json:"size_bytes,omitempty"`
}

type ColumnInfo struct {
    Name         string `json:"name"`
    DataType     string `json:"data_type"`
    Nullable     bool   `json:"nullable"`
    DefaultValue string `json:"default_value,omitempty"`
    IsPrimaryKey bool   `json:"is_primary_key"`
    Description  string `json:"description,omitempty"`
}
```

## Datasource Manager

```go
type Manager struct {
    datasources map[string]*Datasource
    connectors  map[string]Connector
}

func NewManager() *Manager {
    return &Manager{
        datasources: make(map[string]*Datasource),
        connectors:  make(map[string]Connector),
    }
}

// CreateDatasource creates and validates a new datasource
func (m *Manager) CreateDatasource(ctx context.Context, ds *Datasource) error {
    // Generate ID if not provided
    if ds.ID == "" {
        ds.ID = uuid.New().String()
    }
    
    // Create connector
    connector, err := m.createConnector(ds)
    if err != nil {
        return fmt.Errorf("failed to create connector: %w", err)
    }
    
    // Validate connection
    if err := connector.Connect(ctx); err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    
    if err := connector.Ping(ctx); err != nil {
        connector.Close()
        return fmt.Errorf("failed to ping: %w", err)
    }
    
    // Store datasource and connector
    m.datasources[ds.ID] = ds
    m.connectors[ds.ID] = connector
    return nil
}

// createConnector creates the appropriate connector based on type
func (m *Manager) createConnector(ds *Datasource) (Connector, error) {
    switch ds.Type {
    case TypePostgres:
        return NewPostgresConnector(ds.Connection), nil
    case TypeMySQL:
        return NewMySQLConnector(ds.Connection), nil
    case TypeSnowflake:
        return NewSnowflakeConnector(ds.Connection), nil
    case TypeBigQuery:
        return NewBigQueryConnector(ds.Connection), nil
    case TypeS3:
        return NewStorageConnector(ds.Type, ds.Connection), nil
    // ... other types
    default:
        return nil, fmt.Errorf("unsupported type: %s", ds.Type)
    }
}
```

## Connection Examples

### PostgreSQL

```json
{
    "name": "Production PostgreSQL",
    "type": "postgres",
    "connection": {
        "host": "db.example.com",
        "port": 5432,
        "database": "production",
        "username": "reader",
        "password": "secret",
        "ssl_mode": "require"
    }
}
```

```go
// PostgresConnector
func (c *PostgresConnector) Connect(ctx context.Context) error {
    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.config.Host, c.config.Port, c.config.Username,
        c.config.Password, c.config.Database, c.config.SSLMode,
    )
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return err
    }
    c.db = db
    return nil
}
```

### Snowflake

```json
{
    "name": "Snowflake Warehouse",
    "type": "snowflake",
    "connection": {
        "account": "xy12345.us-east-1",
        "username": "OPENDQ_USER",
        "password": "secret",
        "database": "PRODUCTION",
        "warehouse": "COMPUTE_WH",
        "schema": "PUBLIC"
    }
}
```

### BigQuery

```json
{
    "name": "BigQuery Dataset",
    "type": "bigquery",
    "connection": {
        "project_id": "my-project",
        "dataset": "analytics",
        "key_file": "/path/to/service-account.json"
    }
}
```

### Databricks

```json
{
    "name": "Databricks SQL",
    "type": "databricks",
    "connection": {
        "host": "adb-12345.azuredatabricks.net",
        "http_path": "/sql/1.0/warehouses/abcd1234",
        "token": "dapi123...",
        "catalog": "main",
        "schema": "default"
    }
}
```

### S3 Storage

```json
{
    "name": "S3 Data Lake",
    "type": "s3",
    "connection": {
        "bucket": "my-data-lake",
        "region": "us-east-1",
        "access_key": "AKIAIOSFODNN7EXAMPLE",
        "secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    }
}
```

## API Operations

### Create Datasource

```bash
POST /api/v1/datasources
{
    "name": "Production DB",
    "type": "postgres",
    "connection": { ... }
}
```

### Test Connection

```bash
POST /api/v1/datasources/test
{
    "type": "postgres",
    "connection": { ... }
}

Response:
{
    "success": true,
    "message": "Connection successful"
}
```

### List Tables

```bash
GET /api/v1/datasources/{id}/tables

Response:
[
    {
        "schema": "public",
        "name": "users",
        "type": "table",
        "row_count": 50000
    },
    {
        "schema": "public",
        "name": "orders",
        "type": "table",
        "row_count": 1000000
    }
]
```

## BaseConnector

Common functionality is shared via BaseConnector:

```go
type BaseConnector struct {
    config ConnectionConfig
    db     *sql.DB
    dsType Type
}

func (c *BaseConnector) Close() error {
    if c.db != nil {
        return c.db.Close()
    }
    return nil
}

func (c *BaseConnector) Ping(ctx context.Context) error {
    if c.db != nil {
        return c.db.PingContext(ctx)
    }
    return fmt.Errorf("not connected")
}

func (c *BaseConnector) Query(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
    rows, err := c.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    columns, err := rows.Columns()
    if err != nil {
        return nil, err
    }
    
    result := &QueryResult{
        Columns: columns,
        Rows:    make([]map[string]interface{}, 0),
    }
    
    for rows.Next() {
        // Scan row into map
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range values {
            valuePtrs[i] = &values[i]
        }
        
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, err
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
```

## Security Considerations

1. **Credential Storage**: Store credentials encrypted at rest
2. **Connection Pooling**: Reuse connections efficiently
3. **Timeout Handling**: Set appropriate connection timeouts
4. **Error Handling**: Don't expose internal connection errors
5. **Read-Only Access**: Recommend read-only database users
6. **Network Security**: Use SSL/TLS for connections
7. **Secret Rotation**: Support for rotating credentials
