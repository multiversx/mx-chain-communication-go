package libp2p

import (
	"context"
	"crypto/rand"
	"fmt"

	ic "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoremem"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/crypto"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/metrics/factory"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
)

// genIPv4Peer will generate a host connected to the provided mocknet (IPv4)
func genIPv4Peer(mockNet mocknet.Mocknet) (host.Host, error) {
	sk, _, err := ic.GenerateECDSAKeyPair(rand.Reader)
	if err != nil {
		return nil, err
	}
	a, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/4242"))
	if err != nil {
		return nil, fmt.Errorf("failed to create test multiaddr: %s", err)
	}

	ps, err := pstoremem.NewPeerstore()
	if err != nil {
		return nil, err
	}

	p, err := updatePeerstore(sk, a, ps)
	if err != nil {
		return nil, err
	}
	h, err := mockNet.AddPeerWithPeerstore(p, ps)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func updatePeerstore(k ic.PrivKey, a ma.Multiaddr, ps peerstore.Peerstore) (peer.ID, error) {
	p, err := peer.IDFromPublicKey(k.GetPublic())
	if err != nil {
		return "", err
	}

	ps.AddAddr(p, a, peerstore.PermanentAddrTTL)
	err = ps.AddPrivKey(p, k)
	if err != nil {
		return "", err
	}
	err = ps.AddPubKey(p, k.GetPublic())
	if err != nil {
		return "", err
	}
	return p, nil
}

// NewMockMessenger creates a new sandbox testable instance of libP2P messenger
// It should not open ports on current machine
// Should be used only in testing!
func NewMockMessenger(
	args ArgsNetworkMessenger,
	mockNet mocknet.Mocknet,
) (*networkMessenger, error) {
	if mockNet == nil {
		return nil, p2p.ErrNilMockNet
	}

	h, err := genIPv4Peer(mockNet)
	// TODO: use mockNet.GenPeer() instead (and remove genIPV4Peer function) when the saving of ipv6 mocked addresses in peerstore is fixed
	//h, err := mockNet.GenPeer()
	if err != nil {
		return nil, err
	}

	p2pSignerArgs := crypto.ArgsP2pSignerWrapper{
		PrivateKey:      args.P2pPrivateKey,
		Signer:          args.P2pSingleSigner,
		KeyGen:          args.P2pKeyGenerator,
		P2PKeyConverter: crypto.NewP2PKeyConverter(),
	}

	signer, err := crypto.NewP2PSignerWrapper(p2pSignerArgs)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	p2pNode := &networkMessenger{
		p2pSigner:  signer,
		p2pHost:    NewConnectableHost(h),
		ctx:        ctx,
		cancelFunc: cancelFunc,
		log:        args.Logger,
	}
	p2pNode.printConnectionsWatcher, err = factory.NewConnectionsWatcher(args.ConnectionWatcherType, ttlConnectionsWatcher, &testscommon.LoggerStub{})
	if err != nil {
		return nil, err
	}

	err = addComponentsToNode(args, p2pNode, withoutMessageSigning)
	if err != nil {
		return nil, err
	}

	return p2pNode, err
}
