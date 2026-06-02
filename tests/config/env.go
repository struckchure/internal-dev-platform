package config

import (
	"github.com/struckchure/idp/internals"
)

func NewEnv() internals.Env {
	return internals.Env{
		DATABASE_URL: "postgresql://test_user:test_password@localhost:5432/test_db?sslmode=disable",

		APP_PORT:    "3000",
		SOCKET_PORT: "9090",

		JWT_ACCESS_KEY:  "JWT_ACCESS_KEY",
		JWT_REFRESH_KEY: "JWT_REFRESH_KEY",

		GH_APP_SLUG:          "GH_APP_SLUG",
		GH_APP_ID:            "GH_APP_ID",
		GH_APP_CLIENT_ID:     "GH_APP_CLIENT_ID",
		GH_APP_CLIENT_SECRET: "GH_APP_CLIENT_SECRET",
		GH_PRIVATE_KEY:       "GH_PRIVATE_KEY",

		DEFAULT_ADMIN_EMAIL: "DEFAULT_ADMIN_EMAIL",
		DEFAULT_ADMIN_PASS:  "DEFAULT_ADMIN_PASS",

		RABBITMQ_URL: "RABBITMQ_URL",

		K8S_CLUSTER_CONFIG: "K8S_CLUSTER_CONFIG",

		INGRESS_ROOT_DOMAIN: "INGRESS_ROOT_DOMAIN",

		REDIS_URL: "REDIS_URL",
	}
}
