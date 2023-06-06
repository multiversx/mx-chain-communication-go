package crypto

import (
	"crypto/rand"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var emptyPrivateKeyBytes = []byte("")

const loggerName = "p2p/libp2p/crypto"

var log = logger.GetOrCreate(loggerName)

type identityGenerator struct {
}

// NewIdentityGenerator creates a new identity generator
func NewIdentityGenerator(loggerPrefix string) *identityGenerator {
	if len(loggerPrefix) > 0 {
		log = logger.GetOrCreate(loggerPrefix + loggerName)
	}

	return &identityGenerator{}
}

// CreateRandomP2PIdentity creates a valid random p2p identity to sign messages on the behalf of other identity
func (generator *identityGenerator) CreateRandomP2PIdentity() ([]byte, core.PeerID, error) {
	sk, err := generator.CreateP2PPrivateKey(emptyPrivateKeyBytes)
	if err != nil {
		return nil, "", err
	}

	skBuff, err := sk.Raw()
	if err != nil {
		return nil, "", err
	}

	pid, err := peer.IDFromPublicKey(sk.GetPublic())
	if err != nil {
		return nil, "", err
	}

	return skBuff, core.PeerID(pid), nil
}

// CreateP2PPrivateKey will create a new P2P private key based on the provided private key bytes. If the byte slice is empty,
// it will use the crypto's random generator to provide a random one. Otherwise, it will try to load the private key from
// the provided bytes (if the bytes are in the correct format).
// This is useful when we want a private key that never changes, such as in the network seeders
func (generator *identityGenerator) CreateP2PPrivateKey(privateKeyBytes []byte) (libp2pCrypto.PrivKey, error) {
	if len(privateKeyBytes) == 0 {
		randReader := rand.Reader
		prvKey, _, err := libp2pCrypto.GenerateSecp256k1Key(randReader)
		if err != nil {
			return nil, err
		}

		log.Info("createP2PPrivateKey: generated a new private key for p2p signing")

		return prvKey, nil
	}

	prvKey, err := libp2pCrypto.UnmarshalSecp256k1PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	log.Info("createP2PPrivateKey: using the provided private key for p2p signing")

	return prvKey, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (generator *identityGenerator) IsInterfaceNil() bool {
	return generator == nil
}
