package mock

import (
	"bytes"

	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// PrivateKeyStub provides stubs for a PrivateKey implementation
type PrivateKeyStub struct {
	ToByteArrayStub    func() ([]byte, error)
	GeneratePublicStub func() crypto.PublicKey
	ScalarStub         func() crypto.Scalar
	SuiteStub          func() crypto.Suite
}

// PublicKeyStub provides stubs for a PublicKey implementation
type PublicKeyStub struct {
	ToByteArrayStub func() ([]byte, error)
	PointStub       func() crypto.Point
	SuiteStub       func() crypto.Suite
}

// ToByteArray returns the byte array representation of the private key
func (privKey *PrivateKeyStub) ToByteArray() ([]byte, error) {
	if privKey.ToByteArrayStub != nil {
		return privKey.ToByteArrayStub()
	}
	return bytes.Repeat([]byte("a"), 32), nil
}

// GeneratePublic builds a public key for the current private key
func (privKey *PrivateKeyStub) GeneratePublic() crypto.PublicKey {
	if privKey.GeneratePublicStub != nil {
		return privKey.GeneratePublicStub()
	}
	return &PublicKeyStub{}
}

// Suite returns the Suite (curve data) used for this private key
func (privKey *PrivateKeyStub) Suite() crypto.Suite {
	if privKey.SuiteStub != nil {
		return privKey.SuiteStub()
	}
	return nil
}

// Scalar returns the Scalar corresponding to this Private Key
func (privKey *PrivateKeyStub) Scalar() crypto.Scalar {
	if privKey.ScalarStub != nil {
		return privKey.ScalarStub()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (privKey *PrivateKeyStub) IsInterfaceNil() bool {
	return privKey == nil
}

// ToByteArray returns the byte array representation of the public key
func (pubKey *PublicKeyStub) ToByteArray() ([]byte, error) {
	if pubKey.ToByteArrayStub != nil {
		return pubKey.ToByteArrayStub()
	}
	return []byte("public key"), nil
}

// Suite returns the Suite (curve data) used for this private key
func (pubKey *PublicKeyStub) Suite() crypto.Suite {
	if pubKey.SuiteStub != nil {
		return pubKey.SuiteStub()
	}
	return nil
}

// Point returns the Point corresponding to this Public Key
func (pubKey *PublicKeyStub) Point() crypto.Point {
	if pubKey.PointStub != nil {
		return pubKey.PointStub()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pubKey *PublicKeyStub) IsInterfaceNil() bool {
	return pubKey == nil
}
