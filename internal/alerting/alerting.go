// Package alerting provides alerting capabilities for data quality check failures.
// Supports multiple alert channels including email, Slack, webhooks, and PagerDuty.
package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ChannelType represents the type of alert channel
type ChannelType string

const (
	ChannelTypeEmail     ChannelType = "email"
	ChannelTypeSlack     ChannelType = "slack"
	ChannelTypeWebhook   ChannelType = "webhook"
	ChannelTypePagerDuty ChannelType = "pagerduty"
	ChannelTypeMSTeams   ChannelType = "msteams"
	ChannelTypeOpsGenie  ChannelType = "opsgenie"
)

// Severity represents alert severity
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Channel represents an alert channel configuration
type Channel struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Type            ChannelType            `json:"type"`
	Configuration   ChannelConfig          `json:"configuration"`
	Active          bool                   `json:"active"`
	MinSeverity     Severity               `json:"min_severity"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ChannelConfig holds channel-specific configuration
type ChannelConfig struct {
	// Email configuration
	EmailAddresses []string `json:"email_addresses,omitempty"`
	SMTPHost       string   `json:"smtp_host,omitempty"`
	SMTPPort       int      `json:"smtp_port,omitempty"`
	SMTPUser       string   `json:"smtp_user,omitempty"`
	SMTPPassword   string   `json:"smtp_password,omitempty"`
	FromAddress    string   `json:"from_address,omitempty"`

	// Slack configuration
	SlackWebhookURL string `json:"slack_webhook_url,omitempty"`
	SlackChannel    string `json:"slack_channel,omitempty"`

	// Webhook configuration
	WebhookURL     string            `json:"webhook_url,omitempty"`
	WebhookMethod  string            `json:"webhook_method,omitempty"`
	WebhookHeaders map[string]string `json:"webhook_headers,omitempty"`

	// PagerDuty configuration
	PagerDutyRoutingKey string `json:"pagerduty_routing_key,omitempty"`

	// MS Teams configuration
	TeamsWebhookURL string `json:"teams_webhook_url,omitempty"`

	// OpsGenie configuration
	OpsGenieAPIKey string `json:"opsgenie_api_key,omitempty"`
}

// Alert represents an alert to be sent
type Alert struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Severity    Severity               `json:"severity"`
	ScheduleID  string                 `json:"schedule_id,omitempty"`
	ExecutionID string                 `json:"execution_id,omitempty"`
	CheckID     string                 `json:"check_id,omitempty"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AlertHistory represents a sent alert record
type AlertHistory struct {
	ID         string    `json:"id"`
	AlertID    string    `json:"alert_id"`
	ChannelID  string    `json:"channel_id"`
	Status     string    `json:"status"`
	SentAt     time.Time `json:"sent_at"`
	Error      string    `json:"error,omitempty"`
}

// Manager handles alerting operations
type Manager struct {
	channels   map[string]*Channel
	history    []*AlertHistory
	httpClient *http.Client
}

