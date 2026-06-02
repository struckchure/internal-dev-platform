package main

import (
	"context"
	"fmt"

	_ "ariga.io/atlas-provider-gorm/gormschema"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"

	"github.com/struckchure/idp/app"
	"github.com/struckchure/idp/dao"
	"github.com/struckchure/idp/handlers"
	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/middlewares"
	"github.com/struckchure/idp/routers"
	"github.com/struckchure/idp/seeds"
	"github.com/struckchure/idp/services"
	"github.com/struckchure/idp/tasks"
)

func start(
	lc fx.Lifecycle,
	app *fiber.App,
	env internals.Env,
	scheduler *internals.Scheduler,
	webSocketHub *internals.WebSocketHub,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go app.Listen(fmt.Sprintf("0.0.0.0:%s", env.APP_PORT))
			go scheduler.Start()
			webSocketHub.Start(env)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			app.Shutdown()
			scheduler.Stop()
			webSocketHub.Stop()

			return nil
		},
	})
}

// @title idp API
// @version 1.0
// @contact.name Mohammed Al-Ameen
// @contact.email ameenmohammed2311@gmail.com
func main() {
	fx.New(
		fx.Provide(internals.NewEnv),

		fx.Provide(internals.NewDatabaseConnection),
		fx.Provide(internals.NewRabbitMQ),
		fx.Provide(internals.NewRabbitMQConnection),
		fx.Provide(internals.NewRedis),
		fx.Provide(internals.NewRedisConnection),
		fx.Provide(internals.NewHasher),
		fx.Provide(internals.NewJwt),
		fx.Provide(internals.NewWebSocketHub),
		fx.Provide(internals.NewScheduler),
		fx.Provide(internals.NewSchedulerCron),

		fx.Provide(internals.NewK8SConfig),
		fx.Provide(internals.NewContainerManager),
		fx.Provide(internals.NewNetworkManager),
		fx.Provide(internals.NewInformer),

		fx.Provide(dao.NewMachineDao),
		fx.Provide(dao.NewNetworkDao),
		fx.Provide(dao.NewRepoConnectionDao),
		fx.Provide(dao.NewUserDao),
		fx.Provide(dao.NewDeploymentDao),
		fx.Provide(dao.NewDeploymentLogDao),
		fx.Provide(dao.NewGithubAccountConnectionDao),

		fx.Provide(services.NewGithubService),
		fx.Provide(services.NewMachineService),
		fx.Provide(services.NewNetworkService),
		fx.Provide(services.NewRepoConnectionService),
		fx.Provide(services.NewUserService),
		fx.Provide(services.NewDeploymentService),
		fx.Provide(services.NewDeploymentLogService),

		fx.Provide(handlers.NewCallbackHandler),
		fx.Provide(handlers.NewDeploymentHandler),
		fx.Provide(handlers.NewGithubHandler),
		fx.Provide(handlers.NewMachineHandler),
		fx.Provide(handlers.NewNetworkHandler),
		fx.Provide(handlers.NewWebhookHandler),
		fx.Provide(handlers.NewRepoConnectionHandler),
		fx.Provide(handlers.NewUserHandler),

		fx.Provide(middlewares.NewUserMiddleware),
		fx.Provide(middlewares.NewJwtMiddleware),

		fx.Invoke(routers.NewProjectRouter),
		fx.Invoke(routers.NewMachineRouter),
		fx.Invoke(routers.NewUsersRouter),
		fx.Invoke(routers.NewWebhookRouter),
		fx.Invoke(routers.NewGithubRouter),
		fx.Invoke(routers.NewCallbackRouter),

		fx.Invoke(seeds.UserSeed),

		fx.Provide(tasks.NewDeploymentTasks),
		fx.Provide(tasks.NewGithubTasks),
		fx.Provide(tasks.NewMachineTasks),
		fx.Invoke(tasks.NewRootTasks),

		fx.Provide(app.NewApp),
		fx.Invoke(start),
	).Run()
}
