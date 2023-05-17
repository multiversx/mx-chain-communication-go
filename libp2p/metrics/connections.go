package metrics

import (
	"sync/atomic"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

// connectionsMetric is a metric that counts connections and disconnections done by the host implementation
type connectionsMetric struct {
	numConnections    uint32
	numDisconnections uint32
}

// NewConnectionsMetric returns a new connectionsMetric instance
func NewConnectionsMetric() *connectionsMetric {
	return &connectionsMetric{
		numConnections:    0,
		numDisconnections: 0,
	}
}

// Listen is called when network starts listening on an addr
func (cm *connectionsMetric) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (cm *connectionsMetric) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected is called when a connection opened. It increments the numConnections counter
func (cm *connectionsMetric) Connected(network.Network, network.Conn) {
	atomic.AddUint32(&cm.numConnections, 1)
}

// Disconnected is called when a connection closed it increments the numDisconnections counter
func (cm *connectionsMetric) Disconnected(network.Network, network.Conn) {
	atomic.AddUint32(&cm.numDisconnections, 1)
}

// ResetNumConnections resets the numConnections counter returning the previous value
func (cm *connectionsMetric) ResetNumConnections() uint32 {
	return atomic.SwapUint32(&cm.numConnections, 0)
}

// ResetNumDisconnections resets the numDisconnections counter returning the previous value
func (cm *connectionsMetric) ResetNumDisconnections() uint32 {
	return atomic.SwapUint32(&cm.numDisconnections, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (cm *connectionsMetric) IsInterfaceNil() bool {
	return cm == nil
}
