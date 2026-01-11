# User Guide

This guide covers all the features and functionality available in OpenDQ.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Datasources](#datasources)
3. [Data Quality Checks](#data-quality-checks)
4. [Schedules](#schedules)
5. [Alert Channels](#alert-channels)
6. [Logical Views](#logical-views)
7. [Tenants](#tenants)

---

## Getting Started

### Login

1. Navigate to your OpenDQ instance URL
2. You'll be redirected to the login page (Keycloak or your OIDC provider)
3. Enter your credentials
4. After successful login, you'll be redirected to the OpenDQ dashboard

### Signup

New users can register through your organization's identity provider:

1. Navigate to the login page
2. Click "Register" (if enabled by your administrator)
3. Fill in your details
4. Verify your email if required
5. Login with your new credentials

### Dashboard Overview

The dashboard provides:
- Summary of datasources
- Recent check executions
- Failed checks requiring attention
- Scheduled jobs status

---

## Datasources

Datasources are connections to your databases, data warehouses, or storage systems.

### Adding a Datasource

1. Navigate to **Datasources** → **Add Datasource**
2. Select the datasource type:
   - **Databases**: PostgreSQL, MySQL, SQL Server, Oracle
   - **Data Warehouses**: Snowflake, Databricks, BigQuery, Trino
   - **Analytics**: DuckDB, ClickHouse
   - **Lakehouse**: Delta Lake, Iceberg, Hudi
   - **Storage**: S3, GCS, Azure Blob
3. Fill in connection details:
   - Name (descriptive identifier)
   - Host/Account
   - Port (if applicable)
   - Database/Schema
   - Credentials
4. Click **Test Connection** to verify
5. Click **Save** to create the datasource

### Connection Examples

#### PostgreSQL
```
Host: db.example.com
Port: 5432
Database: production
Username: reader
Password: ****
SSL Mode: require
```

#### Snowflake
```
Account: xy12345.us-east-1
Warehouse: COMPUTE_WH
Database: PRODUCTION
Schema: PUBLIC
Username: OPENDQ_USER
Password: ****
```

#### BigQuery
```
Project ID: my-project
Dataset: analytics
Key File: [Upload service account JSON]
```

### Managing Datasources

- **View**: Click on datasource name to see details and tables
- **Edit**: Update connection details or description
- **Delete**: Remove datasource (will also remove associated checks)
- **Test**: Re-verify connection is working

---

## Data Quality Checks

Checks validate your data against defined rules.

### Check Types

| Type | Description | Example |
|------|-------------|---------|
| Row Count | Validate table row count | Ensure ETL loaded data |
| Null Check | Check for null values | Data completeness |
| Uniqueness | Validate unique values | Primary key validity |
| Freshness | Check data recency | Pipeline latency |
| Custom SQL | Custom SQL validation | Business rules |
| Value Range | Value within range | Age between 0-150 |
| Pattern Match | Regex pattern matching | Email format |
| Referential | Foreign key validation | Data integrity |

### Creating a Check

1. Navigate to **Checks** → **Create Check**
2. Select the datasource
3. Choose check type
4. Configure parameters:

#### Row Count Check
```
Table: users
Minimum Rows: 1000
Maximum Rows: (optional)
```

#### Null Check
```
Table: users
Column: email
Max Null Percentage: 0%
```

#### Freshness Check
```
Table: orders
Timestamp Column: created_at
Max Age: 24 hours
```

#### Custom SQL Check
```
SQL: SELECT COUNT(*) FROM users WHERE status = 'active'
Expected Value: >= 1000
```

5. Set severity (Critical, High, Medium, Low)
6. Add tags for organization
7. Click **Save**

### Running Checks

- **Manual Run**: Click **Run Now** on any check
- **Scheduled Run**: Associate check with a schedule (see Schedules)
- **Bulk Run**: Run all checks for a datasource

### Viewing Results

Check results show:
- Status (Passed, Failed, Warning, Error)
- Actual value vs. expected
- Execution duration
- Timestamp
- Error details (if any)

---

## Schedules

Schedules automate check execution on a recurring basis.

### Creating a Schedule

1. Navigate to **Schedules** → **Create Schedule**
2. Enter schedule name
3. Configure timing using cron expression or visual picker:
   - `0 * * * *` - Every hour
   - `0 0 * * *` - Daily at midnight
   - `0 8 * * 1-5` - Weekdays at 8 AM
4. Select timezone
5. Add checks to run
6. Select alert channels (optional)
7. Click **Save**

### Schedule Management

- **Enable/Disable**: Toggle schedule on/off
- **Run Now**: Execute schedule immediately
- **View Executions**: See history of runs
- **Edit**: Modify schedule configuration

### Execution History

Each execution shows:
- Start/end time
- Status (completed, failed)
- Check results summary
- Alerts sent count

---

## Alert Channels

Get notified when checks fail.

### Supported Channels

| Channel | Description |
|---------|-------------|
| Email | Send email notifications |
| Slack | Post to Slack channels |
| Webhook | Generic HTTP webhooks |
| PagerDuty | Create incidents |
| Microsoft Teams | Post to Teams channels |
| OpsGenie | Create alerts |

### Creating an Alert Channel

#### Slack
1. Navigate to **Alerts** → **Add Channel**
2. Select **Slack**
3. Enter:
   - Name: "Data Team Slack"
   - Webhook URL: https://hooks.slack.com/services/...
   - Channel: #data-alerts
4. Click **Test** to send a test message
5. Click **Save**

#### Email
1. Select **Email**
2. Enter:
   - Name: "Data Team Email"
   - SMTP Host: smtp.gmail.com
   - SMTP Port: 587
   - SMTP User: alerts@company.com
   - Recipients: team@company.com
3. Click **Save**

#### PagerDuty
1. Select **PagerDuty**
2. Enter:
   - Name: "On-Call Alerts"
   - API Key: xxx
   - Service ID: PXXXXXX
3. Click **Save**

### Alert History

View all alerts sent:
- Channel used
- Status (sent/failed)
- Timestamp
- Associated check/schedule

---

## Logical Views

Create virtual datasets for running checks on transformed data.

### Creating a View

1. Navigate to **Views** → **Create View**
2. Select base datasource
3. Enter view name
4. Write SQL query:
   ```sql
   SELECT 
     customer_id,
     COUNT(*) as order_count,
     SUM(total) as total_revenue
   FROM orders
   WHERE created_at > CURRENT_DATE - INTERVAL '30 days'
   GROUP BY customer_id
   ```
5. Click **Validate** to check SQL
6. Click **Save**

### Using Views

Views can be:
- Used as target for checks (instead of physical tables)
- Previewed with **Query** button
- Used to check derived/calculated data

---

## Tenants

If multi-tenancy is enabled, you can manage separate workspaces.

### Switching Tenants

1. Click tenant selector in header
2. Choose from available tenants
3. All resources will be scoped to selected tenant

### Tenant Roles

| Role | Permissions |
|------|-------------|
| Owner | Full control, can delete tenant |
| Admin | Manage users and settings |
| Editor | Create/modify resources |
| Viewer | Read-only access |

### Inviting Users

(Admin/Owner only)
1. Navigate to **Settings** → **Users**
2. Click **Invite User**
3. Enter email
4. Select role
5. Send invitation

---

## Best Practices

### Datasource Setup
- Use read-only database users
- Enable SSL/TLS connections
- Store credentials securely

### Check Design
- Start with critical data first
- Set appropriate thresholds
- Use descriptive names and tags
- Document business context

### Scheduling
- Stagger schedules to avoid load spikes
- Run during low-traffic periods
- Set appropriate timeouts

### Alerting
- Route by severity (PagerDuty for critical, Slack for others)
- Avoid alert fatigue
- Test channels regularly
