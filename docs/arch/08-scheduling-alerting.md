# Scheduling & Alerting

OpenDQ provides scheduling capabilities to run checks automatically and alerting to notify teams when checks fail.

## Overview

```
┌────────────────────────────────────────────────────────────────────────────┐
│                     Scheduling & Alerting Architecture                      │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                        Schedule Manager                               │ │
│  │                                                                       │ │
│  │  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐         │ │
│  │  │   Schedule   │────▶│  Cron Parser │────▶│  Ticker      │         │ │
│  │  │   Config     │     │              │     │  (Time-based)│         │ │
│  │  └──────────────┘     └──────────────┘     └──────────────┘         │ │
│  │                                                     │                 │ │
│  │                                                     ▼                 │ │
│  │                                            ┌──────────────┐          │ │
│  │                                            │ Check Runner │          │ │
│  │                                            │              │          │ │
│  │                                            └──────┬───────┘          │ │
│  └───────────────────────────────────────────────────┼──────────────────┘ │
│                                                      │                     │
│                                                      ▼                     │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │                         Alert Manager                                 │ │
│  │                                                                       │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐ │ │
│  │  │  Email   │  │  Slack   │  │ Webhook  │  │PagerDuty │  │ Teams  │ │ │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └────────┘ │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────────┘
```

## Schedules

### Schedule Model

```go
// internal/scheduler/scheduler.go

type Schedule struct {
    ID              string                 `json:"id"`
    TenantID        string                 `json:"tenant_id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    CronExpression  string                 `json:"cron_expression"`
    Timezone        string                 `json:"timezone"`
    CheckIDs        []string               `json:"check_ids"`
    AlertChannelIDs []string               `json:"alert_channel_ids"`
    Enabled         bool                   `json:"enabled"`
    Metadata        map[string]interface{} `json:"metadata"`
    CreatedAt       time.Time              `json:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at"`
    LastRunAt       *time.Time             `json:"last_run_at,omitempty"`
    NextRunAt       *time.Time             `json:"next_run_at,omitempty"`
}

type ScheduleExecution struct {
    ID           string                 `json:"id"`
    ScheduleID   string                 `json:"schedule_id"`
    Status       string                 `json:"status"` // running, completed, failed
    StartedAt    time.Time              `json:"started_at"`
    CompletedAt  *time.Time             `json:"completed_at,omitempty"`
    CheckResults []*check.CheckResult   `json:"check_results"`
    AlertsSent   int                    `json:"alerts_sent"`
    Error        string                 `json:"error,omitempty"`
    Metadata     map[string]interface{} `json:"metadata"`
}
```

### Cron Expression Examples

| Expression | Description |
|------------|-------------|
| `0 * * * *` | Every hour at minute 0 |
| `0 0 * * *` | Daily at midnight |
| `0 0 * * 0` | Weekly on Sunday at midnight |
| `0 8 * * 1-5` | Weekdays at 8 AM |
| `*/15 * * * *` | Every 15 minutes |
| `0 9,18 * * *` | At 9 AM and 6 PM |

### Schedule Manager

```go
type Manager struct {
    schedules    map[string]*Schedule
    executions   map[string][]*ScheduleExecution
    checkManager *check.Manager
    alertManager *alerting.Manager
    cron         *cron.Cron
}

func NewManager(checkManager *check.Manager, alertManager *alerting.Manager) *Manager {
    return &Manager{
        schedules:    make(map[string]*Schedule),
        executions:   make(map[string][]*ScheduleExecution),
        checkManager: checkManager,
        alertManager: alertManager,
        cron:         cron.New(cron.WithSeconds()),
    }
}

func (m *Manager) Start() {
    m.cron.Start()
}

func (m *Manager) Stop() {
    m.cron.Stop()
}

