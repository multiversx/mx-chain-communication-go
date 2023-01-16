package mock

import "github.com/libp2p/go-libp2p/core/peer"

// IDProviderStub -
type IDProviderStub struct {
	IDCalled func() peer.ID
}

// ID -
func (stub *IDProviderStub) ID() peer.ID {
	if stub.IDCalled != nil {
		return stub.IDCalled()
	}
	return ""
}

// IsInterfaceNil -
func (stub *IDProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
