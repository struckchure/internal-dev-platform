package services

import (
	"github.com/gofiber/fiber/v3"
	"github.com/samber/lo"

	"github.com/struckchure/idp/dao"
	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/prisma/db"
	"github.com/struckchure/idp/types"
)

type IUserService interface {
	RegisterUser(types.RegisterUserArgs) (*types.RegisterUserResult, error)
	LoginUser(types.LoginUserArgs) (*types.LoginUserResult, error)
	RefreshAccessToken(types.RefreshAccessTokenArgs) (*types.LoginUserResult, error)

	GetProfileUser(types.GetUserArgs) (*db.UserModel, error)
	UpdateUser(types.UpdateUserArgs) (*db.UserModel, error)
}

type UserService struct {
	hasher  internals.IHasher
	jwt     internals.IJwt
	userDAO dao.IUserDao
}

func (u *UserService) GetProfileUser(args types.GetUserArgs) (*db.UserModel, error) {
	user, err := u.userDAO.GetUser(types.GetUserArgs{ID: args.ID})

	userWithoutPassword := internals.RemoveField(user, internals.RemoveFieldOptionFunc("password"))

	return userWithoutPassword, err
}

func (u *UserService) LoginUser(args types.LoginUserArgs) (*types.LoginUserResult, error) {
	user, err := u.userDAO.GetUser(types.GetUserArgs{Email: &args.Email})
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid credentials")
	}

	userPassword, _ := user.Password()
	passwordIsNotCorrect := !u.hasher.PasswordIsCorrect(userPassword, args.Password)
	if passwordIsNotCorrect {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid credentials")
	}

	tokens, err := u.jwt.GenerateJWT(user.ID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return &types.LoginUserResult{Tokens: *tokens}, nil
}

func (u *UserService) RefreshAccessToken(args types.RefreshAccessTokenArgs) (*types.LoginUserResult, error) {
	claims, err := u.jwt.VerifyJWT(args.RefreshToken, internals.REFRESH_TOKEN_TYPE)
	if err != nil {
		return nil, err
	}

	userId, err := claims.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	tokens, err := u.jwt.GenerateJWT(userId)
	if err != nil {
		return nil, err
	}

	return &types.LoginUserResult{Tokens: *tokens}, nil
}

func (u *UserService) RegisterUser(args types.RegisterUserArgs) (*types.RegisterUserResult, error) {
	if args.Password != "" {
		args.Password = u.hasher.HashPassword(args.Password)
	}

	user, err := u.userDAO.CreateUser(types.CreateUserArgs{
		FirstName: args.FirstName,
		LastName:  args.LastName,
		Email:     args.Email,
		Password:  args.Password,
	})

	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	tokens, err := u.jwt.GenerateJWT(user.ID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user = internals.RemoveField(user, internals.RemoveFieldOptionFunc("password"))

	return &types.RegisterUserResult{UserModel: *user, Tokens: *tokens}, nil
}

func (u *UserService) UpdateUser(args types.UpdateUserArgs) (*db.UserModel, error) {
	if args.Password != nil && len(*args.Password) > 0 {
		args.Password = lo.ToPtr(u.hasher.HashPassword(*args.Password))
	}

	user, err := u.userDAO.UpdateUser(types.UpdateUserArgs{
		ID:        args.ID,
		FirstName: args.FirstName,
		LastName:  args.LastName,
		Email:     args.Email,
		Password:  args.Password,
	})
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user = internals.RemoveField(user, internals.RemoveFieldOptionFunc("password"))

	return user, nil
}

func NewUserService(
	hasher internals.IHasher,
	jwt internals.IJwt,
	userDAO dao.IUserDao,
) IUserService {
	return &UserService{
		hasher:  hasher,
		jwt:     jwt,
		userDAO: userDAO,
	}
}
