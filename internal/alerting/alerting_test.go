package alerting

import (
	"context"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.channels == nil {
		t.Fatal("channels map is nil")
	}
	if m.history == nil {
		t.Fatal("history slice is nil")
	}
}

func TestChannelType_Values(t *testing.T) {
	types := []ChannelType{
		ChannelTypeEmail,
		ChannelTypeSlack,
		ChannelTypeWebhook,
		ChannelTypePagerDuty,
		ChannelTypeMSTeams,
		ChannelTypeOpsGenie,
	}

	for _, channelType := range types {
		if channelType == "" {
			t.Error("channel type should not be empty")
		}
	}
}

func TestSeverity_Values(t *testing.T) {
	severities := []Severity{
		SeverityCritical,
		SeverityHigh,
		SeverityMedium,
		SeverityLow,
		SeverityInfo,
	}

	for _, severity := range severities {
		if severity == "" {
			t.Error("severity should not be empty")
		}
	}
}

func TestManager_CreateChannel(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	channel := &Channel{
		TenantID: "tenant-1",
		Name:     "Slack Channel",
		Type:     ChannelTypeSlack,
		Configuration: ChannelConfig{
			SlackWebhookURL: "https://hooks.slack.com/services/xxx",
		},
	}

	err := m.CreateChannel(ctx, channel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if channel.ID == "" {
		t.Error("channel ID should be generated")
	}
	if !channel.Active {
		t.Error("channel should be active by default")
	}
}

func TestManager_GetChannel(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	channel := &Channel{
		TenantID: "tenant-1",
		Name:     "Slack Channel",
		Type:     ChannelTypeSlack,
	}
	m.CreateChannel(ctx, channel)

	retrieved, err := m.GetChannel(ctx, channel.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.Name != channel.Name {
		t.Errorf("expected name %s, got %s", channel.Name, retrieved.Name)
	}
}

func TestManager_GetChannel_NotFound(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	_, err := m.GetChannel(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent channel")
	}
}

func TestManager_DeleteChannel(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	channel := &Channel{
		TenantID: "tenant-1",
		Name:     "Slack Channel",
		Type:     ChannelTypeSlack,
	}
	m.CreateChannel(ctx, channel)

	err := m.DeleteChannel(ctx, channel.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.GetChannel(ctx, channel.ID)
	if err == nil {
		t.Fatal("expected error for deleted channel")
	}
}

func TestManager_ListChannels(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	m.CreateChannel(ctx, &Channel{
		TenantID: "tenant-1",
		Name:     "Channel 1",
		Type:     ChannelTypeSlack,
	})
	m.CreateChannel(ctx, &Channel{
		TenantID: "tenant-2",
		Name:     "Channel 2",
		Type:     ChannelTypeEmail,
	})

	// List all
	channels, err := m.ListChannels(ctx, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(channels))
	}

	// Filter by tenant
	channels, err = m.ListChannels(ctx, "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 1 {
		t.Errorf("expected 1 channel for tenant-1, got %d", len(channels))
	}
}

func TestSeverityMeetsThreshold(t *testing.T) {
	testCases := []struct {
		alertSeverity   Severity
		minSeverity     Severity
		expectedResult  bool
	}{
		{SeverityCritical, SeverityCritical, true},
		{SeverityCritical, SeverityHigh, true},
		{SeverityCritical, SeverityInfo, true},
		{SeverityHigh, SeverityCritical, false},
		{SeverityMedium, SeverityHigh, false},
		{SeverityInfo, SeverityCritical, false},
		{SeverityInfo, SeverityInfo, true},
	}

	for _, tc := range testCases {
		t.Run(string(tc.alertSeverity)+"_"+string(tc.minSeverity), func(t *testing.T) {
			result := severityMeetsThreshold(tc.alertSeverity, tc.minSeverity)
			if result != tc.expectedResult {
				t.Errorf("severityMeetsThreshold(%s, %s) = %v, want %v",
					tc.alertSeverity, tc.minSeverity, result, tc.expectedResult)
			}
		})
	}
}

func TestGetSeverityColor(t *testing.T) {
	testCases := []struct {
		severity Severity
	}{
		{SeverityCritical},
		{SeverityHigh},
		{SeverityMedium},
		{SeverityLow},
		{SeverityInfo},
	}

	for _, tc := range testCases {
		t.Run(string(tc.severity), func(t *testing.T) {
			color := getSeverityColor(tc.severity)
			if color == "" {
				t.Errorf("getSeverityColor(%s) returned empty string", tc.severity)
			}
			if color[0] != '#' {
				t.Errorf("getSeverityColor(%s) = %s, expected hex color", tc.severity, color)
			}
		})
	}
}

func TestMapSeverityToPagerDuty(t *testing.T) {
	testCases := []struct {
		severity Severity
		expected string
	}{
		{SeverityCritical, "critical"},
		{SeverityHigh, "error"},
		{SeverityMedium, "warning"},
		{SeverityLow, "info"},
		{SeverityInfo, "info"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.severity), func(t *testing.T) {
			result := mapSeverityToPagerDuty(tc.severity)
			if result != tc.expected {
				t.Errorf("mapSeverityToPagerDuty(%s) = %s, want %s", tc.severity, result, tc.expected)
			}
		})
	}
}

func TestMapSeverityToOpsGenie(t *testing.T) {
	testCases := []struct {
		severity Severity
		expected string
	}{
		{SeverityCritical, "P1"},
		{SeverityHigh, "P2"},
		{SeverityMedium, "P3"},
		{SeverityLow, "P4"},
		{SeverityInfo, "P5"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.severity), func(t *testing.T) {
			result := mapSeverityToOpsGenie(tc.severity)
			if result != tc.expected {
				t.Errorf("mapSeverityToOpsGenie(%s) = %s, want %s", tc.severity, result, tc.expected)
			}
		})
	}
}
