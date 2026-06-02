package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAction_stormArray(t *testing.T) {
	t.Parallel()

	content := `
name: Storm workflow
on:
  push:
    branches: [main]
jobs:
  - name: build
    runs-on: self-hosted
    steps:
      - name: Build
        run: echo build
`
	action, err := ParseAction(content)
	require.NoError(t, err)
	require.Len(t, action.Jobs, 1)
	require.Equal(t, "build", action.Jobs[0].Name)
	require.Equal(t, "self-hosted", action.Jobs[0].RunsOn)
	require.Len(t, action.Jobs[0].Steps, 1)
}

func TestParseAction_ghaMap(t *testing.T) {
	t.Parallel()

	content := `
name: GHA style
on:
  push:
    branches: [main]
jobs:
  build:
    steps:
      - name: Build App
        run: echo "build app ..."
  deploy:
    steps:
      - name: Deploy App
        run: echo "deploying app ..."
`
	action, err := ParseAction(content)
	require.NoError(t, err)
	require.Len(t, action.Jobs, 2)
	require.Equal(t, "build", action.Jobs[0].Name)
	require.Equal(t, "deploy", action.Jobs[1].Name)
	require.Equal(t, "self-hosted", action.Jobs[0].RunsOn)
}

func TestParseAction_listOfSingleKeyMaps(t *testing.T) {
	t.Parallel()

	content := `
name: Legacy sample
on:
  push:
    branches: [main]
jobs:
  - build:
      steps:
        - name: Build App
          run: echo "build app ..."
  - deploy:
      steps:
        - name: Deploy App
          run: echo "deploying app ..."
`
	action, err := ParseAction(content)
	require.NoError(t, err)
	require.Len(t, action.Jobs, 2)
	require.Equal(t, "build", action.Jobs[0].Name)
	require.Equal(t, "deploy", action.Jobs[1].Name)
}
