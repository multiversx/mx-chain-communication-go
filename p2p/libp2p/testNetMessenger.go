package libp2p

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-core-go/core"
)

type testNetMessenger struct {
	*networkMessenger
}

// NewTestNetMessenger creates a wrapper over networkMessenger that exposes more functionality
func NewTestNetMessenger(args ArgsNetworkMessenger) (*testNetMessenger, error) {
	netMessenger, err := NewNetworkMessenger(args)
	if err != nil {
		return nil, err
	}

	return &testNetMessenger{
		networkMessenger: netMessenger,
	}, nil
}

// ClosePeer tries to close the provided peer ID
func (netMes *testNetMessenger) ClosePeer(pid core.PeerID) error {
	return netMes.p2pHost.Network().ClosePeer(peer.ID(pid))
}