func (m *Manager) CreateSchedule(ctx context.Context, schedule *Schedule) error {
    if schedule.ID == "" {
        schedule.ID = uuid.New().String()
    }
    
    // Validate cron expression
    parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
    _, err := parser.Parse(schedule.CronExpression)
    if err != nil {
        return fmt.Errorf("invalid cron expression: %w", err)
    }
    
    // Calculate next run time
    schedule.NextRunAt = m.calculateNextRun(schedule)
    
    // Add to cron scheduler
    entryID, err := m.cron.AddFunc(schedule.CronExpression, func() {
        m.executeSchedule(context.Background(), schedule.ID)
    })
    if err != nil {
        return err
    }
    
    schedule.Metadata["cron_entry_id"] = entryID
    m.schedules[schedule.ID] = schedule
    
    return nil
}

func (m *Manager) executeSchedule(ctx context.Context, scheduleID string) error {
    schedule, err := m.GetSchedule(ctx, scheduleID)
    if err != nil {
        return err
    }
    
    if !schedule.Enabled {
        return nil
    }
    
    execution := &ScheduleExecution{
        ID:         uuid.New().String(),
        ScheduleID: scheduleID,
        Status:     "running",
        StartedAt:  time.Now(),
    }
    
    // Run all checks
    var results []*check.CheckResult
    var failedChecks []string
    
    for _, checkID := range schedule.CheckIDs {
        result, err := m.checkManager.RunCheck(ctx, checkID)
        if err != nil {
            execution.Error = err.Error()
            continue
        }
        results = append(results, result)
        
        if result.Status == check.StatusFailed {
            failedChecks = append(failedChecks, checkID)
        }
    }
    
    execution.CheckResults = results
    
    // Send alerts for failures
    if len(failedChecks) > 0 {
        for _, channelID := range schedule.AlertChannelIDs {
            err := m.alertManager.SendAlert(ctx, channelID, AlertPayload{
                Schedule:     schedule,
                FailedChecks: failedChecks,
                Execution:    execution,
            })
            if err == nil {
                execution.AlertsSent++
            }
        }
    }
    
    // Complete execution
    now := time.Now()
    execution.CompletedAt = &now
    execution.Status = "completed"
    if execution.Error != "" {
        execution.Status = "failed"
    }
    
    // Store execution
    m.executions[scheduleID] = append(m.executions[scheduleID], execution)
    
    // Update schedule
    schedule.LastRunAt = &now
    schedule.NextRunAt = m.calculateNextRun(schedule)
    
    return nil
}
```

## Alert Channels

### Channel Types

| Type | Description | Configuration |
|------|-------------|---------------|
| `email` | Email notifications | SMTP settings, recipients |
| `slack` | Slack webhooks | Webhook URL, channel |
| `webhook` | Generic HTTP webhooks | URL, headers, auth |
| `pagerduty` | PagerDuty incidents | API key, service ID |
| `msteams` | Microsoft Teams | Webhook URL |
| `opsgenie` | OpsGenie alerts | API key, recipients |

### Alert Channel Model

```go
// internal/alerting/alerting.go

type ChannelType string

const (
    ChannelTypeEmail     ChannelType = "email"
    ChannelTypeSlack     ChannelType = "slack"
    ChannelTypeWebhook   ChannelType = "webhook"
    ChannelTypePagerDuty ChannelType = "pagerduty"
    ChannelTypeMSTeams   ChannelType = "msteams"
    ChannelTypeOpsGenie  ChannelType = "opsgenie"
)

