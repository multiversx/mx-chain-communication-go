package crypto_test

import (
	"crypto/rand"
	"testing"

	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p/crypto"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIdentityGenerator(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		generator, err := crypto.NewIdentityGenerator(nil)
		assert.Equal(t, p2p.ErrNilLogger, err)
		assert.Nil(t, generator)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		generator, err := crypto.NewIdentityGenerator(&testscommon.LoggerStub{})
		assert.NoError(t, err)
		assert.NotNil(t, generator)
	})
}

func TestIdentityGenerator_CreateP2PPrivateKey(t *testing.T) {
	t.Parallel()

	generator, _ := crypto.NewIdentityGenerator(&testscommon.LoggerStub{})

	skKey1, _, errGenerate := libp2pCrypto.GenerateSecp256k1Key(rand.Reader)
	require.Nil(t, errGenerate)
	skBuff1, errMarshal := skKey1.Raw()
	require.Nil(t, errMarshal)

	skKey2, _, errGenerate := libp2pCrypto.GenerateSecp256k1Key(rand.Reader)
	require.Nil(t, errGenerate)
	skBuff2, errMarshal := skKey2.Raw()
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

	generator, _ := crypto.NewIdentityGenerator(&testscommon.LoggerStub{})
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

func TestIdentityGenerator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	generator, _ := crypto.NewIdentityGenerator(nil)
	assert.True(t, generator.IsInterfaceNil())

	generator, _ = crypto.NewIdentityGenerator(&testscommon.LoggerStub{})
	assert.False(t, generator.IsInterfaceNil())
}
