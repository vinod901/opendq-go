package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Check holds the schema definition for the Check entity (data quality check).
type Check struct {
	ent.Schema
}

// Fields of the Check.
func (Check) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("type").
			NotEmpty().
			Comment("row_count, null_check, uniqueness, freshness, custom_sql, min_value, max_value, etc."),
		field.String("table").
			NotEmpty().
			Comment("Target table name"),
		field.String("column").
			Optional().
			Comment("Target column name if applicable"),
		field.JSON("parameters", map[string]interface{}{}).
			Comment("Check-specific parameters"),
		field.JSON("threshold", map[string]interface{}{}).
			Comment("Pass/fail threshold configuration"),
		field.String("severity").
			Default("medium").
			Comment("critical, high, medium, low, info"),
		field.JSON("tags", []string{}).
			Optional(),
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
		field.Bool("active").
			Default(true),
		field.String("schedule_id").
			Optional().
			Comment("Associated schedule ID"),
		field.String("view_id").
			Optional().
			Comment("Associated view ID for logical view checks"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("last_run_at").
			Optional(),
		field.String("last_status").
			Default("pending").
			Comment("pending, running, passed, failed, warning, error, skipped"),
	}
}

// Edges of the Check.
func (Check) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("checks").
			Unique().
			Required(),
		edge.From("datasource", Datasource.Type).
			Ref("checks").
			Unique().
			Required(),
		edge.To("results", CheckResult.Type),
		edge.From("schedule", Schedule.Type).
			Ref("checks").
			Unique(),
		edge.From("view", View.Type).
			Ref("checks").
			Unique(),
	}
}

// Indexes of the Check.
func (Check) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type"),
		index.Fields("active"),
		index.Fields("severity"),
		index.Fields("last_status"),
	}
}
