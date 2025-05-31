package repo

import "context"

type Auth interface {
	BeginTx(ctx context.Context) (Auth, error)
	Commit() error
	Rollback() error

	Create(opts CreateUserOpts) (CreateUserResult, error)
	Read(ctx context.Context) (ReadUsersResult, error)
	Update(ctx context.Context, opts UpdateUsersOpts) error
	Delete(ctx context.Context, id int32) error
	Authorize(opts AuthorizeOpts) (int, error)
	GetEmailsByIDs(ctx context.Context, userID []int32) ([]string, error)
}
