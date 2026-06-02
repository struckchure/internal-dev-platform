package internals

import (
	"fmt"
)

type ISecrets interface {
	Load() error
	Get(key string) string
	GetOrPanic(key string) string
}

func SecretsFactory(factory string) (ISecrets, error) {
	var factoryObj ISecrets

	switch factory {
	case "local":
		factoryObj = LocalSecrets{}
	case "aws":
		factoryObj = AwsSecrets{}
	}

	fmt.Println("Using secrets factory: ", factory)

	if factoryObj != nil {
		factoryObj.Load()

		return factoryObj, nil
	}

	return nil, fmt.Errorf("no secrets factory for %s", factory)
}
