package factory

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-p2p/common"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p/metrics"
)

// NewConnectionsWatcher creates a new ConnectionWatcher instance based on the input parameters
func NewConnectionsWatcher(connectionsWatcherType string, timeToLive time.Duration) (common.ConnectionsWatcher, error) {
	switch connectionsWatcherType {
	case common.ConnectionWatcherTypePrint:
		return metrics.NewPrintConnectionsWatcher(timeToLive)
	case common.ConnectionWatcherTypeDisabled, common.ConnectionWatcherTypeEmpty:
		return metrics.NewDisabledConnectionsWatcher(), nil
	default:
		return nil, fmt.Errorf("%w %s", errUnknownConnectionWatcherType, connectionsWatcherType)
	}
}
