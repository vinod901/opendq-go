package view

import (
	"context"
	"testing"

	"github.com/vinod901/opendq-go/internal/datasource"
)

func TestNewManager(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.views == nil {
		t.Fatal("views map is nil")
	}
}

func TestManager_CreateView(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	view := &View{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Test View",
		Definition: ViewDefinition{
			SQL: "SELECT * FROM users WHERE active = true",
		},
	}

	err := m.CreateView(ctx, view)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if view.ID == "" {
		t.Error("view ID should be generated")
	}
	if !view.Active {
		t.Error("view should be active by default")
	}
}

func TestManager_CreateView_WithBaseTable(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	view := &View{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Test View",
		Definition: ViewDefinition{
			BaseTable: "users",
			Columns: []ColumnDef{
				{Name: "id"},
				{Name: "email"},
			},
			Filters: []FilterDef{
				{Column: "active", Operator: "eq", Value: true},
			},
		},
	}

	err := m.CreateView(ctx, view)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManager_CreateView_InvalidDefinition(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	view := &View{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Invalid View",
		Definition:   ViewDefinition{}, // Empty definition
	}

	err := m.CreateView(ctx, view)
	if err == nil {
		t.Fatal("expected error for invalid view definition")
	}
}

func TestManager_GetView(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	view := &View{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Test View",
		Definition: ViewDefinition{
			SQL: "SELECT * FROM users",
		},
	}
	m.CreateView(ctx, view)

	retrieved, err := m.GetView(ctx, view.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.Name != view.Name {
		t.Errorf("expected name %s, got %s", view.Name, retrieved.Name)
	}
}

func TestManager_GetView_NotFound(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	_, err := m.GetView(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent view")
	}
}

func TestManager_DeleteView(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	view := &View{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "Test View",
		Definition: ViewDefinition{
			SQL: "SELECT * FROM users",
		},
	}
	m.CreateView(ctx, view)

	err := m.DeleteView(ctx, view.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.GetView(ctx, view.ID)
	if err == nil {
		t.Fatal("expected error for deleted view")
	}
}

func TestManager_ListViews(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	m.CreateView(ctx, &View{
		TenantID:     "tenant-1",
		DatasourceID: "ds-1",
		Name:         "View 1",
		Definition:   ViewDefinition{SQL: "SELECT 1"},
	})
	m.CreateView(ctx, &View{
		TenantID:     "tenant-1",
		DatasourceID: "ds-2",
		Name:         "View 2",
		Definition:   ViewDefinition{SQL: "SELECT 2"},
	})

	// List all
	views, err := m.ListViews(ctx, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 2 {
		t.Errorf("expected 2 views, got %d", len(views))
	}

	// Filter by datasource
	views, err = m.ListViews(ctx, "", "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 1 {
		t.Errorf("expected 1 view for ds-1, got %d", len(views))
	}
}

func TestBuildFilterCondition(t *testing.T) {
	testCases := []struct {
		name     string
		filter   FilterDef
		expected string
	}{
		{
			"equals",
			FilterDef{Column: "status", Operator: "eq", Value: "active"},
			"status = 'active'",
		},
		{
			"not equals",
			FilterDef{Column: "status", Operator: "ne", Value: "inactive"},
			"status <> 'inactive'",
		},
		{
			"less than",
			FilterDef{Column: "age", Operator: "lt", Value: 18},
			"age < 18",
		},
		{
			"greater than",
			FilterDef{Column: "age", Operator: "gt", Value: 21},
			"age > 21",
		},
		{
			"is null",
			FilterDef{Column: "deleted_at", Operator: "is_null"},
			"deleted_at IS NULL",
		},
		{
			"is not null",
			FilterDef{Column: "email", Operator: "is_not_null"},
			"email IS NOT NULL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := buildFilterCondition(tc.filter)
			if result != tc.expected {
				t.Errorf("buildFilterCondition() = %s, want %s", result, tc.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"string", "hello", "'hello'"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"nil", nil, "NULL"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatValue(tc.value)
			if result != tc.expected {
				t.Errorf("formatValue(%v) = %s, want %s", tc.value, result, tc.expected)
			}
		})
	}
}

func TestValidateViewDefinition(t *testing.T) {
	dsManager := datasource.NewManager()
	m := NewManager(dsManager)
	ctx := context.Background()

	testCases := []struct {
		name    string
		def     ViewDefinition
		wantErr bool
	}{
		{
			"valid SQL",
			ViewDefinition{SQL: "SELECT * FROM users"},
			false,
		},
		{
			"valid base table",
			ViewDefinition{BaseTable: "users"},
			false,
		},
		{
			"valid union",
			ViewDefinition{UnionTables: []string{"table1", "table2"}},
			false,
		},
		{
			"empty definition",
			ViewDefinition{},
			true,
		},
		{
			"invalid join type",
			ViewDefinition{
				BaseTable: "users",
				Joins: []JoinDef{
					{Table: "orders", Type: "invalid"},
				},
			},
			true,
		},
		{
			"missing join table",
			ViewDefinition{
				BaseTable: "users",
				Joins: []JoinDef{
					{Type: "inner"},
				},
			},
			true,
		},
		{
			"invalid filter operator",
			ViewDefinition{
				BaseTable: "users",
				Filters: []FilterDef{
					{Column: "status", Operator: "invalid"},
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			view := &View{
				TenantID:     "tenant-1",
				DatasourceID: "ds-1",
				Name:         "Test",
				Definition:   tc.def,
			}
			err := m.validateViewDefinition(ctx, view)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateViewDefinition() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
