package routers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/struckchure/idp/handlers"
	"github.com/struckchure/idp/middlewares"
)

func NewUsersRouter(
	app *fiber.App,
	userHandler handlers.IUserHandler,
	userMiddleware middlewares.IUserMiddleware,
	jwtMiddleware middlewares.IJwtMiddleware,
) {
	authGroup := app.Group("/api/v1/auth")
	authGroup.Post("/register/", userHandler.RegisterUser)
	authGroup.Post("/login/", userHandler.LoginUser)
	authGroup.Post("/refresh-access-token/", userHandler.RefreshAccessToken)

	userGroup := app.Group("/api/v1/user", jwtMiddleware.Use, userMiddleware.Use)
	userGroup.Get("/profile/", userHandler.GetProfileUser)
	userGroup.Patch("/profile/", userHandler.UpdateProfileUser)
}
