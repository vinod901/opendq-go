package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Schedule holds the schema definition for the Schedule entity.
type Schedule struct {
	ent.Schema
}

// Fields of the Schedule.
func (Schedule) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("cron_expression").
			NotEmpty().
			Comment("Cron expression for scheduling"),
		field.String("timezone").
			Default("UTC"),
		field.String("datasource_id").
			Optional().
			Comment("Run all checks for this datasource"),
		field.JSON("alert_channel_ids", []string{}).
			Optional().
			Comment("Alert channels to notify on failure"),
		field.Bool("active").
			Default(true),
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("last_run_at").
			Optional(),
		field.Time("next_run_at").
			Optional(),
	}
}

// Edges of the Schedule.
func (Schedule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("schedules").
			Unique().
			Required(),
		edge.To("checks", Check.Type),
		edge.To("executions", ScheduleExecution.Type),
	}
}

// Indexes of the Schedule.
func (Schedule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("active"),
		index.Fields("next_run_at"),
	}
}
