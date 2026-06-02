package internals

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListDeploymentPods_emptyName(t *testing.T) {
	t.Parallel()

	manager := &ContainerManager{}
	_, err := manager.ListDeploymentPods("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "deployment name is required")
}
