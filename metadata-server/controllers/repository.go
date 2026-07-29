package controllers

import (
	"context"
	"fmt"
	"github.com/bfbarry/coop-storage/metadata-server/relational_db"
	"github.com/bfbarry/coop-storage/metadata-server/relational_db/generated"
)

type Metadata = db.Metadata
type Account = db.Account
type CreateAccountParams = db.CreateAccountParams
type Permission = db.Permission
type CreateMetadataParams = db.CreateMetadataParams
type UpdateMetaParams = db.UpdateMetaParams
type UpdateAccountParams = db.UpdateAccountParams
type CreatePermissionParams = db.CreatePermissionParams
type GetPermissionParams = db.GetPermissionParams

type IMetadataRepo interface {
	// Metadata
	CreateMetadata(ctx context.Context, params *db.CreateMetadataParams) (*db.Metadata, error)
	GetMetadata(ctx context.Context, id int32) (*db.Metadata, error)
	GetChildren(ctx context.Context, parentID *int32) ([]*db.Metadata, error)
	GetChildrenHome(ctx context.Context, ownerID int32) ([]*db.Metadata, error)
	UpdateMeta(ctx context.Context, arg *db.UpdateMetaParams) (*db.Metadata, error)
	DeleteMetadata(ctx context.Context, id int32) error

	// Accounts
	CreateAccount(ctx context.Context, params *db.CreateAccountParams) (*db.Account, error)
	GetAccount(ctx context.Context, id int32) (*db.Account, error)
	GetAccountByClerkID(ctx context.Context, clerkID string) (*db.Account, error)
	UpdateAccount(ctx context.Context, arg *db.UpdateAccountParams) (*db.Account, error)
	DeleteAccount(ctx context.Context, id int32) error

	// Permissions
	CreatePermission(ctx context.Context, arg *db.CreatePermissionParams) (*db.Permission, error)
	GetPermission(ctx context.Context, arg *db.GetPermissionParams) (*db.Permission, error)
	DeletePermission(ctx context.Context, id int32) error
}

type MetadataRepo struct {
	SQLdb relational_db.DbPoolEngine
}

func NewMetaRepo(SQLdb relational_db.DbPoolEngine) *MetadataRepo {
	return &MetadataRepo{
		SQLdb,
	}
}

func (this *MetadataRepo) CreateMetadata(ctx context.Context, params *db.CreateMetadataParams) (*db.Metadata, error) {
	m, err := this.SQLdb.Queries.CreateMetadata(ctx, *params)
	if err != nil {
		return nil, fmt.Errorf("Failed to create meta: %v", err)
	}
	return &m, nil

}
func (this *MetadataRepo) GetMetadata(ctx context.Context, id int32) (*db.Metadata, error) {
	m, err := this.SQLdb.Queries.GetMetadata(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get metadata: %v", err)
	}
	return &m, nil
}
func (this *MetadataRepo) GetChildren(ctx context.Context, parentID *int32) ([]*db.Metadata, error) {
	rows, err := this.SQLdb.Queries.GetChildren(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get children: %v", err)
	}
	items := make([]*db.Metadata, len(rows))
	for i := range rows {
		items[i] = &rows[i]
	}
	return items, nil
}
func (this *MetadataRepo) GetChildrenHome(ctx context.Context, ownerID int32) ([]*db.Metadata, error) {
	rows, err := this.SQLdb.Queries.GetChildrenHome(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get home children: %v", err)
	}
	items := make([]*db.Metadata, len(rows))
	for i := range rows {
		items[i] = &rows[i]
	}
	return items, nil
}
func (this *MetadataRepo) UpdateMeta(ctx context.Context, arg *db.UpdateMetaParams) (*db.Metadata, error) {
	m, err := this.SQLdb.Queries.UpdateMeta(ctx, *arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to update metadata: %v", err)
	}
	return &m, nil
}
func (this *MetadataRepo) DeleteMetadata(ctx context.Context, id int32) error {
	return this.SQLdb.Queries.DeleteMetadata(ctx, id)
}

// Accounts
func (this *MetadataRepo) CreateAccount(ctx context.Context, params *db.CreateAccountParams) (*db.Account, error) {
	a, err := this.SQLdb.Queries.CreateAccount(ctx, *params)
	if err != nil {
		return nil, fmt.Errorf("Failed to create account: %v", err)
	}
	return &a, nil
}
func (this *MetadataRepo) GetAccount(ctx context.Context, id int32) (*db.Account, error) {
	a, err := this.SQLdb.Queries.GetAccount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get account: %v", err)
	}
	return &a, nil
}
func (this *MetadataRepo) GetAccountByClerkID(ctx context.Context, clerkID string) (*db.Account, error) {
	a, err := this.SQLdb.Queries.GetAccountByClerkID(ctx, &clerkID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get account by clerk_id: %v", err)
	}
	return &a, nil
}
func (this *MetadataRepo) UpdateAccount(ctx context.Context, arg *db.UpdateAccountParams) (*db.Account, error) {
	a, err := this.SQLdb.Queries.UpdateAccount(ctx, *arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to update account: %v", err)
	}
	return &a, nil
}
func (this *MetadataRepo) DeleteAccount(ctx context.Context, id int32) error {
	return this.SQLdb.Queries.DeleteAccount(ctx, id)
}

// Permissions
func (this *MetadataRepo) CreatePermission(ctx context.Context, arg *db.CreatePermissionParams) (*db.Permission, error) {
	p, err := this.SQLdb.Queries.CreatePermission(ctx, *arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to create permission: %v", err)
	}
	return &p, nil
}
func (this *MetadataRepo) GetPermission(ctx context.Context, arg *db.GetPermissionParams) (*db.Permission, error) {
	p, err := this.SQLdb.Queries.GetPermission(ctx, *arg)
	if err != nil {
		return nil, fmt.Errorf("Failed to get permission: %v", err)
	}
	return &p, nil
}
func (this *MetadataRepo) DeletePermission(ctx context.Context, id int32) error {
	return this.SQLdb.Queries.DeletePermission(ctx, id)
}
