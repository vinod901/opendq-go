# Features

Detailed documentation of OpenDQ features and capabilities.

## Core Features

### Multi-Datasource Connectivity

OpenDQ supports connecting to a wide variety of data sources:

#### Traditional Databases
- **PostgreSQL**: Full support including schemas, views
- **MySQL**: Including MariaDB compatibility
- **SQL Server**: Azure SQL and on-premises
- **Oracle**: Including Oracle Cloud

#### Cloud Data Warehouses
- **Snowflake**: Native driver support with warehouse selection
- **Databricks**: Unity Catalog and workspace support
- **BigQuery**: Service account authentication
- **Trino**: Distributed query engine

#### Analytics Databases
- **DuckDB**: Embedded analytics
- **ClickHouse**: High-performance OLAP

#### Lakehouse Formats
- **Delta Lake**: Open table format
- **Apache Iceberg**: Versioned tables
- **Apache Hudi**: Incremental processing

#### Cloud Storage
- **Amazon S3**: Including MinIO compatibility
- **Google Cloud Storage**: With service account
- **Azure Blob Storage**: SAS token or connection string

---

### Data Quality Checks

#### Row Count Checks
Validate that tables contain expected number of rows.

**Configuration:**
```json
{
  "type": "row_count",
  "table": "users",
  "parameters": {
    "min_rows": 1000,
    "max_rows": 100000
  }
}
```

**Use Cases:**
- ETL job validation
- Data pipeline monitoring
- Source-to-target reconciliation

#### Null Checks
Ensure columns have expected completeness.

**Configuration:**
```json
{
  "type": "null_check",
  "table": "users",
  "column": "email",
  "parameters": {
    "max_null_percentage": 0,
    "max_null_count": 0
  }
}
```

**Use Cases:**
- Required field validation
- Data completeness monitoring
- PII field verification

#### Uniqueness Checks
Validate unique constraints on columns.

**Configuration:**
```json
{
  "type": "uniqueness",
  "table": "users",
  "parameters": {
    "unique_columns": ["email"],
    "allow_nulls": false
  }
}
```

**Use Cases:**
- Primary key validation
- Natural key verification
- Deduplication monitoring

#### Freshness Checks
Monitor data recency.

**Configuration:**
```json
{
  "type": "freshness",
  "table": "orders",
  "parameters": {
    "timestamp_column": "created_at",
    "max_age_hours": 24
  }
}
```

**Use Cases:**
- Pipeline latency monitoring
- SLA compliance
- Real-time data validation

#### Custom SQL Checks
Execute any SQL for custom validation.

**Configuration:**
```json
{
  "type": "custom_sql",
  "parameters": {
    "custom_sql": "SELECT COUNT(*) FROM orders WHERE total < 0",
    "expected_value": "0"
  }
}
```

**Use Cases:**
- Business rule validation
- Complex data quality rules
- Cross-table validation

#### Value Checks
Validate statistical properties.

**Types:**
- `min_value`: Minimum value validation
- `max_value`: Maximum value validation
- `mean_value`: Average validation
- `sum_value`: Total validation
- `std_dev`: Standard deviation check

**Configuration:**
```json
{
  "type": "min_value",
  "table": "products",
  "column": "price",
  "parameters": {
    "expected_min": 0
  }
}
```

#### Pattern Checks
Validate data format using regex.

**Configuration:**
```json
{
  "type": "regex",
  "table": "users",
  "column": "email",
  "parameters": {
    "pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
  }
}
```

#### Set Membership
Validate values are in allowed set.

**Configuration:**
```json
{
  "type": "set_membership",
  "table": "orders",
  "column": "status",
  "parameters": {
    "allowed_values": ["pending", "processing", "shipped", "delivered", "cancelled"]
  }
}
```

#### Referential Integrity
Validate foreign key relationships.

**Configuration:**
```json
{
  "type": "referential_integrity",
  "table": "orders",
  "column": "customer_id",
  "parameters": {
    "reference_table": "customers",
    "reference_column": "id"
  }
}
```

#### Schema Checks
Validate table structure.

**Types:**
- `schema_match`: Full schema comparison
- `column_count`: Expected column count
- `column_type`: Specific column type validation

---

### Scheduling

#### Cron-Based Scheduling
Define schedules using cron expressions:

| Expression | Description |
|------------|-------------|
| `0 * * * *` | Every hour |
| `0 0 * * *` | Daily at midnight |
| `0 0 * * 0` | Weekly on Sunday |
| `0 8 * * 1-5` | Weekdays at 8 AM |
| `*/15 * * * *` | Every 15 minutes |

#### Timezone Support
Schedules run in specified timezone:
- UTC (recommended)
- America/New_York
- Europe/London
- Asia/Tokyo

#### Schedule Grouping
Group related checks into single schedule for:
- Atomic execution
- Consolidated alerting
- Execution dependency

---

### Alerting

#### Alert Routing
Route alerts based on severity:

| Severity | Recommended Channel |
|----------|---------------------|
| Critical | PagerDuty, OpsGenie |
| High | Slack + Email |
| Medium | Slack |
| Low | Email digest |

#### Alert Content
Alerts include:
- Check name and description
- Actual vs expected values
- Failure message
- Link to details
- Execution timestamp

#### Alert Templates
Customize alert messages:
- Slack: Rich formatting with attachments
- Email: HTML templates
- Webhook: Custom JSON payload

---

### Multi-Tenancy

#### Tenant Isolation
- Complete data separation
- Independent configurations
- Separate user management
- Individual billing (optional)

#### Tenant Roles
| Role | Capabilities |
|------|--------------|
| Owner | Full control, billing |
| Admin | User management, settings |
| Editor | Resource CRUD |
| Viewer | Read-only access |

#### Tenant Onboarding
1. Create tenant
2. Configure settings
3. Add datasources
4. Create checks
5. Set up schedules
6. Configure alerts
7. Invite team members

---

### Logical Views

#### View Types
- **Aggregation Views**: Summarize data
- **Filter Views**: Subset of data
- **Join Views**: Combine tables
- **Transform Views**: Computed columns

#### View SQL Examples

**Aggregation:**
```sql
SELECT 
  date_trunc('day', created_at) as date,
  COUNT(*) as order_count,
  SUM(total) as revenue
FROM orders
GROUP BY 1
```

**Filter:**
```sql
SELECT * FROM customers
WHERE country = 'US'
AND status = 'active'
```

**Join:**
```sql
SELECT 
  o.id,
  o.total,
  c.name as customer_name,
  c.email
FROM orders o
JOIN customers c ON o.customer_id = c.id
```

---

### Data Lineage

OpenDQ integrates with OpenLineage for data lineage tracking.

#### Lineage Events
- Job start/complete events
- Input/output dataset tracking
- Run metadata

#### Visualization
View lineage in Marquez Web UI:
- Dataset lineage graph
- Job dependencies
- Run history

---

## Enterprise Features

### High Availability
- Stateless API servers
- Database replication
- Redis clustering

### Security
- OIDC authentication
- Fine-grained authorization
- TLS encryption
- Audit logging

### Scalability
- Horizontal scaling
- Connection pooling
- Async processing

### Observability
- Prometheus metrics
- OpenTelemetry tracing
- Structured logging
