package libp2p_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p"
	"github.com/stretchr/testify/assert"
)

func TestUnknownPeerShardResolver_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	upsr := libp2p.NewUnknownPeerShardResolver()
	assert.False(t, check.IfNil(upsr))
}

func TestUnknownPeerShardResolver_GetPeerInfoShouldReturnUnknownId(t *testing.T) {
	t.Parallel()

	upsr := libp2p.NewUnknownPeerShardResolver()
	expectedPeerInfo := core.P2PPeerInfo{
		PeerType: core.UnknownPeer,
		ShardID:  0,
	}

	assert.Equal(t, expectedPeerInfo, upsr.GetPeerInfo(""))
}
