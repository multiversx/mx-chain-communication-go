package libp2p

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-p2p/common"
)

type topicProcessors struct {
	processors    map[string]common.MessageProcessor
	mutProcessors sync.RWMutex
}

func newTopicProcessors() *topicProcessors {
	return &topicProcessors{
		processors: make(map[string]common.MessageProcessor),
	}
}

func (tp *topicProcessors) addTopicProcessor(identifier string, processor common.MessageProcessor) error {
	tp.mutProcessors.Lock()
	defer tp.mutProcessors.Unlock()

	_, alreadyExists := tp.processors[identifier]
	if alreadyExists {
		return fmt.Errorf("%w, in addTopicProcessor, identifier %s",
			common.ErrMessageProcessorAlreadyDefined,
			identifier,
		)
	}

	tp.processors[identifier] = processor

	return nil
}

func (tp *topicProcessors) removeTopicProcessor(identifier string) error {
	tp.mutProcessors.Lock()
	defer tp.mutProcessors.Unlock()

	_, alreadyExists := tp.processors[identifier]
	if !alreadyExists {
		return fmt.Errorf("%w, in removeTopicProcessor, identifier %s",
			common.ErrMessageProcessorDoesNotExists,
			identifier,
		)
	}

	delete(tp.processors, identifier)

	return nil
}

func (tp *topicProcessors) getList() ([]string, []common.MessageProcessor) {
	tp.mutProcessors.RLock()
	defer tp.mutProcessors.RUnlock()

	list := make([]common.MessageProcessor, 0, len(tp.processors))
	identifiers := make([]string, 0, len(tp.processors))

	for identifier, handler := range tp.processors {
		list = append(list, handler)
		identifiers = append(identifiers, identifier)
	}

	return identifiers, list
}
