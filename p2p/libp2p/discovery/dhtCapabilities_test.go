package discovery_test

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	libp2pHost "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dhtPb "github.com/libp2p/go-libp2p-kad-dht/pb"
	record "github.com/libp2p/go-libp2p-record"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/discovery"
)

const kadProtocolSuffix = "/kad/1.0.0"

type testConnectableHost struct {
	host.Host
}

func (tch *testConnectableHost) ConnectToPeer(ctx context.Context, address string) error {
	peerInfo, err := tch.AddressToPeerInfo(address)
	if err != nil {
		return err
	}

	return tch.Connect(ctx, *peerInfo)
}

func (tch *testConnectableHost) AddressToPeerInfo(address string) (*peer.AddrInfo, error) {
	peerAddress, err := multiaddr.NewMultiaddr(address)
	if err != nil {
		return nil, err
	}

	return peer.AddrInfoFromP2pAddr(peerAddress)
}

func (tch *testConnectableHost) IsInterfaceNil() bool {
	return tch == nil
}

func TestContinuousKadDhtDiscoverer_RecordOperationsAreDisabled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	arg := createTestArgument()
	arg.Context = ctx
	arg.InitialPeersList = nil
	discoverer, err := discovery.NewContinuousKadDhtDiscoverer(arg)
	require.NoError(t, err)
	require.NoError(t, discoverer.Bootstrap())
	defer func() {
		require.NoError(t, discoverer.StopDHT())
	}()

	assertRecordOperationsDisabled(t, discoverer.KadDHT())
}

func TestOptimizedKadDhtDiscoverer_RecordOperationsAreDisabled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	arg := createTestArgument()
	arg.Context = ctx
	arg.InitialPeersList = nil
	discoverer, err := discovery.NewOptimizedKadDhtDiscoverer(arg)
	require.NoError(t, err)
	require.NoError(t, discoverer.Bootstrap())

	kadDHT := discoverer.KadDHT()
	defer func() {
		require.NoError(t, kadDHT.Close())
	}()

	assertRecordOperationsDisabled(t, kadDHT)
}

func TestContinuousKadDhtDiscoverer_RecordMessagesAreRejectedAndRoutingRemainsAvailable(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	targetHost := newTestConnectableHost(t)
	remoteHost, err := libp2pHost.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, remoteHost.Close())
	})

	arg := createTestArgument()
	arg.Context = ctx
	arg.Host = targetHost
	arg.InitialPeersList = nil
	discoverer, err := discovery.NewContinuousKadDhtDiscoverer(arg)
	require.NoError(t, err)
	require.NoError(t, discoverer.Bootstrap())
	t.Cleanup(func() {
		require.NoError(t, discoverer.StopDHT())
	})

	assertRecordMessagesRejectedAndRoutingAvailable(t, remoteHost, targetHost.Host, arg.ProtocolID)
}

func TestOptimizedKadDhtDiscoverer_RecordMessagesAreRejectedAndRoutingRemainsAvailable(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	targetHost := newTestConnectableHost(t)
	remoteHost, err := libp2pHost.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, remoteHost.Close())
	})

	arg := createTestArgument()
	arg.Context = ctx
	arg.Host = targetHost
	arg.InitialPeersList = nil
	discoverer, err := discovery.NewOptimizedKadDhtDiscoverer(arg)
	require.NoError(t, err)
	require.NoError(t, discoverer.Bootstrap())
	t.Cleanup(func() {
		require.NoError(t, discoverer.KadDHT().Close())
	})

	assertRecordMessagesRejectedAndRoutingAvailable(t, remoteHost, targetHost.Host, arg.ProtocolID)
}

func assertRecordOperationsDisabled(t *testing.T, kadDHT *dht.IpfsDHT) {
	t.Helper()
	require.NotNil(t, kadDHT)

	err := kadDHT.PutValue(context.Background(), "/pk/test", []byte("value"))
	require.ErrorIs(t, err, routing.ErrNotSupported)

	_, err = kadDHT.GetValue(context.Background(), "/pk/test")
	require.ErrorIs(t, err, routing.ErrNotSupported)

	err = kadDHT.Provide(context.Background(), cid.Undef, true)
	require.ErrorIs(t, err, routing.ErrNotSupported)

	_, err = kadDHT.FindProviders(context.Background(), cid.Undef)
	require.ErrorIs(t, err, routing.ErrNotSupported)
}

func newTestConnectableHost(t *testing.T) *testConnectableHost {
	t.Helper()

	libp2pHostInstance, err := libp2pHost.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, libp2pHostInstance.Close())
	})

	return &testConnectableHost{Host: libp2pHostInstance}
}

func assertRecordMessagesRejectedAndRoutingAvailable(
	t *testing.T,
	remoteHost host.Host,
	targetHost host.Host,
	protocolPrefix string,
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	targetInfo := peer.AddrInfo{
		ID:    targetHost.ID(),
		Addrs: targetHost.Addrs(),
	}
	require.NoError(t, remoteHost.Connect(ctx, targetInfo))

	dhtProtocol := protocol.ID(protocolPrefix + kadProtocolSuffix)
	assertDHTMessageRejected(t, remoteHost, targetHost.ID(), dhtProtocol, createValidPutValueMessage(t))
	assertDHTMessageRejected(t, remoteHost, targetHost.ID(), dhtProtocol, dhtPb.NewMessage(dhtPb.Message_ADD_PROVIDER, []byte("provider-key"), 0))

	stream, err := remoteHost.NewStream(ctx, targetHost.ID(), dhtProtocol)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = stream.Close()
	})
	require.NoError(t, stream.SetDeadline(time.Now().Add(5*time.Second)))
	require.NoError(t, pbio.NewDelimitedWriter(stream).WriteMsg(
		dhtPb.NewMessage(dhtPb.Message_FIND_NODE, []byte(targetHost.ID()), 0),
	))

	response := &dhtPb.Message{}
	require.NoError(t, pbio.NewDelimitedReader(stream, 1<<20).ReadMsg(response))
	require.Equal(t, dhtPb.Message_FIND_NODE, response.GetType())
}

func createValidPutValueMessage(t *testing.T) *dhtPb.Message {
	t.Helper()

	_, publicKey, err := libp2pCrypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)
	peerID, err := peer.IDFromPublicKey(publicKey)
	require.NoError(t, err)
	publicKeyBytes, err := libp2pCrypto.MarshalPublicKey(publicKey)
	require.NoError(t, err)
	key := routing.KeyForPublicKey(peerID)

	message := dhtPb.NewMessage(dhtPb.Message_PUT_VALUE, []byte(key), 0)
	message.Record = record.MakePutRecord(key, publicKeyBytes)
	return message
}

func assertDHTMessageRejected(
	t *testing.T,
	remoteHost host.Host,
	targetPeerID peer.ID,
	dhtProtocol protocol.ID,
	message *dhtPb.Message,
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := remoteHost.NewStream(ctx, targetPeerID, dhtProtocol)
	require.NoError(t, err)
	require.NoError(t, stream.SetDeadline(time.Now().Add(5*time.Second)))

	err = pbio.NewDelimitedWriter(stream).WriteMsg(message)
	if err == nil {
		err = pbio.NewDelimitedReader(stream, 1<<20).ReadMsg(&dhtPb.Message{})
	}
	require.ErrorIs(t, err, network.ErrReset)
	_ = stream.Close()
}
