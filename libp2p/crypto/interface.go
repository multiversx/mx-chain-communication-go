package crypto

import (
	coreCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-crypto-go"
)

// P2PKeyConverter defines what a p2p key converter can do
type P2PKeyConverter interface {
	ConvertPrivateKeyToLibp2pPrivateKey(privateKey crypto.PrivateKey) (coreCrypto.PrivKey, error)
	ConvertPeerIDToPublicKey(keyGen crypto.KeyGenerator, pid core.PeerID) (crypto.PublicKey, error)
	ConvertPublicKeyToPeerID(pk crypto.PublicKey) (core.PeerID, error)
	IsInterfaceNil() bool
}
