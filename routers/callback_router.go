package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/struckchure/idp/handlers"
)

func NewCallbackRouter(app *fiber.App, callbackHandler handlers.ICallbackHandler) {
	callbackGroup := app.Group("/api/v1/callback")

	callbackGroup.Get("/github/", callbackHandler.GithubCallbackHandler)
}
