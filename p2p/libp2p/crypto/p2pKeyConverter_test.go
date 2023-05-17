package crypto_test

import (
	"errors"
	"testing"

	p2pCrypto "github.com/multiversx/mx-chain-communication-go/p2p/libp2p/crypto"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/secp256k1"
	"github.com/stretchr/testify/assert"
)

func TestConvertPublicKeyToPeerID(t *testing.T) {
	t.Parallel()

	t.Run("from a nil public key should error", func(t *testing.T) {
		t.Parallel()

		conv := p2pCrypto.NewP2PKeyConverter()
		pid, err := conv.ConvertPublicKeyToPeerID(nil)
		assert.Empty(t, pid)
		assert.Equal(t, p2pCrypto.ErrNilPublicKey, err)
	})
	t.Run("ToByteArray errors, should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		mockPk := &mock.PublicKeyStub{
			ToByteArrayStub: func() ([]byte, error) {
				return nil, expectedErr
			},
		}

		conv := p2pCrypto.NewP2PKeyConverter()
		pid, err := conv.ConvertPublicKeyToPeerID(mockPk)
		assert.Empty(t, pid)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("from a key that is not compatible with libp2p, should error", func(t *testing.T) {
		t.Parallel()

		mockPk := &mock.PublicKeyStub{
			ToByteArrayStub: func() ([]byte, error) {
				return []byte("too short byte slice"), nil
			},
		}

		conv := p2pCrypto.NewP2PKeyConverter()
		pid, err := conv.ConvertPublicKeyToPeerID(mockPk)
		assert.Empty(t, pid)
		assert.NotNil(t, err)
		assert.Equal(t, "malformed public key: invalid length: 20", err.Error())
	})
	t.Run("should work using a generated key with the KeyGenerator", func(t *testing.T) {
		t.Parallel()

		keyGen := signing.NewKeyGenerator(secp256k1.NewSecp256k1())
		_, pk := keyGen.GeneratePair()

		conv := p2pCrypto.NewP2PKeyConverter()
		pid, err := conv.ConvertPublicKeyToPeerID(pk)
		assert.NotEmpty(t, pid)
		assert.Nil(t, err)
	})
	t.Run("should work using a generated identity", func(t *testing.T) {
		t.Parallel()

		generator := p2pCrypto.NewIdentityGenerator()
		skBytes, pid, err := generator.CreateRandomP2PIdentity()
		assert.Nil(t, err)

		keyGen := signing.NewKeyGenerator(secp256k1.NewSecp256k1())
		sk, err := keyGen.PrivateKeyFromByteArray(skBytes)
		assert.Nil(t, err)

		pk := sk.GeneratePublic()
		conv := p2pCrypto.NewP2PKeyConverter()
		recoveredPid, err := conv.ConvertPublicKeyToPeerID(pk)
		assert.Nil(t, err)

		assert.Equal(t, pid, recoveredPid)
	})
}
