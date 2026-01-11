package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertHistory holds the schema definition for the AlertHistory entity.
type AlertHistory struct {
	ent.Schema
}

// Fields of the AlertHistory.
func (AlertHistory) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("alert_id").
			NotEmpty(),
		field.String("title").
			NotEmpty(),
		field.String("message").
			Optional(),
		field.String("severity").
			NotEmpty().
			Comment("critical, high, medium, low, info"),
		field.String("status").
			NotEmpty().
			Comment("sent, failed"),
		field.String("schedule_id").
			Optional(),
		field.String("execution_id").
			Optional(),
		field.String("check_id").
			Optional(),
		field.JSON("details", map[string]interface{}{}).
			Optional(),
		field.String("error").
			Optional(),
		field.Time("sent_at").
			Default(time.Now),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the AlertHistory.
func (AlertHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", AlertChannel.Type).
			Ref("history").
			Unique().
			Required(),
	}
}

// Indexes of the AlertHistory.
func (AlertHistory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("severity"),
		index.Fields("sent_at"),
	}
}
