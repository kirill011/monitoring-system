package repo

import "auth-service/internal/models"

type CreateUserOpts struct {
	Name     string
	Email    string
	Password string
}

type CreateUserResult struct {
	ID    int
	Name  string
	Email string
}

type UpdateUsersOpts struct {
	ID       int
	Name     *string
	Email    *string
	Password *string
}

type ReadUsersResult struct {
	Users []models.User
}

type AuthorizeOpts struct {
	Email    string
	Password string
}
