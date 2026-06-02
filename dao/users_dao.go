package dao

import (
	"context"

	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/prisma/db"
	"github.com/struckchure/idp/types"
)

type IUserDao interface {
	CreateUser(args types.CreateUserArgs) (*db.UserModel, error)
	DeleteUser(args types.DeleteUserArgs) error
	GetUser(args types.GetUserArgs) (*db.UserModel, error)
	ListUsers(args types.ListUsersArgs) ([]db.UserModel, error)
	UpdateUser(args types.UpdateUserArgs) (*db.UserModel, error)
}

type UserDao struct {
	client *db.PrismaClient
	ctx    context.Context
}

func (d *UserDao) ListUsers(args types.ListUsersArgs) ([]db.UserModel, error) {
	args.Skip = internals.UseDefaultValueIf(0, args.Skip, 0)
	args.Take = internals.UseDefaultValueIf(0, args.Take, 10)

	return d.client.User.
		FindMany().
		Skip(internals.UseDefault(args.Skip, 0)).
		Take(internals.UseDefault(args.Take, 10)).
		Exec(d.ctx)
}

func (d *UserDao) CreateUser(args types.CreateUserArgs) (*db.UserModel, error) {
	return d.client.User.
		CreateOne(
			db.User.FirstName.Set(args.FirstName),
			db.User.LastName.Set(args.LastName),
			db.User.Email.Set(args.Email),
			db.User.Password.Set(args.Password),
			db.User.Roles.Set(args.Roles),
		).
		Exec(d.ctx)
}

func (d *UserDao) GetUser(args types.GetUserArgs) (*db.UserModel, error) {
	return d.client.User.
		FindFirst(
			db.User.ID.EqualsIfPresent(args.ID),
			db.User.Email.EqualsIfPresent(args.Email),
		).Exec(d.ctx)
}

func (d *UserDao) UpdateUser(args types.UpdateUserArgs) (*db.UserModel, error) {
	return d.client.User.
		FindUnique(db.User.ID.Equals(args.ID)).
		Update(
			db.User.FirstName.SetIfPresent(args.FirstName),
			db.User.LastName.SetIfPresent(args.LastName),
			db.User.Email.SetIfPresent(args.Email),
			db.User.Password.SetIfPresent(args.Password),
			db.User.Roles.SetIfPresent(args.Roles),
		).
		Exec(d.ctx)
}

func (d *UserDao) DeleteUser(args types.DeleteUserArgs) error {
	_, err := d.client.User.
		FindUnique(db.User.ID.Equals(args.ID)).
		Delete().
		Exec(d.ctx)

	return err
}

func NewUserDao(connection *internals.DatabaseConnection) IUserDao {
	return &UserDao{
		client: connection.Client,
		ctx:    context.Background(),
	}
}
