package services

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/struckchure/idp/prisma/db"
)

func TestMachineDeploymentName(t *testing.T) {
	t.Parallel()

	name, err := MachineDeploymentName(&db.MachineModel{})
	require.ErrorIs(t, err, ErrMachineNoContainer)
	require.Empty(t, name)

	id := "my-deployment-abc"
	machine := &db.MachineModel{}
	machine.InnerMachine.ContainerID = &id

	name, err = MachineDeploymentName(machine)
	require.NoError(t, err)
	require.Equal(t, "my-deployment-abc", name)

	blank := "   "
	machine.InnerMachine.ContainerID = &blank
	name, err = MachineDeploymentName(machine)
	require.ErrorIs(t, err, ErrMachineNoContainer)
	require.Empty(t, name)
}
