package main

import (
	"context"
	"fmt"
	"github.com/bfbarry/coop-storage/metadata-server/relational_db"
	"github.com/bfbarry/coop-storage/metadata-server/relational_db/generated"
)

// type Metadata = db.Metadata
// type Account = db.Account
// type CreateMetadataParams = db.CreateMetadataParams

type IMetadataRepo interface {
	// Metadata
	CreateMetadata(ctx context.Context, params *db.CreateMetadataParams) (*db.Metadata, error)
	GetMetadata(ctx context.Context, id int32) (*db.Metadata, error)
	GetChildren(ctx context.Context, parentID *int32) ([]*db.Metadata, error)
	GetChildrenHome(ctx context.Context, ownerID int32) ([]*db.Metadata, error)
	UpdateMeta(ctx context.Context, arg *db.UpdateMetaParams) (*db.Metadata, error)
	DeleteMetadata(ctx context.Context, id int32) error

	// Accounts
	CreateAccount(ctx context.Context, email string) (*db.Account, error)
	GetAccount(ctx context.Context, id int32) (*db.Account, error)
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
