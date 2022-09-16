package factory

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-p2p/common"
	"github.com/stretchr/testify/assert"
)

func TestNewConnectionsWatcher(t *testing.T) {
	t.Parallel()

	t.Run("print connections watcher", func(t *testing.T) {
		t.Parallel()

		cw, err := NewConnectionsWatcher(common.ConnectionWatcherTypePrint, time.Second)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(cw))
		assert.Equal(t, "*metrics.printConnectionsWatcher", fmt.Sprintf("%T", cw))
	})
	t.Run("disabled connections watcher", func(t *testing.T) {
		t.Parallel()

		cw, err := NewConnectionsWatcher(common.ConnectionWatcherTypeDisabled, time.Second)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(cw))
		assert.Equal(t, "*metrics.disabledConnectionsWatcher", fmt.Sprintf("%T", cw))
	})
	t.Run("empty connections watcher", func(t *testing.T) {
		t.Parallel()

		cw, err := NewConnectionsWatcher(common.ConnectionWatcherTypeEmpty, time.Second)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(cw))
		assert.Equal(t, "*metrics.disabledConnectionsWatcher", fmt.Sprintf("%T", cw))
	})
	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()

		cw, err := NewConnectionsWatcher("unknown", time.Second)
		assert.True(t, errors.Is(err, errUnknownConnectionWatcherType))
		assert.True(t, check.IfNil(cw))
	})
}
