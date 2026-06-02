package types

import "github.com/struckchure/idp/internals"

type ListProjectArgs struct {
	internals.BaseListFilterArgs

	OwnerId string `json:"ownerId"`
}

type CreateProjectArgs struct {
	OwnerId string `json:"ownerId"`
	Name    string `json:"name"`
}

type GetProjectArgs struct {
	Id string `json:"id"`
}

type UpdateProjectArgs struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type DeleteProjectArgs struct {
	GetProjectArgs
}
