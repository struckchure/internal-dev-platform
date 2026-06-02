package tasks

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/types"
)

type DeploymentTasks struct {
	rmq          internals.RabbitMQ
	webSocketHub *internals.WebSocketHub
}

func (t *DeploymentTasks) DeploymentNotificationTask() {
	t.rmq.Subscribe(internals.SubscribeArgs{
		Queue: types.DEPLOYMENT_NOTIFICATION_EVENT,
		Callback: func(body string) error {
			var dto map[string]any
			err := json.Unmarshal([]byte(body), &dto)
			if err != nil {
				log.Println(err)

				return err
			}

			t.webSocketHub.Publish(fmt.Sprintf("%s/%v", types.DEPLOYMENT_NOTIFICATION_EVENT, dto["machineId"]), dto)

			return nil
		},
	})
}

func (t *DeploymentTasks) DeploymentLogsTask() {
	t.rmq.Subscribe(internals.SubscribeArgs{
		Queue: types.DEPLOYMENT_LOG_EVENT_QUEUE,
		Callback: func(body string) error {
			// var payload db.DeploymentLogModel
			// err := json.Unmarshal([]byte(body), &payload)
			// if err != nil {
			// 	log.Println(err)

			// 	return err
			// }

			t.webSocketHub.Publish(types.DEPLOYMENT_LOG_STREAM_EVENT, body)

			return nil
		},
	})
}

func NewDeploymentTasks(
	rmq internals.RabbitMQ,
	webSocketHub *internals.WebSocketHub,
) DeploymentTasks {
	return DeploymentTasks{
		rmq:          rmq,
		webSocketHub: webSocketHub,
	}
}
