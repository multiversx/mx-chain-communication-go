package crypto

import (
	"fmt"

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

// ConvertPrivateKeyToPeerID will convert common private key to core peer id
func ConvertPrivateKeyToPeerID(privateKey crypto.PrivateKey) (core.PeerID, error) {
	pk := privateKey.GeneratePublic()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return "", err
	}

	pubKey, err := libp2pCrypto.UnmarshalSecp256k1PublicKey(pkBytes)
	if err != nil {
		return "", err
	}

	pid, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return "", err
	}

	return core.PeerID(pid), nil
}

// ConvertPeerIDToPublicKey will convert core peer id to common public key
func ConvertPeerIDToPublicKey(keyGen crypto.KeyGenerator, pid core.PeerID) (crypto.PublicKey, error) {
	libp2pPid, err := peer.IDFromBytes(pid.Bytes())
	if err != nil {
		return nil, err
	}

	pubk, err := libp2pPid.ExtractPublicKey()
	if err != nil {
		return nil, fmt.Errorf("cannot extract signing key: %s", err.Error())
	}

	pubKeyBytes, err := pubk.Raw()
	if err != nil {
		return nil, err
	}

	return keyGen.PublicKeyFromByteArray(pubKeyBytes)
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
