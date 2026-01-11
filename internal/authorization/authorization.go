package authorization

import (
	"context"
	"fmt"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
)

// Manager handles OpenFGA authorization
type Manager struct {
	client  *client.OpenFgaClient
	storeID string
}

// Config contains OpenFGA configuration
type Config struct {
	APIHost   string
	StoreID   string
	AuthModel string
}

// NewManager creates a new authorization manager
func NewManager(cfg Config) (*Manager, error) {
	configuration := client.ClientConfiguration{
		ApiUrl:  cfg.APIHost,
		StoreId: cfg.StoreID,
	}

	fgaClient, err := client.NewSdkClient(&configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenFGA client: %w", err)
	}

	return &Manager{
		client:  fgaClient,
		storeID: cfg.StoreID,
	}, nil
}

// Check checks if a user has permission to perform an action on a resource
func (m *Manager) Check(ctx context.Context, user, relation, object string) (bool, error) {
	body := client.ClientCheckRequest{
		User:     user,
		Relation: relation,
		Object:   object,
	}

	data, err := m.client.Check(ctx).Body(body).Execute()
	if err != nil {
		return false, fmt.Errorf("authorization check failed: %w", err)
	}

	return data.GetAllowed(), nil
}

// WriteTuple writes a relationship tuple to OpenFGA
func (m *Manager) WriteTuple(ctx context.Context, user, relation, object string) error {
	body := client.ClientWriteRequest{
		Writes: []openfga.TupleKey{
			{
				User:     user,
				Relation: relation,
				Object:   object,
			},
		},
	}

	_, err := m.client.Write(ctx).Body(body).Execute()
	if err != nil {
		return fmt.Errorf("failed to write tuple: %w", err)
	}

	return nil
}

// DeleteTuple deletes a relationship tuple from OpenFGA
func (m *Manager) DeleteTuple(ctx context.Context, user, relation, object string) error {
	body := client.ClientWriteRequest{
		Deletes: []openfga.TupleKeyWithoutCondition{
			{
				User:     user,
				Relation: relation,
				Object:   object,
			},
		},
	}

	_, err := m.client.Write(ctx).Body(body).Execute()
	if err != nil {
		return fmt.Errorf("failed to delete tuple: %w", err)
	}

	return nil
}

// ListObjects lists objects a user has access to for a given relation
func (m *Manager) ListObjects(ctx context.Context, user, relation, objectType string) ([]string, error) {
	body := client.ClientListObjectsRequest{
		User:     user,
		Relation: relation,
		Type:     objectType,
	}

	data, err := m.client.ListObjects(ctx).Body(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return data.GetObjects(), nil
}

// Permission represents a permission check
type Permission struct {
	User     string
	Relation string
	Object   string
}

// CheckMultiple checks multiple permissions at once
func (m *Manager) CheckMultiple(ctx context.Context, permissions []Permission) (map[int]bool, error) {
	results := make(map[int]bool)

	for i, perm := range permissions {
		allowed, err := m.Check(ctx, perm.User, perm.Relation, perm.Object)
		if err != nil {
			return nil, fmt.Errorf("failed to check permission %d: %w", i, err)
		}
		results[i] = allowed
	}

	return results, nil
}

// Common relations for multi-tenant applications
const (
	RelationOwner  = "owner"
	RelationAdmin  = "admin"
	RelationEditor = "editor"
	RelationViewer = "viewer"
	RelationMember = "member"
)

// Common object types
const (
	TypeTenant   = "tenant"
	TypeUser     = "user"
	TypePolicy   = "policy"
	TypeWorkflow = "workflow"
	TypeLineage  = "lineage"
)

// FormatObject formats an object identifier
func FormatObject(objectType, objectID string) string {
	return fmt.Sprintf("%s:%s", objectType, objectID)
}

// FormatUser formats a user identifier
func FormatUser(userType, userID string) string {
	return fmt.Sprintf("%s:%s", userType, userID)
}

// GrantTenantAccess grants a user access to a tenant
func (m *Manager) GrantTenantAccess(ctx context.Context, userID, tenantID, relation string) error {
	user := FormatUser(TypeUser, userID)
	object := FormatObject(TypeTenant, tenantID)
	return m.WriteTuple(ctx, user, relation, object)
}

// RevokeTenantAccess revokes a user's access to a tenant
func (m *Manager) RevokeTenantAccess(ctx context.Context, userID, tenantID, relation string) error {
	user := FormatUser(TypeUser, userID)
	object := FormatObject(TypeTenant, tenantID)
	return m.DeleteTuple(ctx, user, relation, object)
}

// CheckTenantAccess checks if a user has access to a tenant
func (m *Manager) CheckTenantAccess(ctx context.Context, userID, tenantID, relation string) (bool, error) {
	user := FormatUser(TypeUser, userID)
	object := FormatObject(TypeTenant, tenantID)
	return m.Check(ctx, user, relation, object)
}
