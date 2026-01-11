package workflow

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
)

// Engine manages workflow state machines
type Engine struct {
	workflows map[string]*fsm.FSM
}

// NewEngine creates a new workflow engine
func NewEngine() *Engine {
	return &Engine{
		workflows: make(map[string]*fsm.FSM),
	}
}

// WorkflowDefinition defines a workflow
type WorkflowDefinition struct {
	Name        string
	InitialState string
	Events      []Event
	Callbacks   map[string]fsm.Callback
}

// Event defines a workflow event/transition
type Event struct {
	Name string
	Src  []string // Source states
	Dst  string   // Destination state
}

// CreateWorkflow creates a new workflow from a definition
func (e *Engine) CreateWorkflow(def WorkflowDefinition) (*fsm.FSM, error) {
	events := make([]fsm.EventDesc, len(def.Events))
	for i, event := range def.Events {
		events[i] = fsm.EventDesc{
			Name: event.Name,
			Src:  event.Src,
			Dst:  event.Dst,
		}
	}

	workflow := fsm.NewFSM(
		def.InitialState,
		events,
		def.Callbacks,
	)

	e.workflows[def.Name] = workflow
	return workflow, nil
}

// GetWorkflow retrieves a workflow by name
func (e *Engine) GetWorkflow(name string) (*fsm.FSM, error) {
	workflow, exists := e.workflows[name]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", name)
	}
	return workflow, nil
}

// Transition executes a workflow transition
func (e *Engine) Transition(ctx context.Context, workflowName, event string) error {
	workflow, err := e.GetWorkflow(workflowName)
	if err != nil {
		return err
	}

	if err := workflow.Event(ctx, event); err != nil {
		return fmt.Errorf("transition failed: %w", err)
	}

	return nil
}

// GetCurrentState returns the current state of a workflow
func (e *Engine) GetCurrentState(workflowName string) (string, error) {
	workflow, err := e.GetWorkflow(workflowName)
	if err != nil {
		return "", err
	}
	return workflow.Current(), nil
}

// CanTransition checks if a transition is possible
func (e *Engine) CanTransition(workflowName, event string) (bool, error) {
	workflow, err := e.GetWorkflow(workflowName)
	if err != nil {
		return false, err
	}
	return workflow.Can(event), nil
}

// AvailableTransitions returns available transitions from current state
func (e *Engine) AvailableTransitions(workflowName string) ([]string, error) {
	workflow, err := e.GetWorkflow(workflowName)
	if err != nil {
		return nil, err
	}
	return workflow.AvailableTransitions(), nil
}

// Standard workflow definitions

// DataQualityWorkflow defines a data quality workflow
func DataQualityWorkflow() WorkflowDefinition {
	return WorkflowDefinition{
		Name:        "data_quality",
		InitialState: "pending",
		Events: []Event{
			{Name: "start", Src: []string{"pending"}, Dst: "running"},
			{Name: "validate", Src: []string{"running"}, Dst: "validating"},
			{Name: "pass", Src: []string{"validating"}, Dst: "passed"},
			{Name: "fail", Src: []string{"validating"}, Dst: "failed"},
			{Name: "retry", Src: []string{"failed"}, Dst: "pending"},
			{Name: "complete", Src: []string{"passed"}, Dst: "completed"},
			{Name: "abort", Src: []string{"pending", "running", "validating"}, Dst: "aborted"},
		},
		Callbacks: map[string]fsm.Callback{},
	}
}

// ApprovalWorkflow defines an approval workflow
func ApprovalWorkflow() WorkflowDefinition {
	return WorkflowDefinition{
		Name:        "approval",
		InitialState: "draft",
		Events: []Event{
			{Name: "submit", Src: []string{"draft"}, Dst: "submitted"},
			{Name: "review", Src: []string{"submitted"}, Dst: "under_review"},
			{Name: "approve", Src: []string{"under_review"}, Dst: "approved"},
			{Name: "reject", Src: []string{"under_review"}, Dst: "rejected"},
			{Name: "request_changes", Src: []string{"under_review"}, Dst: "changes_requested"},
			{Name: "resubmit", Src: []string{"changes_requested", "rejected"}, Dst: "submitted"},
			{Name: "cancel", Src: []string{"draft", "submitted", "under_review"}, Dst: "cancelled"},
		},
		Callbacks: map[string]fsm.Callback{},
	}
}

// DataPipelineWorkflow defines a data pipeline workflow
func DataPipelineWorkflow() WorkflowDefinition {
	return WorkflowDefinition{
		Name:        "data_pipeline",
		InitialState: "pending",
		Events: []Event{
			{Name: "start", Src: []string{"pending"}, Dst: "running"},
			{Name: "extract", Src: []string{"running"}, Dst: "extracting"},
			{Name: "transform", Src: []string{"extracting"}, Dst: "transforming"},
			{Name: "load", Src: []string{"transforming"}, Dst: "loading"},
			{Name: "complete", Src: []string{"loading"}, Dst: "completed"},
			{Name: "fail", Src: []string{"extracting", "transforming", "loading"}, Dst: "failed"},
			{Name: "retry", Src: []string{"failed"}, Dst: "pending"},
			{Name: "abort", Src: []string{"pending", "running", "extracting", "transforming", "loading"}, Dst: "aborted"},
		},
		Callbacks: map[string]fsm.Callback{},
	}
}

// RegisterStandardWorkflows registers standard workflow definitions
func (e *Engine) RegisterStandardWorkflows() error {
	workflows := []WorkflowDefinition{
		DataQualityWorkflow(),
		ApprovalWorkflow(),
		DataPipelineWorkflow(),
	}

	for _, def := range workflows {
		if _, err := e.CreateWorkflow(def); err != nil {
			return fmt.Errorf("failed to register workflow %s: %w", def.Name, err)
		}
	}

	return nil
}
