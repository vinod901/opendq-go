package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// View holds the schema definition for the View entity (logical view).
type View struct {
	ent.Schema
}

// Fields of the View.
func (View) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.JSON("definition", map[string]interface{}{}).
			Comment("View definition (SQL or structured definition)"),
		field.JSON("schema", []map[string]interface{}{}).
			Optional().
			Comment("Inferred or defined schema"),
		field.JSON("tags", []string{}).
			Optional(),
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
		field.Time("validated_at").
			Optional(),
	}
}

// Edges of the View.
func (View) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("views").
			Unique().
			Required(),
		edge.From("datasource", Datasource.Type).
			Ref("views").
			Unique().
			Required(),
		edge.To("checks", Check.Type),
	}
}

// Indexes of the View.
func (View) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("active"),
	}
}
