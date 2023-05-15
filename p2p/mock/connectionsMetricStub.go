package mock

import "github.com/libp2p/go-libp2p/core"

// ConnectionsMetricStub -
type ConnectionsMetricStub struct {
	ResetNumConnectionsCalled    func() uint32
	ResetNumDisconnectionsCalled func() uint32
	ListenCalled                 func(core.Network, core.Multiaddr)
	ListenCloseCalled            func(core.Network, core.Multiaddr)
	ConnectedCalled              func(core.Network, core.Conn)
	DisconnectedCalled           func(core.Network, core.Conn)
}

// ResetNumConnections -
func (stub *ConnectionsMetricStub) ResetNumConnections() uint32 {
	if stub.ResetNumConnectionsCalled != nil {
		return stub.ResetNumConnectionsCalled()
	}
	return 0
}

// ResetNumDisconnections -
func (stub *ConnectionsMetricStub) ResetNumDisconnections() uint32 {
	if stub.ResetNumDisconnectionsCalled != nil {
		return stub.ResetNumDisconnectionsCalled()
	}
	return 0
}

// Listen -
func (stub *ConnectionsMetricStub) Listen(network core.Network, addr core.Multiaddr) {
	if stub.ListenCalled != nil {
		stub.ListenCalled(network, addr)
	}
}

// ListenClose -
func (stub *ConnectionsMetricStub) ListenClose(network core.Network, addr core.Multiaddr) {
	if stub.ListenCloseCalled != nil {
		stub.ListenCloseCalled(network, addr)
	}
}

// Connected -
func (stub *ConnectionsMetricStub) Connected(network core.Network, conn core.Conn) {
	if stub.ConnectedCalled != nil {
		stub.ConnectedCalled(network, conn)
	}
}

// Disconnected -
func (stub *ConnectionsMetricStub) Disconnected(network core.Network, conn core.Conn) {
	if stub.DisconnectedCalled != nil {
		stub.DisconnectedCalled(network, conn)
	}
}

// IsInterfaceNil -
func (stub *ConnectionsMetricStub) IsInterfaceNil() bool {
	return stub == nil
}
