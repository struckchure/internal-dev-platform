package routers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/struckchure/idp/handlers"
	"github.com/struckchure/idp/middlewares"
)

func NewGithubRouter(
	app *fiber.App,
	githubHandler handlers.IGithubHandler,
	userMiddleware middlewares.IUserMiddleware,
	jwtMiddleware middlewares.IJwtMiddleware,
) {
	githubGroup := app.Group("/api/v1/gh", jwtMiddleware.Use, userMiddleware.Use)

	githubGroup.Get("/repos", githubHandler.ListRepositories)
	githubGroup.Get("/authorize", githubHandler.AuthorizeGithubAccount)
	githubGroup.Get("/update-app-access", githubHandler.UpdateAppAccess)
	githubGroup.Get("/account-connections", githubHandler.ListAccountConnections)
}
