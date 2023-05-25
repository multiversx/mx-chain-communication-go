package networksharding_test

import (
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/networksharding"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNilListSharderSharder(t *testing.T) {
	nls := networksharding.NewNilListSharder()

	assert.False(t, check.IfNil(nls))
	assert.Equal(t, 0, len(nls.ComputeEvictionList(nil)))
	assert.False(t, nls.Has("", nil))
	assert.Nil(t, nls.SetPeerShardResolver(nil))
}
