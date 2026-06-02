package internals

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateMachineImage(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateMachineImage(MachineImageAlpine))
	require.NoError(t, ValidateMachineImage(MachineImageUbuntu))
	require.Error(t, ValidateMachineImage("nginx:alpine"))
	require.Error(t, ValidateMachineImage(""))
}
