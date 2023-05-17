package factory

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-p2p-go"
	"github.com/multiversx/mx-chain-p2p-go/libp2p/metrics"
)

// NewConnectionsWatcher creates a new ConnectionWatcher instance based on the input parameters
func NewConnectionsWatcher(connectionsWatcherType string, timeToLive time.Duration) (p2p.ConnectionsWatcher, error) {
	switch connectionsWatcherType {
	case p2p.ConnectionWatcherTypePrint:
		return metrics.NewPrintConnectionsWatcher(timeToLive)
	case p2p.ConnectionWatcherTypeDisabled, p2p.ConnectionWatcherTypeEmpty:
		return metrics.NewDisabledConnectionsWatcher(), nil
	default:
		return nil, fmt.Errorf("%w %s", ErrUnknownConnectionWatcherType, connectionsWatcherType)
	}
}
