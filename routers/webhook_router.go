package routers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/struckchure/idp/handlers"
)

func NewWebhookRouter(app *fiber.App, webhookHandler handlers.IWebhookHandler) {
	webhookGroup := app.Group("/api/v1/")

	webhookGroup.Post("/webhook/github/", webhookHandler.GithubWebhookHandler)
}
