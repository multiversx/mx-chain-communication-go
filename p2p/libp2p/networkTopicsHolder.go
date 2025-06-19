package libp2p

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

type networkTopicsHolder struct {
	networkTopics map[string]p2p.NetworkType
	log           p2p.Logger
	mut           sync.RWMutex
	mainNetwork   p2p.NetworkType
}

func newNetworkTopicsHolder(log p2p.Logger, mainNetwork p2p.NetworkType) *networkTopicsHolder {
	return &networkTopicsHolder{
		networkTopics: make(map[string]p2p.NetworkType),
		log:           log,
		mainNetwork:   mainNetwork,
	}
}

// AddTopicOnNetworkIfNeeded saves a new topic for a specific network
func (holder *networkTopicsHolder) AddTopicOnNetworkIfNeeded(networkType p2p.NetworkType, topic string) {
	holder.mut.Lock()
	defer holder.mut.Unlock()

	_, found := holder.networkTopics[topic]
	if found {
		return
	}

	holder.networkTopics[topic] = networkType
}

// GetNetworkTypeForTopic returns the network type a topic lives on
func (holder *networkTopicsHolder) GetNetworkTypeForTopic(topic string) p2p.NetworkType {
	networkType, found := holder.networkTopics[topic]
	if !found {
		holder.log.Debug(
			fmt.Sprintf("p2p network not found for topic=%s, returning main network: %s",
				topic, holder.mainNetwork,
			),
		)

		return holder.mainNetwork
	}

	return networkType
}

// RemoveTopic removes a topic from the map
func (holder *networkTopicsHolder) RemoveTopic(topic string) {
	holder.mut.Lock()
	defer holder.mut.Unlock()

	delete(holder.networkTopics, topic)
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *networkTopicsHolder) IsInterfaceNil() bool {
	return holder == nil
}
