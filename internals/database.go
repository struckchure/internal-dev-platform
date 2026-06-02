package internals

import (
	"log"

	"github.com/struckchure/idp/prisma/db"
)

type BaseListFilterArgs struct {
	Skip   int    `query:"skip" swag-validate:"optional"`
	Take   int    `query:"take" swag-validate:"optional"`
	Search string `query:"search" swag-validate:"optional"`
	SortBy string `query:"sortBy" swag-validate:"optional"`
}

type DatabaseConnection struct {
	Client *db.PrismaClient
}

func NewDatabaseConnection(env Env) *DatabaseConnection {
	client := db.NewClient(db.WithDatasourceURL(env.DATABASE_URL))
	if err := client.Prisma.Connect(); err != nil {
		log.Println(err)
	}

	return &DatabaseConnection{Client: client}
}
