package internals

import (
	"log"
	"os"
)

type Env struct {
	DATABASE_URL string

	APP_PORT    string
	SOCKET_PORT string

	JWT_ACCESS_KEY  string
	JWT_REFRESH_KEY string

	GH_APP_SLUG          string
	GH_APP_ID            string
	GH_APP_CLIENT_ID     string
	GH_APP_CLIENT_SECRET string
	GH_PRIVATE_KEY       string
	GH_APP_REDIRECT_URL  string

	DEFAULT_ADMIN_EMAIL string
	DEFAULT_ADMIN_PASS  string

	RABBITMQ_URL string

	K8S_CLUSTER_CONFIG string

	INGRESS_ROOT_DOMAIN string

	REDIS_URL string
}

func NewEnv() Env {
	var secrets, err = SecretsFactory(UseDefault(os.Getenv("SECRET_FROM"), "local"))
	if err != nil {
		log.Println(err)
	}

	return Env{
		DATABASE_URL: secrets.GetOrPanic("DATABASE_URL"),

		APP_PORT:    UseDefault(secrets.Get("APP_PORT"), "3000"),
		SOCKET_PORT: UseDefault(secrets.Get("SOCKET_PORT"), "9090"),

		JWT_ACCESS_KEY:  secrets.GetOrPanic("JWT_ACCESS_KEY"),
		JWT_REFRESH_KEY: secrets.GetOrPanic("JWT_REFRESH_KEY"),

		GH_APP_SLUG:          secrets.GetOrPanic("GH_APP_SLUG"),
		GH_APP_ID:            secrets.GetOrPanic("GH_APP_ID"),
		GH_APP_CLIENT_ID:     secrets.GetOrPanic("GH_APP_CLIENT_ID"),
		GH_APP_CLIENT_SECRET: secrets.GetOrPanic("GH_APP_CLIENT_SECRET"),
		GH_PRIVATE_KEY:       secrets.GetOrPanic("GH_PRIVATE_KEY"),
		GH_APP_REDIRECT_URL:  UseDefault(secrets.Get("GH_APP_REDIRECT_URL"), "http://localhost:3000/api/v1/callback/github/"),

		DEFAULT_ADMIN_EMAIL: secrets.GetOrPanic("DEFAULT_ADMIN_EMAIL"),
		DEFAULT_ADMIN_PASS:  secrets.GetOrPanic("DEFAULT_ADMIN_PASS"),

		RABBITMQ_URL: secrets.GetOrPanic("RABBITMQ_URL"),
		REDIS_URL:    secrets.GetOrPanic("REDIS_URL"),

		K8S_CLUSTER_CONFIG:  secrets.GetOrPanic("K8S_CLUSTER_CONFIG"),
		INGRESS_ROOT_DOMAIN: secrets.GetOrPanic("INGRESS_ROOT_DOMAIN"),
	}
}
