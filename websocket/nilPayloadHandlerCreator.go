package websocket

type nilPayloadHandlerCreator struct{}

func NewNilPayloadHandlerCreator() *nilPayloadHandlerCreator {
	return new(nilPayloadHandlerCreator)
}

// Create will create a new instance of nilPayloadHandler
func (phc *nilPayloadHandlerCreator) Create() (PayloadHandler, error) {
	return NewNilPayloadHandler(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (phc *nilPayloadHandlerCreator) IsInterfaceNil() bool {
	return phc == nil
}
