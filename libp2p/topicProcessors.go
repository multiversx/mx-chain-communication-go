package libp2p

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-p2p-go"
)

type topicProcessors struct {
	processors    map[string]p2p.MessageProcessor
	mutProcessors sync.RWMutex
}

func newTopicProcessors() *topicProcessors {
	return &topicProcessors{
		processors: make(map[string]p2p.MessageProcessor),
	}
}

// AddTopicProcessor adds a new topic processor for the provided identifier
func (tp *topicProcessors) AddTopicProcessor(identifier string, processor p2p.MessageProcessor) error {
	tp.mutProcessors.Lock()
	defer tp.mutProcessors.Unlock()

	_, alreadyExists := tp.processors[identifier]
	if alreadyExists {
		return fmt.Errorf("%w, in addTopicProcessor, identifier %s",
			p2p.ErrMessageProcessorAlreadyDefined,
			identifier,
		)
	}

	tp.processors[identifier] = processor

	return nil
}

// RemoveTopicProcessor removes the topic processor for the provided identifier
func (tp *topicProcessors) RemoveTopicProcessor(identifier string) error {
	tp.mutProcessors.Lock()
	defer tp.mutProcessors.Unlock()

	_, alreadyExists := tp.processors[identifier]
	if !alreadyExists {
		return fmt.Errorf("%w, in removeTopicProcessor, identifier %s",
			p2p.ErrMessageProcessorDoesNotExists,
			identifier,
		)
	}

	delete(tp.processors, identifier)

	return nil
}

// GetList returns the list of identifiers and processors
func (tp *topicProcessors) GetList() ([]string, []p2p.MessageProcessor) {
	tp.mutProcessors.RLock()
	defer tp.mutProcessors.RUnlock()

	list := make([]p2p.MessageProcessor, 0, len(tp.processors))
	identifiers := make([]string, 0, len(tp.processors))

	for identifier, handler := range tp.processors {
		list = append(list, handler)
		identifiers = append(identifiers, identifier)
	}

	return identifiers, list
}

// IsInterfaceNil returns true if there is no value under the interface
func (tp *topicProcessors) IsInterfaceNil() bool {
	return tp == nil
}
