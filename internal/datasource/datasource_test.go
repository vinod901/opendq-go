package datasource

import (
	"context"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.datasources == nil {
		t.Fatal("datasources map is nil")
	}
	if m.connectors == nil {
		t.Fatal("connectors map is nil")
	}
}

func TestDatasource_Type(t *testing.T) {
	testCases := []struct {
		name     string
		dsType   Type
		expected Type
	}{
		{"PostgreSQL", TypePostgres, TypePostgres},
		{"MySQL", TypeMySQL, TypeMySQL},
		{"Snowflake", TypeSnowflake, TypeSnowflake},
		{"S3", TypeS3, TypeS3},
		{"DeltaLake", TypeDeltaLake, TypeDeltaLake},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := &Datasource{
				Type: tc.dsType,
			}
			if ds.Type != tc.expected {
				t.Errorf("expected type %s, got %s", tc.expected, ds.Type)
			}
		})
	}
}

func TestManager_ListDatasources_Empty(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	datasources, err := m.ListDatasources(ctx, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if datasources != nil && len(datasources) > 0 {
		t.Errorf("expected empty list, got %d datasources", len(datasources))
	}
}

func TestManager_GetDatasource_NotFound(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	_, err := m.GetDatasource(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent datasource")
	}
}

func TestManager_DeleteDatasource_NotFound(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	err := m.DeleteDatasource(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent datasource")
	}
}

func TestManager_UpdateDatasource_NotFound(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	err := m.UpdateDatasource(ctx, "nonexistent", map[string]interface{}{
		"name": "updated",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent datasource")
	}
}

func TestConnectionConfig_Fields(t *testing.T) {
	config := ConnectionConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "pass",
		SSLMode:  "disable",
	}

	if config.Host != "localhost" {
		t.Errorf("expected Host 'localhost', got '%s'", config.Host)
	}
	if config.Port != 5432 {
		t.Errorf("expected Port 5432, got %d", config.Port)
	}
}

func TestPostgresConnector_Type(t *testing.T) {
	config := ConnectionConfig{}
	connector := NewPostgresConnector(config)
	
	if connector.Type() != TypePostgres {
		t.Errorf("expected type %s, got %s", TypePostgres, connector.Type())
	}
}

func TestMySQLConnector_Type(t *testing.T) {
	config := ConnectionConfig{}
	connector := NewMySQLConnector(config)
	
	if connector.Type() != TypeMySQL {
		t.Errorf("expected type %s, got %s", TypeMySQL, connector.Type())
	}
}

func TestSQLServerConnector_Type(t *testing.T) {
	config := ConnectionConfig{}
	connector := NewSQLServerConnector(config)
	
	if connector.Type() != TypeSQLServer {
		t.Errorf("expected type %s, got %s", TypeSQLServer, connector.Type())
	}
}

func TestSnowflakeConnector_Type(t *testing.T) {
	config := ConnectionConfig{}
	connector := NewSnowflakeConnector(config)
	
	if connector.Type() != TypeSnowflake {
		t.Errorf("expected type %s, got %s", TypeSnowflake, connector.Type())
	}
}

func TestStorageConnector_Type(t *testing.T) {
	config := ConnectionConfig{}
	connector := NewStorageConnector(TypeS3, config)
	
	if connector.Type() != TypeS3 {
		t.Errorf("expected type %s, got %s", TypeS3, connector.Type())
	}
}

func TestLakehouseConnector_Type(t *testing.T) {
	config := ConnectionConfig{}
	connector := NewLakehouseConnector(TypeDeltaLake, config)
	
	if connector.Type() != TypeDeltaLake {
		t.Errorf("expected type %s, got %s", TypeDeltaLake, connector.Type())
	}
}

func TestDetectFormat(t *testing.T) {
	testCases := []struct {
		path     string
		expected FileFormat
	}{
		{"data.csv", FormatCSV},
		{"data.parquet", FormatParquet},
		{"data.avro", FormatAvro},
		{"data.json", FormatJSON},
		{"data.jsonl", FormatJSONL},
		{"data.ndjson", FormatJSONL},
		{"data.orc", FormatORC},
		{"data.txt", FormatUnknown},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			format := DetectFormat(tc.path)
			if format != tc.expected {
				t.Errorf("expected format %s, got %s", tc.expected, format)
			}
		})
	}
}
