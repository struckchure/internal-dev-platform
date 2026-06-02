package tasks

import (
	"context"

	"github.com/struckchure/idp/internals"
	"go.uber.org/fx"
)

func NewRootTasks(
	lc fx.Lifecycle,

	machineTasks MachineTasks,
	deploymentTasks DeploymentTasks,
	githubTasks GithubTasks,

	k8sInformer internals.IInformer,
) {
	k8sInformerChannel := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go machineTasks.CreateMachineTask()
			go machineTasks.UpdateMachineTask()
			go machineTasks.DeleteMachineTask()
			go machineTasks.RedeployOnMachineUpdateTask()

			go deploymentTasks.DeploymentNotificationTask()
			go deploymentTasks.DeploymentLogsTask()

			go githubTasks.DeployRepoTask()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			k8sInformer.Stop(k8sInformerChannel)

			return nil
		},
	})
}
