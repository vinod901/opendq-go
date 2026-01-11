package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertChannel holds the schema definition for the AlertChannel entity.
type AlertChannel struct {
	ent.Schema
}

// Fields of the AlertChannel.
func (AlertChannel) Fields() []ent.Field {
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
			Comment("email, slack, webhook, pagerduty, msteams, opsgenie"),
		field.JSON("configuration", map[string]interface{}{}).
			Comment("Channel-specific configuration"),
		field.Bool("active").
			Default(true),
		field.String("min_severity").
			Default("info").
			Comment("Minimum severity to trigger alert: critical, high, medium, low, info"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the AlertChannel.
func (AlertChannel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("alert_channels").
			Unique().
			Required(),
		edge.To("history", AlertHistory.Type),
	}
}

// Indexes of the AlertChannel.
func (AlertChannel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type"),
		index.Fields("active"),
	}
}
