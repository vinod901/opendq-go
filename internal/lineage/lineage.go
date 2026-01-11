package lineage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client manages OpenLineage event publishing
type Client struct {
	endpoint  string
	namespace string
	httpClient *http.Client
}

// Config contains OpenLineage configuration
type Config struct {
	Endpoint  string
	Namespace string
}

// NewClient creates a new OpenLineage client
func NewClient(cfg Config) *Client {
	return &Client{
		endpoint:  cfg.Endpoint,
		namespace: cfg.Namespace,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Event represents an OpenLineage event
type Event struct {
	EventType  string    `json:"eventType"`
	EventTime  time.Time `json:"eventTime"`
	Run        Run       `json:"run"`
	Job        Job       `json:"job"`
	Inputs     []Dataset `json:"inputs,omitempty"`
	Outputs    []Dataset `json:"outputs,omitempty"`
	Producer   string    `json:"producer"`
	SchemaURL  string    `json:"schemaURL"`
}

// Run represents a run in OpenLineage
type Run struct {
	RunID  string                 `json:"runId"`
	Facets map[string]interface{} `json:"facets,omitempty"`
}

// Job represents a job in OpenLineage
type Job struct {
	Namespace string                 `json:"namespace"`
	Name      string                 `json:"name"`
	Facets    map[string]interface{} `json:"facets,omitempty"`
}

// Dataset represents a dataset in OpenLineage
type Dataset struct {
	Namespace string                 `json:"namespace"`
	Name      string                 `json:"name"`
	Facets    map[string]interface{} `json:"facets,omitempty"`
}

// EventType constants
const (
	EventTypeStart    = "START"
	EventTypeRunning  = "RUNNING"
	EventTypeComplete = "COMPLETE"
	EventTypeFail     = "FAIL"
	EventTypeAbort    = "ABORT"
)

// EmitEvent publishes an OpenLineage event
func (c *Client) EmitEvent(ctx context.Context, event Event) error {
	// Set default values
	if event.Producer == "" {
		event.Producer = "opendq-go"
	}
	if event.SchemaURL == "" {
		event.SchemaURL = "https://openlineage.io/spec/2-0-2/OpenLineage.json"
	}
	if event.Job.Namespace == "" {
		event.Job.Namespace = c.namespace
	}

	// Serialize event
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Send to OpenLineage endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/v1/lineage", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// EmitStartEvent emits a START event
func (c *Client) EmitStartEvent(ctx context.Context, runID, jobName string, inputs, outputs []Dataset) error {
	event := Event{
		EventType: EventTypeStart,
		EventTime: time.Now().UTC(),
		Run: Run{
			RunID: runID,
			Facets: map[string]interface{}{},
		},
		Job: Job{
			Namespace: c.namespace,
			Name:      jobName,
			Facets:    map[string]interface{}{},
		},
		Inputs:  inputs,
		Outputs: outputs,
	}

	return c.EmitEvent(ctx, event)
}

// EmitCompleteEvent emits a COMPLETE event
func (c *Client) EmitCompleteEvent(ctx context.Context, runID, jobName string, inputs, outputs []Dataset) error {
	event := Event{
		EventType: EventTypeComplete,
		EventTime: time.Now().UTC(),
		Run: Run{
			RunID: runID,
			Facets: map[string]interface{}{},
		},
		Job: Job{
			Namespace: c.namespace,
			Name:      jobName,
			Facets:    map[string]interface{}{},
		},
		Inputs:  inputs,
		Outputs: outputs,
	}

	return c.EmitEvent(ctx, event)
}

// EmitFailEvent emits a FAIL event
func (c *Client) EmitFailEvent(ctx context.Context, runID, jobName string, err error) error {
	event := Event{
		EventType: EventTypeFail,
		EventTime: time.Now().UTC(),
		Run: Run{
			RunID: runID,
			Facets: map[string]interface{}{
				"errorMessage": map[string]interface{}{
					"_producer": "opendq-go",
					"_schemaURL": "https://openlineage.io/spec/facets/1-0-0/ErrorMessageRunFacet.json",
					"message":    err.Error(),
					"programmingLanguage": "go",
				},
			},
		},
		Job: Job{
			Namespace: c.namespace,
			Name:      jobName,
			Facets:    map[string]interface{}{},
		},
	}

	return c.EmitEvent(ctx, event)
}

// Builder helps construct OpenLineage events
type Builder struct {
	event Event
}

// NewEventBuilder creates a new event builder
func NewEventBuilder(eventType, runID, jobName, namespace string) *Builder {
	return &Builder{
		event: Event{
			EventType: eventType,
			EventTime: time.Now().UTC(),
			Run: Run{
				RunID:  runID,
				Facets: make(map[string]interface{}),
			},
			Job: Job{
				Namespace: namespace,
				Name:      jobName,
				Facets:    make(map[string]interface{}),
			},
			Producer:  "opendq-go",
			SchemaURL: "https://openlineage.io/spec/2-0-2/OpenLineage.json",
		},
	}
}

// WithInputs adds input datasets
func (b *Builder) WithInputs(inputs []Dataset) *Builder {
	b.event.Inputs = inputs
	return b
}

// WithOutputs adds output datasets
func (b *Builder) WithOutputs(outputs []Dataset) *Builder {
	b.event.Outputs = outputs
	return b
}

// WithRunFacet adds a run facet
func (b *Builder) WithRunFacet(name string, facet interface{}) *Builder {
	b.event.Run.Facets[name] = facet
	return b
}

// WithJobFacet adds a job facet
func (b *Builder) WithJobFacet(name string, facet interface{}) *Builder {
	b.event.Job.Facets[name] = facet
	return b
}

// Build returns the constructed event
func (b *Builder) Build() Event {
	return b.event
}
