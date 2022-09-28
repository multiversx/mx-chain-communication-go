package crypto_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-p2p/libp2p/crypto"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIdentityGenerator(t *testing.T) {
	t.Parallel()

	generator := crypto.NewIdentityGenerator()
	assert.False(t, check.IfNil(generator))
}

func TestIdentityGenerator_CreateP2PPrivateKey(t *testing.T) {
	t.Parallel()

	generator := crypto.NewIdentityGenerator()

	skKey1, errGenerate := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	require.Nil(t, errGenerate)
	skBuff1, errMarshal := (*libp2pCrypto.Secp256k1PrivateKey)(skKey1).Raw()
	require.Nil(t, errMarshal)

	skKey2, errGenerate := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	require.Nil(t, errGenerate)
	skBuff2, errMarshal := (*libp2pCrypto.Secp256k1PrivateKey)(skKey2).Raw()
	require.Nil(t, errMarshal)

	t.Run("same private key bytes should produce the same private key", func(t *testing.T) {

		sk1, err := generator.CreateP2PPrivateKey(skBuff1)
		assert.Nil(t, err)

		sk2, err := generator.CreateP2PPrivateKey(skBuff1)
		assert.Nil(t, err)

		assert.Equal(t, sk1, sk2)
	})
	t.Run("different private key bytes should produce different private key", func(t *testing.T) {
		sk1, err := generator.CreateP2PPrivateKey(skBuff1)
		assert.Nil(t, err)

		sk2, err := generator.CreateP2PPrivateKey(skBuff2)
		assert.Nil(t, err)

		assert.NotEqual(t, sk1, sk2)
	})
	t.Run("empty private key bytes should produce different private key", func(t *testing.T) {
		sk1, err := generator.CreateP2PPrivateKey(make([]byte, 0))
		assert.Nil(t, err)

		sk2, err := generator.CreateP2PPrivateKey(make([]byte, 0))
		assert.Nil(t, err)

		assert.NotEqual(t, sk1, sk2)
	})
	t.Run("nil private key bytes should produce different private key", func(t *testing.T) {
		sk1, err := generator.CreateP2PPrivateKey(nil)
		assert.Nil(t, err)

		sk2, err := generator.CreateP2PPrivateKey(nil)
		assert.Nil(t, err)

		assert.NotEqual(t, sk1, sk2)
	})
}

func TestIdentityGenerator_CreateRandomP2PIdentity(t *testing.T) {
	t.Parallel()

	generator := crypto.NewIdentityGenerator()
	sk1, pid1, err := generator.CreateRandomP2PIdentity()
	assert.Nil(t, err)

	sk2, pid2, err := generator.CreateRandomP2PIdentity()
	assert.Nil(t, err)

	assert.NotEqual(t, sk1, sk2)
	assert.NotEqual(t, pid1, pid2)
	assert.Equal(t, 32, len(sk1))
	assert.Equal(t, 39, len(pid1))
	assert.Equal(t, 32, len(sk2))
	assert.Equal(t, 39, len(pid2))
}
