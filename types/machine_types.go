package types

import (
	"github.com/struckchure/idp/internals"
	"github.com/struckchure/idp/prisma/db"
)

type ListMachineArgs struct {
	internals.BaseListFilterArgs

	UserId *string `swag-validate:"optional"`
}

type CreateMachineArgs struct {
	OwnerId       string           `swaggerignore:"true"`
	CPU           string           `json:"cpu"`
	Memory        string           `json:"memory"`
	MachineName   string           `json:"machineName"`
	MachineImage  string           `json:"machineImage"`
	ContainerId   string           `swaggerignore:"true"`
	MachineStatus db.MachineStatus `swaggerignore:"true"`
}

type GetMachineArgs struct {
	Id          *string
	ContainerId *string
}

type UpdateMachineArgs struct {
	ID            string                   `swaggerignore:"true"`
	OwnerId       *string                  `json:"ownerId" swag-validate:"optional"`
	CPU           *string                  `json:"cpu" swag-validate:"optional"`
	Memory        *string                  `json:"memory" swag-validate:"optional"`
	ContainerId   *string                  `swaggerignore:"true"`
	MachineName   *string                  `json:"machineName" swag-validate:"optional"`
	MachineImage  *string                  `json:"machineImage" swag-validate:"optional"`
	MachineStatus *db.MachineStatus        `swaggerignore:"true"`
	Ports         *[]internals.NetworkPort `swaggerignore:"true"` // TODO: fix type
}

type DeleteMachineArgs struct {
	Id string
}

type CreateMachineEventHandlerArgs struct {
	Id string
}

type DeleteMachineEventHandlerArgs struct {
	Id          string
	ContainerId *string
}
