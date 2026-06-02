package tasks

import (
	"encoding/json"
	"log"

	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/services"
	"github.com/struckchure/idp/types"
)

type GithubTasks struct {
	rmq           internals.RabbitMQ
	githubService services.IGithubService
}

func (t *GithubTasks) DeployRepoTask() {
	t.rmq.SubscribeWithWorkers(2, internals.SubscribeArgs{
		Queue: types.DEPLOYMENT_DEPLOY_REPO_QUEUE,
		Callback: func(body string) error {
			var payload types.DeployRepoArgs
			err := json.Unmarshal([]byte(body), &payload)
			if err != nil {
				return err
			}

			err = t.githubService.DeployRepoHandler(payload)
			if err != nil {
				log.Println(err)

				return err
			}

			return nil
		},
	})
}

func NewGithubTasks(
	rmq internals.RabbitMQ,
	githubService services.IGithubService,
) GithubTasks {
	return GithubTasks{
		rmq:           rmq,
		githubService: githubService,
	}
}
