package mock

import (
	"crypto/ecdsa"
	"crypto/rand"

	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/btcsuite/btcd/btcec"
)

// PrivateKeyMock implements common PrivateKey interface
type PrivateKeyMock struct {
	privateKey *ecdsa.PrivateKey
}

// NewPrivateKeyMock will create a new PrivateKeyMock instance
func NewPrivateKeyMock() *PrivateKeyMock {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)

	return &PrivateKeyMock{
		privateKey: prvKey,
	}
}

// ToByteArray returns the byte array representation of the key
func (p *PrivateKeyMock) ToByteArray() ([]byte, error) {
	return p.privateKey.X.Bytes(), nil
}

// GeneratePublic builds a public key for the current private key
func (p *PrivateKeyMock) GeneratePublic() crypto.PublicKey {
	return nil
}

// Suite returns the suite used by this key
func (p *PrivateKeyMock) Suite() crypto.Suite {
	return nil
}

// Scalar returns the Scalar corresponding to this Private Key
func (p *PrivateKeyMock) Scalar() crypto.Scalar {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *PrivateKeyMock) IsInterfaceNil() bool {
	return p == nil
}
