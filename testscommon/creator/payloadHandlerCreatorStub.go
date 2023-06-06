package creator

import "github.com/multiversx/mx-chain-communication-go/websocket"

// PayloadHandlerCreatorStub -
type PayloadHandlerCreatorStub struct {
	CreateCalled func() (websocket.PayloadHandler, error)
}

// Create -
func (pcs *PayloadHandlerCreatorStub) Create() (websocket.PayloadHandler, error) {
	if pcs.CreateCalled != nil {
		return pcs.CreateCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (pcs *PayloadHandlerCreatorStub) IsInterfaceNil() bool {
	return pcs == nil
}