type Channel struct {
    ID          string                 `json:"id"`
    TenantID    string                 `json:"tenant_id"`
    Name        string                 `json:"name"`
    Type        ChannelType            `json:"type"`
    Config      ChannelConfig          `json:"config"`
    Enabled     bool                   `json:"enabled"`
    Metadata    map[string]interface{} `json:"metadata"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

type ChannelConfig struct {
    // Email
    SMTPHost     string   `json:"smtp_host,omitempty"`
    SMTPPort     int      `json:"smtp_port,omitempty"`
    SMTPUser     string   `json:"smtp_user,omitempty"`
    SMTPPassword string   `json:"smtp_password,omitempty"`
    FromEmail    string   `json:"from_email,omitempty"`
    ToEmails     []string `json:"to_emails,omitempty"`
    
    // Slack/Teams/Webhook
    WebhookURL string            `json:"webhook_url,omitempty"`
    Channel    string            `json:"channel,omitempty"`
    Headers    map[string]string `json:"headers,omitempty"`
    
    // PagerDuty
    APIKey    string `json:"api_key,omitempty"`
    ServiceID string `json:"service_id,omitempty"`
    
    // OpsGenie
    APIKey    string   `json:"api_key,omitempty"`
    Teams     []string `json:"teams,omitempty"`
    Priority  string   `json:"priority,omitempty"`
}

type AlertHistory struct {
    ID          string                 `json:"id"`
    ChannelID   string                 `json:"channel_id"`
    CheckID     string                 `json:"check_id,omitempty"`
    ScheduleID  string                 `json:"schedule_id,omitempty"`
    Status      string                 `json:"status"` // sent, failed
    Message     string                 `json:"message"`
    Error       string                 `json:"error,omitempty"`
    SentAt      time.Time              `json:"sent_at"`
    Metadata    map[string]interface{} `json:"metadata"`
}
```

### Alert Manager

```go
type Manager struct {
    channels map[string]*Channel
    history  []*AlertHistory
}

func (m *Manager) SendAlert(ctx context.Context, channelID string, payload AlertPayload) error {
    channel, err := m.GetChannel(ctx, channelID)
    if err != nil {
        return err
    }
    
    if !channel.Enabled {
        return nil
    }
    
    // Send based on channel type
    var sendErr error
    switch channel.Type {
    case ChannelTypeEmail:
        sendErr = m.sendEmailAlert(ctx, channel, payload)
    case ChannelTypeSlack:
        sendErr = m.sendSlackAlert(ctx, channel, payload)
    case ChannelTypeWebhook:
        sendErr = m.sendWebhookAlert(ctx, channel, payload)
    case ChannelTypePagerDuty:
        sendErr = m.sendPagerDutyAlert(ctx, channel, payload)
    case ChannelTypeMSTeams:
        sendErr = m.sendMSTeamsAlert(ctx, channel, payload)
    case ChannelTypeOpsGenie:
        sendErr = m.sendOpsGenieAlert(ctx, channel, payload)
    default:
        sendErr = fmt.Errorf("unsupported channel type: %s", channel.Type)
    }
    
    // Record history
    history := &AlertHistory{
        ID:        uuid.New().String(),
        ChannelID: channelID,
        SentAt:    time.Now(),
    }
    if sendErr != nil {
        history.Status = "failed"
        history.Error = sendErr.Error()
    } else {
        history.Status = "sent"
    }
    m.history = append(m.history, history)
    
    return sendErr
}

func (m *Manager) sendSlackAlert(ctx context.Context, channel *Channel, payload AlertPayload) error {
    message := map[string]interface{}{
        "text": fmt.Sprintf("🚨 Data Quality Alert: %s", payload.Schedule.Name),
        "attachments": []map[string]interface{}{
            {
                "color": "danger",
                "fields": []map[string]string{
                    {"title": "Schedule", "value": payload.Schedule.Name, "short": true},
                    {"title": "Failed Checks", "value": strings.Join(payload.FailedChecks, ", "), "short": true},
                },
            },
        },
    }
    
    body, _ := json.Marshal(message)
    resp, err := http.Post(channel.Config.WebhookURL, "application/json", bytes.NewReader(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return fmt.Errorf("slack returned status %d", resp.StatusCode)
    }
    
    return nil
}
```

## API Examples

### Create Schedule

```bash
curl -X POST http://localhost:8080/api/v1/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Nightly Data Quality Checks",
    "cron_expression": "0 0 * * *",
    "timezone": "UTC",
    "check_ids": ["check-123", "check-456", "check-789"],
    "alert_channel_ids": ["channel-slack", "channel-email"],
    "enabled": true
  }'
```

### Run Schedule Now

```bash
curl -X POST http://localhost:8080/api/v1/schedules/schedule-123/run
```

### Get Schedule Executions

```bash
curl http://localhost:8080/api/v1/schedules/schedule-123/executions

Response:
[
    {
        "id": "exec-456",
        "schedule_id": "schedule-123",
        "status": "completed",
        "started_at": "2024-01-15T00:00:00Z",
        "completed_at": "2024-01-15T00:05:23Z",
        "check_results": [...],
        "alerts_sent": 2
    }
]
```

### Create Alert Channel (Slack)

```bash
curl -X POST http://localhost:8080/api/v1/alerts/channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Data Team Slack",
    "type": "slack",
    "config": {
      "webhook_url": "https://hooks.slack.com/services/T00/B00/xxx",
      "channel": "#data-alerts"
    },
    "enabled": true
  }'
```

### Create Alert Channel (Email)

```bash
curl -X POST http://localhost:8080/api/v1/alerts/channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Data Team Email",
    "type": "email",
    "config": {
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "smtp_user": "alerts@company.com",
      "smtp_password": "xxx",
      "from_email": "alerts@company.com",
      "to_emails": ["data-team@company.com"]
    },
    "enabled": true
  }'
```

### Create Alert Channel (PagerDuty)

```bash
curl -X POST http://localhost:8080/api/v1/alerts/channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "On-Call Alerts",
    "type": "pagerduty",
    "config": {
      "api_key": "xxx",
      "service_id": "PXXXXXX"
    },
    "enabled": true
  }'
```

### Test Alert Channel

```bash
curl -X POST http://localhost:8080/api/v1/alerts/channels/channel-123/test

Response:
{
    "success": true,
    "message": "Test alert sent successfully"
}
```

### Get Alert History

```bash
curl http://localhost:8080/api/v1/alerts/history?channel_id=channel-123

Response:
[
    {
        "id": "history-456",
        "channel_id": "channel-123",
        "status": "sent",
        "sent_at": "2024-01-15T00:05:25Z"
    },
    {
        "id": "history-455",
        "channel_id": "channel-123",
        "status": "failed",
        "error": "webhook returned 500",
        "sent_at": "2024-01-14T00:05:25Z"
    }
]
```

## Alert Message Templates

### Slack Message

```json
{
    "text": "🚨 Data Quality Alert",
    "attachments": [
        {
            "color": "danger",
            "title": "Schedule: Nightly Checks",
            "fields": [
                {"title": "Status", "value": "Failed", "short": true},
                {"title": "Time", "value": "2024-01-15 00:00 UTC", "short": true},
                {"title": "Failed Checks", "value": "3 of 10"},
                {"title": "Details", "value": "• users_row_count: Row count 800 below minimum 1000\n• orders_freshness: Data is 48h old"}
            ],
            "actions": [
                {"type": "button", "text": "View Details", "url": "https://opendq.example.com/schedules/123"}
            ]
        }
    ]
}
```

### Email Template

```html
Subject: 🚨 Data Quality Alert: Nightly Checks Failed

<h2>Data Quality Alert</h2>
<p><strong>Schedule:</strong> Nightly Checks</p>
<p><strong>Status:</strong> <span style="color:red">Failed</span></p>
<p><strong>Time:</strong> 2024-01-15 00:00 UTC</p>

<h3>Failed Checks (3 of 10)</h3>
<table>
  <tr><th>Check</th><th>Status</th><th>Message</th></tr>
  <tr><td>users_row_count</td><td>Failed</td><td>Row count 800 below minimum 1000</td></tr>
  <tr><td>orders_freshness</td><td>Failed</td><td>Data is 48h old</td></tr>
</table>

<p><a href="https://opendq.example.com/schedules/123">View Details</a></p>
```

## Best Practices

### Scheduling

1. **Stagger Schedules**: Avoid running all checks at once
2. **Off-Peak Hours**: Schedule during low-traffic periods
3. **Timeouts**: Set appropriate execution timeouts
4. **Retries**: Implement retry logic for transient failures
5. **Dependencies**: Order checks based on dependencies

### Alerting

1. **Severity Routing**: Route critical alerts to PagerDuty, others to Slack
2. **Alert Fatigue**: Don't over-alert; use appropriate thresholds
3. **Grouping**: Group related alerts together
4. **Escalation**: Set up escalation paths for unacked alerts
5. **Testing**: Regularly test alert channels
