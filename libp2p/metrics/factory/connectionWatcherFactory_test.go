package factory_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/mutliversx/mx-chain-p2p-go"
	"github.com/mutliversx/mx-chain-p2p-go/libp2p/metrics/factory"
	"github.com/stretchr/testify/assert"
)

func TestNewConnectionsWatcher(t *testing.T) {
	t.Parallel()

	t.Run("print connections watcher", func(t *testing.T) {
		t.Parallel()

		cw, err := factory.NewConnectionsWatcher(p2p.ConnectionWatcherTypePrint, time.Second)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(cw))
		assert.Equal(t, "*metrics.printConnectionsWatcher", fmt.Sprintf("%T", cw))
	})
	t.Run("disabled connections watcher", func(t *testing.T) {
		t.Parallel()

		cw, err := factory.NewConnectionsWatcher(p2p.ConnectionWatcherTypeDisabled, time.Second)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(cw))
		assert.Equal(t, "*metrics.disabledConnectionsWatcher", fmt.Sprintf("%T", cw))
	})
	t.Run("empty connections watcher", func(t *testing.T) {
		t.Parallel()

		cw, err := factory.NewConnectionsWatcher(p2p.ConnectionWatcherTypeEmpty, time.Second)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(cw))
		assert.Equal(t, "*metrics.disabledConnectionsWatcher", fmt.Sprintf("%T", cw))
	})
	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()

		cw, err := factory.NewConnectionsWatcher("unknown", time.Second)
		assert.True(t, errors.Is(err, factory.ErrUnknownConnectionWatcherType))
		assert.True(t, check.IfNil(cw))
	})
}
