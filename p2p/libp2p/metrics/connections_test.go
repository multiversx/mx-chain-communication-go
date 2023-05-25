package metrics_test

import (
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/metrics"
	"github.com/stretchr/testify/assert"
)

func TestConnectionsMetric_EmptyFunctionsDoNotPanicWhenCalled(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "test should not have failed")
		}
	}()

	cm := metrics.NewConnectionsMetric()

	cm.Listen(nil, nil)
	cm.ListenClose(nil, nil)
}

func TestConnectionsMetric_ResetNumConnectionsShouldWork(t *testing.T) {
	t.Parallel()

	cm := metrics.NewConnectionsMetric()

	cm.Connected(nil, nil)
	cm.Connected(nil, nil)

	existing := cm.ResetNumConnections()
	assert.Equal(t, uint32(2), existing)

	existing = cm.ResetNumConnections()
	assert.Equal(t, uint32(0), existing)
}

func TestConnectionsMetric_ResetNumDisconnectionsShouldWork(t *testing.T) {
	t.Parallel()

	cm := metrics.NewConnectionsMetric()

	cm.Disconnected(nil, nil)
	cm.Disconnected(nil, nil)

	existing := cm.ResetNumDisconnections()
	assert.Equal(t, uint32(2), existing)

	existing = cm.ResetNumDisconnections()
	assert.Equal(t, uint32(0), existing)
}
