package ui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatWSStreamLine_deploymentLogEnvelope(t *testing.T) {
	t.Parallel()

	raw := `{"event":"deployment-log-stream-event","content":{"id":"log-1","deploymentId":"dep-1","jobId":"build.Deploy","message":"$ git clone https://github.com/org/repo.git code"}}`
	require.Equal(t, "$ git clone https://github.com/org/repo.git code", formatWSStreamLine(raw))
}

func TestFormatWSStreamLine_nestedJSONStringContent(t *testing.T) {
	t.Parallel()

	raw := `{"event":"deployment-log-stream-event","content":"{\"id\":\"log-2\",\"message\":\"Storm is Ready!\"}"}`
	require.Equal(t, "Storm is Ready!", formatWSStreamLine(raw))
}

func TestFormatWSStreamLine_plainMessagePassthrough(t *testing.T) {
	t.Parallel()

	require.Equal(t, "websocket closed: EOF", formatWSStreamLine("websocket closed: EOF"))
}
