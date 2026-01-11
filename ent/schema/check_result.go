package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CheckResult holds the schema definition for the CheckResult entity.
type CheckResult struct {
	ent.Schema
}

// Fields of the CheckResult.
func (CheckResult) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("status").
			NotEmpty().
			Comment("passed, failed, warning, error, skipped"),
		field.JSON("actual_value", map[string]interface{}{}).
			Optional().
			Comment("The actual value observed"),
		field.JSON("expected_value", map[string]interface{}{}).
			Optional().
			Comment("The expected value"),
		field.String("message").
			Optional(),
		field.JSON("details", map[string]interface{}{}).
			Optional(),
		field.Int64("duration_ms").
			Default(0).
			Comment("Execution duration in milliseconds"),
		field.String("error").
			Optional(),
		field.Time("timestamp").
			Default(time.Now),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the CheckResult.
func (CheckResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("check", Check.Type).
			Ref("results").
			Unique().
			Required(),
		edge.From("execution", ScheduleExecution.Type).
			Ref("results").
			Unique(),
	}
}

// Indexes of the CheckResult.
func (CheckResult) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("timestamp"),
	}
}
