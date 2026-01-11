package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Datasource holds the schema definition for the Datasource entity.
type Datasource struct {
	ent.Schema
}

// Fields of the Datasource.
func (Datasource) Fields() []ent.Field {
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
			Comment("postgres, mysql, sqlserver, oracle, snowflake, databricks, bigquery, trino, duckdb, clickhouse, hdfs, deltalake, iceberg, hudi, s3, gcs, azure_blob, local"),
		field.JSON("connection", map[string]interface{}{}).
			Comment("Connection configuration"),
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
		field.Time("last_connected_at").
			Optional(),
	}
}

// Edges of the Datasource.
func (Datasource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("datasources").
			Unique().
			Required(),
		edge.To("checks", Check.Type),
		edge.To("views", View.Type),
	}
}

// Indexes of the Datasource.
func (Datasource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type"),
		index.Fields("active"),
	}
}
