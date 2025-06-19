package libp2p

import (
	"fmt"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-communication-go/p2p"
)

const requestTopicSuffix = "_REQUEST"

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

	baseTopic := strings.Split(topic, requestTopicSuffix)[0]

	_, found := holder.networkTopics[baseTopic]
	if found {
		return
	}

	holder.networkTopics[baseTopic] = networkType
}

// GetNetworkTypeForTopic returns the network type a topic lives on
func (holder *networkTopicsHolder) GetNetworkTypeForTopic(topic string) p2p.NetworkType {
	baseTopic := strings.Split(topic, requestTopicSuffix)[0]
	networkType, found := holder.networkTopics[baseTopic]
	if !found {
		holder.log.Warn(fmt.Sprintf("p2p network not found for baseTopic=%s, initial topic=%s, returning main network %s", baseTopic, topic, holder.mainNetwork))

		return holder.mainNetwork
	}

	return networkType
}

// RemoveTopic removes a topic from the map
func (holder *networkTopicsHolder) RemoveTopic(topic string) {
	holder.mut.Lock()
	defer holder.mut.Unlock()

	baseTopic := strings.Split(topic, requestTopicSuffix)[0]
	delete(holder.networkTopics, baseTopic)
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *networkTopicsHolder) IsInterfaceNil() bool {
	return holder == nil
}
