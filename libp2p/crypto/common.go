package crypto

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/secp256k1"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

// ConvertPrivateKeyToLibp2pPrivateKey will convert common private key to libp2p private key
func ConvertPrivateKeyToLibp2pPrivateKey(privateKey crypto.PrivateKey) (libp2pCrypto.PrivKey, error) {
	if check.IfNil(privateKey) {
		return nil, ErrNilPrivateKey
	}

	p2pPrivateKeyBytes, err := privateKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	return libp2pCrypto.UnmarshalSecp256k1PrivateKey(p2pPrivateKeyBytes)
}

// CreateRandomP2PIdentity creates a random p2p crypto components
func CreateRandomP2PIdentity() ([]byte, core.PeerID, error) {
	keyGen := signing.NewKeyGenerator(secp256k1.NewSecp256k1())
	sk, _ := keyGen.GeneratePair()

	skBuff, err := sk.ToByteArray()
	if err != nil {
		return nil, "", err
	}

	pk := sk.GeneratePublic()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return nil, "", err
	}

	pubKey, err := libp2pCrypto.UnmarshalSecp256k1PublicKey(pkBytes)
	if err != nil {
		return nil, "", err
	}

	pid, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return nil, "", err
	}

	return skBuff, core.PeerID(pid), nil
}
