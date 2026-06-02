package handlers

import (
	"github.com/gofiber/fiber/v3"

	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/services"
	"github.com/struckchure/idp/types"
)

type ICallbackHandler interface {
	GithubCallbackHandler(fiber.Ctx) error
}

type CallbackHandler struct {
	githubService services.IGithubService
}

// GithubCallbackHandler implements CallbackHandlerInterface.
func (w *CallbackHandler) GithubCallbackHandler(c fiber.Ctx) error {
	state := fiber.Query[string](c, "state")
	code := fiber.Query[string](c, "code")

	redirectUrl, err := w.githubService.ConnectGithubAccount(types.ConnectGithubAccountArgs{
		State: state,
		Code:  code,
	})
	if err != nil {
		return internals.TranslateHandlerError(c, err)
	}

	return c.Redirect().Status(fiber.StatusFound).To(*redirectUrl)
}

func NewCallbackHandler(githubService services.IGithubService) ICallbackHandler {
	return &CallbackHandler{githubService: githubService}
}
