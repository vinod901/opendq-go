package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ScheduleExecution holds the schema definition for the ScheduleExecution entity.
type ScheduleExecution struct {
	ent.Schema
}

// Fields of the ScheduleExecution.
func (ScheduleExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("status").
			Default("running").
			Comment("running, completed, failed, partial"),
		field.Time("started_at").
			Default(time.Now),
		field.Time("completed_at").
			Optional(),
		field.Int64("duration_ms").
			Default(0).
			Comment("Total execution duration in milliseconds"),
		field.Int("total_checks").
			Default(0),
		field.Int("passed_checks").
			Default(0),
		field.Int("failed_checks").
			Default(0),
		field.Int("warning_checks").
			Default(0),
		field.Int("error_checks").
			Default(0),
		field.Int("skipped_checks").
			Default(0),
		field.String("error").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the ScheduleExecution.
func (ScheduleExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("schedule", Schedule.Type).
			Ref("executions").
			Unique().
			Required(),
		edge.To("results", CheckResult.Type),
	}
}

// Indexes of the ScheduleExecution.
func (ScheduleExecution) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("started_at"),
	}
}
