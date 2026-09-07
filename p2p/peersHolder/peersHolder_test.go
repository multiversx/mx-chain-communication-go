package peersHolder

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewPeersHolder(t *testing.T) {
	t.Parallel()

	t.Run("invalid addresses should error", func(t *testing.T) {
		t.Parallel()

		preferredPeers := []string{"10.100.100", "invalid string"}
		ph, err := NewPeersHolder(preferredPeers)
		assert.True(t, check.IfNil(ph))
		assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ph, _ := NewPeersHolder([]string{"10.100.100.100"})
		assert.False(t, check.IfNil(ph))
	})
}

func TestPeersHolder_PutConnectionAddress(t *testing.T) {
	t.Parallel()

	t.Run("not preferred should not add", func(t *testing.T) {
		t.Parallel()

		preferredPeers := []string{"10.100.100.100"}
		ph, _ := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		unknownConnection := "/ip4/20.200.200.200/tcp/8080/p2p/some-random-pid" // preferredPeers[0]
		providedPid := core.PeerID("provided pid")
		ph.PutConnectionAddress(providedPid, unknownConnection)

		_, found := ph.tempPeerIDsWaitingForShard[providedPid]
		assert.False(t, found)

		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
	t.Run("new connection should add to intermediate maps", func(t *testing.T) {
		t.Parallel()

		preferredPeers := []string{"10.100.100.100", "10.100.100.101"}
		ph, _ := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		newConnection := "/ip4/10.100.100.100/tcp/38191/p2p/some-random-pid" // preferredPeers[0]
		providedPid := core.PeerID("provided pid")
		ph.PutConnectionAddress(providedPid, newConnection)

		knownConnection, found := ph.tempPeerIDsWaitingForShard[providedPid]
		assert.True(t, found)
		assert.Equal(t, preferredPeers[0], knownConnection)

		peersInfo := ph.connAddrToPeersInfo[knownConnection]
		assert.Equal(t, 1, len(peersInfo))
		assert.Equal(t, providedPid, peersInfo[0].pid)
		assert.Equal(t, core.AllShardId, peersInfo[0].shardID)

		// not in the final map yet
		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
	t.Run("should save second pid on same address", func(t *testing.T) {
		t.Parallel()

		preferredPeers := []string{"10.100.100.100", "10.100.100.101", "16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ"}
		ph, _ := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		newConnection := "/ip4/10.100.100.102/tcp/38191/p2p/16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ" // preferredPeers[2]
		providedPid := core.PeerID("provided pid")
		ph.PutConnectionAddress(providedPid, newConnection)

		knownConnection, found := ph.tempPeerIDsWaitingForShard[providedPid]
		assert.True(t, found)
		assert.Equal(t, preferredPeers[2], knownConnection)

		peersInfo := ph.connAddrToPeersInfo[knownConnection]
		assert.Equal(t, 1, len(peersInfo))
		assert.Equal(t, providedPid, peersInfo[0].pid)
		assert.Equal(t, core.AllShardId, peersInfo[0].shardID)

		ph.PutConnectionAddress(providedPid, newConnection) // try to update with same connection for coverage

		newPid := core.PeerID("new pid")
		ph.PutConnectionAddress(newPid, newConnection)
		knownConnection, found = ph.tempPeerIDsWaitingForShard[providedPid]
		assert.True(t, found)
		assert.Equal(t, preferredPeers[2], knownConnection)

		peersInfo = ph.connAddrToPeersInfo[knownConnection]
		assert.Equal(t, 2, len(peersInfo))
		assert.Equal(t, newPid, peersInfo[1].pid)
		assert.Equal(t, core.AllShardId, peersInfo[1].shardID)

		// not in the final map yet
		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
}

func TestPeersHolder_PutShardID(t *testing.T) {
	t.Parallel()

	t.Run("peer not added in the waiting list should be skipped", func(t *testing.T) {
		t.Parallel()

		preferredPeers := []string{"10.100.100.100"}
		ph, _ := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		providedPid := core.PeerID("provided pid")
		providedShardID := uint32(123)
		ph.PutShardID(providedPid, providedShardID)

		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
	t.Run("peer not added in map should be skipped", func(t *testing.T) {
		t.Parallel()

		preferredPeers := []string{"10.100.100.100"}
		ph, _ := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		providedPid := core.PeerID("provided pid")
		providedShardID := uint32(123)
		ph.tempPeerIDsWaitingForShard[providedPid] = preferredPeers[0]
		ph.PutShardID(providedPid, providedShardID)

		peers := ph.Get()
		assert.Equal(t, 0, len(peers))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		preferredPeers := []string{"10.100.100.100", "10.100.100.101", "16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ"}
		ph, _ := NewPeersHolder(preferredPeers)
		assert.False(t, check.IfNil(ph))

		newConnection := "/ip4/10.100.100.101/tcp/38191/p2p/some-random-pid" // preferredPeers[1]
		providedPid := core.PeerID("provided pid")
		ph.PutConnectionAddress(providedPid, newConnection)

		providedShardID := uint32(123)
		ph.PutShardID(providedPid, providedShardID)

		peers := ph.Get()
		assert.Equal(t, 1, len(peers))
		peersInShard, found := peers[providedShardID]
		assert.True(t, found)
		assert.Equal(t, providedPid, peersInShard[0])

		pidData := ph.peerIDs[providedPid]
		assert.Equal(t, preferredPeers[1], pidData.connectionAddress)
		assert.Equal(t, providedShardID, pidData.shardID)
		assert.Equal(t, 0, pidData.index)

		_, found = ph.tempPeerIDsWaitingForShard[providedPid]
		assert.False(t, found)
	})
}

func TestPeersHolder_Contains(t *testing.T) {
	t.Parallel()

	preferredPeers := []string{"10.100.100.100", "10.100.100.101"}
	ph, _ := NewPeersHolder(preferredPeers)
	assert.False(t, check.IfNil(ph))

	newConnection := "/ip4/10.100.100.101/tcp/38191/p2p/some-random-pid" // preferredPeers[1]
	providedPid := core.PeerID("provided pid")
	ph.PutConnectionAddress(providedPid, newConnection)

	providedShardID := uint32(123)
	ph.PutShardID(providedPid, providedShardID)

	assert.True(t, ph.Contains(providedPid))

	ph.Remove(providedPid)
	assert.False(t, ph.Contains(providedPid))

	unknownPid := core.PeerID("unknown pid")
	ph.Remove(unknownPid) // for code coverage
}

func TestPeersHolder_Remove_UpdatesIndexesOfRemainingPeers(t *testing.T) {
	t.Parallel()

	preferredPeers := []string{"10.100.100.100", "10.100.100.101", "10.100.100.102"}
	ph, _ := NewPeersHolder(preferredPeers)
	assert.False(t, check.IfNil(ph))

	shardID := uint32(1)
	pid1 := core.PeerID("pid1")
	pid2 := core.PeerID("pid2")
	pid3 := core.PeerID("pid3")

	conn1 := "/ip4/10.100.100.100/tcp/38191/p2p/some-random-pid" // preferredPeers[0]
	conn2 := "/ip4/10.100.100.101/tcp/38191/p2p/some-random-pid" // preferredPeers[1]
	conn3 := "/ip4/10.100.100.102/tcp/38191/p2p/some-random-pid" // preferredPeers[2]

	ph.PutConnectionAddress(pid1, conn1)
	ph.PutShardID(pid1, shardID)
	ph.PutConnectionAddress(pid2, conn2)
	ph.PutShardID(pid2, shardID)
	ph.PutConnectionAddress(pid3, conn3)
	ph.PutShardID(pid3, shardID)

	assert.Equal(t, 0, ph.peerIDs[pid1].index)
	assert.Equal(t, 1, ph.peerIDs[pid2].index)
	assert.Equal(t, 2, ph.peerIDs[pid3].index)

	ph.Remove(pid2) // remove the middle peer

	assert.Equal(t, 0, ph.peerIDs[pid1].index)
	assert.Equal(t, 1, ph.peerIDs[pid3].index)

	peers := ph.Get()
	assert.Equal(t, 2, len(peers[shardID]))
	assert.Equal(t, pid1, peers[shardID][0])
	assert.Equal(t, pid3, peers[shardID][1])

	// removing the remaining peers afterwards should still work and yield the correct set
	ph.Remove(pid3)
	peers = ph.Get()
	assert.Equal(t, 1, len(peers[shardID]))
	assert.Equal(t, pid1, peers[shardID][0])

	ph.Remove(pid1)
	peers = ph.Get()
	assert.Equal(t, 0, len(peers[shardID]))
}

func TestPeersHolder_Clear(t *testing.T) {
	t.Parallel()

	preferredPeers := []string{"10.100.100.100", "16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ"}
	ph, _ := NewPeersHolder(preferredPeers)
	assert.False(t, check.IfNil(ph))

	newConnection1 := "/ip4/10.100.100.100/tcp/38191/p2p/some-random-pid" // preferredPeers[0]
	providedPid1 := core.PeerID("provided pid 1")
	ph.PutConnectionAddress(providedPid1, newConnection1)
	providedShardID := uint32(123)
	ph.PutShardID(providedPid1, providedShardID)
	assert.True(t, ph.Contains(providedPid1))

	newConnection2 := "/ip4/10.100.100.102/tcp/38191/p2p/16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ" // preferredPeers[1]
	providedPid2 := core.PeerID("provided pid 2")
	ph.PutConnectionAddress(providedPid2, newConnection2)
	ph.PutShardID(providedPid2, providedShardID)
	assert.True(t, ph.Contains(providedPid2))

	peers := ph.Get()
	assert.Equal(t, 1, len(peers))
	assert.Equal(t, 2, len(peers[providedShardID]))

	ph.Clear()
	peers = ph.Get()
	assert.Equal(t, 0, len(peers))
}
