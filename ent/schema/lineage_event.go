package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LineageEvent holds the schema definition for the LineageEvent entity (OpenLineage compatible).
type LineageEvent struct {
	ent.Schema
}

// Fields of the LineageEvent.
func (LineageEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("event_type").
			NotEmpty().
			Comment("START, RUNNING, COMPLETE, FAIL, ABORT"),
		field.Time("event_time").
			Default(time.Now),
		field.JSON("run", map[string]interface{}{}).
			Comment("OpenLineage Run facet"),
		field.JSON("job", map[string]interface{}{}).
			Comment("OpenLineage Job facet"),
		field.JSON("inputs", []map[string]interface{}{}).
			Optional().
			Comment("Input datasets"),
		field.JSON("outputs", []map[string]interface{}{}).
			Optional().
			Comment("Output datasets"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the LineageEvent.
func (LineageEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("lineage_events").
			Unique().
			Required(),
		edge.From("workflow", Workflow.Type).
			Ref("lineage_events").
			Unique().
			Optional(),
	}
}

// Indexes of the LineageEvent.
func (LineageEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_type"),
		index.Fields("event_time"),
	}
}
