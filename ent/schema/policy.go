package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Policy holds the schema definition for the Policy entity.
type Policy struct {
	ent.Schema
}

// Fields of the Policy.
func (Policy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.JSON("rules", map[string]interface{}{}).
			Comment("OpenFGA policy rules"),
		field.String("resource_type").
			NotEmpty().
			Comment("Type of resource this policy applies to"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional(),
		field.Bool("active").
			Default(true),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Policy.
func (Policy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("policies").
			Unique().
			Required(),
	}
}

// Indexes of the Policy.
func (Policy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_type"),
		index.Fields("active"),
	}
}