// NewManager creates a new alerting manager
func NewManager() *Manager {
	return &Manager{
		channels: make(map[string]*Channel),
		history:  make([]*AlertHistory, 0),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateChannel creates a new alert channel
func (m *Manager) CreateChannel(ctx context.Context, channel *Channel) error {
	if channel.ID == "" {
		channel.ID = uuid.New().String()
	}
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()
	channel.Active = true

	m.channels[channel.ID] = channel
	return nil
}

// GetChannel retrieves a channel by ID
func (m *Manager) GetChannel(ctx context.Context, id string) (*Channel, error) {
	channel, exists := m.channels[id]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", id)
	}
	return channel, nil
}

// UpdateChannel updates a channel
func (m *Manager) UpdateChannel(ctx context.Context, id string, updates map[string]interface{}) error {
	channel, exists := m.channels[id]
	if !exists {
		return fmt.Errorf("channel not found: %s", id)
	}

	if name, ok := updates["name"].(string); ok {
		channel.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		channel.Description = description
	}
	if active, ok := updates["active"].(bool); ok {
		channel.Active = active
	}
	if minSeverity, ok := updates["min_severity"].(Severity); ok {
		channel.MinSeverity = minSeverity
	}
	if config, ok := updates["configuration"].(ChannelConfig); ok {
		channel.Configuration = config
	}

	channel.UpdatedAt = time.Now()
	return nil
}

// DeleteChannel deletes a channel
func (m *Manager) DeleteChannel(ctx context.Context, id string) error {
	if _, exists := m.channels[id]; !exists {
		return fmt.Errorf("channel not found: %s", id)
	}
	delete(m.channels, id)
	return nil
}

// ListChannels lists channels with optional filters
func (m *Manager) ListChannels(ctx context.Context, tenantID string) ([]*Channel, error) {
	var result []*Channel
	for _, channel := range m.channels {
		if tenantID == "" || channel.TenantID == tenantID {
			result = append(result, channel)
		}
	}
	return result, nil
}

// SendAlert sends an alert to a channel
func (m *Manager) SendAlert(ctx context.Context, channelID string, alert *Alert) error {
	channel, err := m.GetChannel(ctx, channelID)
	if err != nil {
		return err
	}

	if !channel.Active {
		return fmt.Errorf("channel is inactive")
	}

	// Check severity threshold
	if !severityMeetsThreshold(alert.Severity, channel.MinSeverity) {
		return nil // Skip sending if below threshold
	}

	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}

	var sendErr error
	switch channel.Type {
	case ChannelTypeEmail:
		sendErr = m.sendEmail(ctx, channel, alert)
	case ChannelTypeSlack:
		sendErr = m.sendSlack(ctx, channel, alert)
	case ChannelTypeWebhook:
		sendErr = m.sendWebhook(ctx, channel, alert)
	case ChannelTypePagerDuty:
		sendErr = m.sendPagerDuty(ctx, channel, alert)
	case ChannelTypeMSTeams:
		sendErr = m.sendMSTeams(ctx, channel, alert)
	case ChannelTypeOpsGenie:
		sendErr = m.sendOpsGenie(ctx, channel, alert)
	default:
		sendErr = fmt.Errorf("unsupported channel type: %s", channel.Type)
	}

	// Record history
	history := &AlertHistory{
		ID:        uuid.New().String(),
		AlertID:   alert.ID,
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

// SendAlertToAll sends an alert to all active channels for a tenant
func (m *Manager) SendAlertToAll(ctx context.Context, tenantID string, alert *Alert) error {
	channels, err := m.ListChannels(ctx, tenantID)
	if err != nil {
		return err
	}

	var lastErr error
	for _, channel := range channels {
		if channel.Active {
			if err := m.SendAlert(ctx, channel.ID, alert); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

// GetAlertHistory returns alert history
func (m *Manager) GetAlertHistory(ctx context.Context, channelID string, limit int) ([]*AlertHistory, error) {
	var result []*AlertHistory
	for _, h := range m.history {
		if channelID == "" || h.ChannelID == channelID {
			result = append(result, h)
		}
	}

	if limit > 0 && len(result) > limit {
		return result[len(result)-limit:], nil
	}
	return result, nil
}

// TestChannel tests a channel by sending a test alert
func (m *Manager) TestChannel(ctx context.Context, channelID string) error {
	testAlert := &Alert{
		Title:    "Test Alert",
		Message:  "This is a test alert from OpenDQ",
		Severity: SeverityInfo,
		Details: map[string]interface{}{
			"test": true,
		},
	}
	return m.SendAlert(ctx, channelID, testAlert)
}

// Channel-specific send implementations

func (m *Manager) sendEmail(ctx context.Context, channel *Channel, alert *Alert) error {
	// In production: use net/smtp or gomail
	// For now, return success for demonstration
	config := channel.Configuration
	if len(config.EmailAddresses) == 0 {
		return fmt.Errorf("no email addresses configured")
	}
	// Actual implementation would send email via SMTP
	return nil
}

func (m *Manager) sendSlack(ctx context.Context, channel *Channel, alert *Alert) error {
	config := channel.Configuration
	if config.SlackWebhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	// Build Slack message
	payload := map[string]interface{}{
		"text": fmt.Sprintf("*%s*\n%s", alert.Title, alert.Message),
		"attachments": []map[string]interface{}{
			{
				"color":  getSeverityColor(alert.Severity),
				"fields": buildSlackFields(alert),
			},
		},
	}

	return m.postJSON(ctx, config.SlackWebhookURL, payload)
}

func (m *Manager) sendWebhook(ctx context.Context, channel *Channel, alert *Alert) error {
	config := channel.Configuration
	if config.WebhookURL == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	method := config.WebhookMethod
	if method == "" {
		method = "POST"
	}

	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range config.WebhookHeaders {
		req.Header.Set(key, value)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func (m *Manager) sendPagerDuty(ctx context.Context, channel *Channel, alert *Alert) error {
	config := channel.Configuration
	if config.PagerDutyRoutingKey == "" {
		return fmt.Errorf("PagerDuty routing key not configured")
	}

	payload := map[string]interface{}{
		"routing_key":  config.PagerDutyRoutingKey,
		"event_action": "trigger",
		"dedup_key":    alert.ID,
		"payload": map[string]interface{}{
			"summary":   alert.Title,
			"severity":  mapSeverityToPagerDuty(alert.Severity),
			"source":    "opendq",
			"timestamp": alert.Timestamp.Format(time.RFC3339),
			"custom_details": map[string]interface{}{
				"message": alert.Message,
				"details": alert.Details,
			},
		},
	}

	return m.postJSON(ctx, "https://events.pagerduty.com/v2/enqueue", payload)
}

func (m *Manager) sendMSTeams(ctx context.Context, channel *Channel, alert *Alert) error {
	config := channel.Configuration
	if config.TeamsWebhookURL == "" {
		return fmt.Errorf("MS Teams webhook URL not configured")
	}

	// Build MS Teams adaptive card
	payload := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": getSeverityColor(alert.Severity),
		"summary":    alert.Title,
		"sections": []map[string]interface{}{
			{
				"activityTitle": alert.Title,
				"facts": []map[string]string{
					{"name": "Severity", "value": string(alert.Severity)},
					{"name": "Message", "value": alert.Message},
				},
			},
		},
	}

	return m.postJSON(ctx, config.TeamsWebhookURL, payload)
}

func (m *Manager) sendOpsGenie(ctx context.Context, channel *Channel, alert *Alert) error {
	config := channel.Configuration
	if config.OpsGenieAPIKey == "" {
		return fmt.Errorf("OpsGenie API key not configured")
	}

	payload := map[string]interface{}{
		"message":     alert.Title,
		"description": alert.Message,
		"priority":    mapSeverityToOpsGenie(alert.Severity),
		"details":     alert.Details,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.opsgenie.com/v2/alerts", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+config.OpsGenieAPIKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to OpsGenie: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("OpsGenie returned status %d", resp.StatusCode)
	}

	return nil
}

// Helper functions

func (m *Manager) postJSON(ctx context.Context, url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request returned status %d", resp.StatusCode)
	}

	return nil
}

func severityMeetsThreshold(alertSeverity, channelMinSeverity Severity) bool {
	severityOrder := map[Severity]int{
		SeverityCritical: 5,
		SeverityHigh:     4,
		SeverityMedium:   3,
		SeverityLow:      2,
		SeverityInfo:     1,
	}

	alertLevel := severityOrder[alertSeverity]
	minLevel := severityOrder[channelMinSeverity]
	if minLevel == 0 {
		minLevel = 1 // Default to info
	}

	return alertLevel >= minLevel
}

func getSeverityColor(severity Severity) string {
	colors := map[Severity]string{
		SeverityCritical: "#FF0000",
		SeverityHigh:     "#FF6600",
		SeverityMedium:   "#FFCC00",
		SeverityLow:      "#0066FF",
		SeverityInfo:     "#00CC00",
	}
	if color, ok := colors[severity]; ok {
		return color
	}
	return "#808080"
}

func mapSeverityToPagerDuty(severity Severity) string {
	mapping := map[Severity]string{
		SeverityCritical: "critical",
		SeverityHigh:     "error",
		SeverityMedium:   "warning",
		SeverityLow:      "info",
		SeverityInfo:     "info",
	}
	if pd, ok := mapping[severity]; ok {
		return pd
	}
	return "info"
}

func mapSeverityToOpsGenie(severity Severity) string {
	mapping := map[Severity]string{
		SeverityCritical: "P1",
		SeverityHigh:     "P2",
		SeverityMedium:   "P3",
		SeverityLow:      "P4",
		SeverityInfo:     "P5",
	}
	if og, ok := mapping[severity]; ok {
		return og
	}
	return "P5"
}

func buildSlackFields(alert *Alert) []map[string]interface{} {
	fields := []map[string]interface{}{
		{"title": "Severity", "value": string(alert.Severity), "short": true},
	}

	if alert.ScheduleID != "" {
		fields = append(fields, map[string]interface{}{
			"title": "Schedule ID", "value": alert.ScheduleID, "short": true,
		})
	}

	for key, value := range alert.Details {
		fields = append(fields, map[string]interface{}{
			"title": key, "value": fmt.Sprintf("%v", value), "short": true,
		})
	}

	return fields
}
