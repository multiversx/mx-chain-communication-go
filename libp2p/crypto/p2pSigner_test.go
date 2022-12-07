package crypto_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/secp256k1"
	p2pCrypto "github.com/ElrondNetwork/elrond-go-p2p/libp2p/crypto"
	"github.com/ElrondNetwork/elrond-go-p2p/mock"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generatePrivateKey() (crypto.PrivateKey, crypto.PublicKey) {
	keyGen := signing.NewKeyGenerator(secp256k1.NewSecp256k1())
	prvKey, pubKey := keyGen.GeneratePair()

	return prvKey, pubKey
}

func createDefaultP2PSignerArgs() p2pCrypto.ArgsP2pSignerWrapper {
	return p2pCrypto.ArgsP2pSignerWrapper{
		PrivateKey: &mock.PrivateKeyStub{},
		Signer:     &mock.SingleSignerStub{},
		KeyGen:     &mock.KeyGenStub{},
	}
}

func TestP2pSigner_NewP2PSigner(t *testing.T) {
	t.Parallel()

	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createDefaultP2PSignerArgs()
		args.PrivateKey = nil

		sig, err := p2pCrypto.NewP2PSignerWrapper(args)
		assert.Equal(t, p2pCrypto.ErrNilPrivateKey, err)
		assert.Nil(t, sig)
	})

	t.Run("nil single signer should error", func(t *testing.T) {
		t.Parallel()

		args := createDefaultP2PSignerArgs()
		args.Signer = nil

		sig, err := p2pCrypto.NewP2PSignerWrapper(args)
		assert.Equal(t, p2pCrypto.ErrNilSingleSigner, err)
		assert.Nil(t, sig)
	})

	t.Run("nil key generator should error", func(t *testing.T) {
		t.Parallel()

		args := createDefaultP2PSignerArgs()
		args.KeyGen = nil

		sig, err := p2pCrypto.NewP2PSignerWrapper(args)
		assert.Equal(t, p2pCrypto.ErrNilKeyGenerator, err)
		assert.Nil(t, sig)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		sig, err := p2pCrypto.NewP2PSignerWrapper(createDefaultP2PSignerArgs())
		assert.Nil(t, err)
		assert.NotNil(t, sig)
	})
}

func TestP2pSigner_Sign(t *testing.T) {
	t.Parallel()

	args := createDefaultP2PSignerArgs()

	wasCalled := false
	args.Signer = &mock.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			wasCalled = true
			return []byte("sig"), nil
		},
	}
	signer, _ := p2pCrypto.NewP2PSignerWrapper(args)

	sig, err := signer.Sign([]byte("payload"))
	assert.Nil(t, err)
	assert.NotNil(t, sig)
	assert.True(t, wasCalled)
}

func convertPublicKeyToP2PPublicKey(pk crypto.PublicKey) libp2pCrypto.PubKey {
	pkBytes, _ := pk.ToByteArray()
	pubKey, _ := libp2pCrypto.UnmarshalSecp256k1PublicKey(pkBytes)
	return pubKey
}

func TestP2pSigner_Verify(t *testing.T) {
	t.Parallel()

	payload := []byte("payload")

	t.Run("fail to verify, should return err", func(t *testing.T) {
		t.Parallel()

		args := createDefaultP2PSignerArgs()

		expectedErr := errors.New("expected error")
		args.Signer = &mock.SingleSignerStub{
			VerifyCalled: func(public crypto.PublicKey, msg, sig []byte) error {
				return expectedErr
			},
		}

		signer, _ := p2pCrypto.NewP2PSignerWrapper(args)

		_, pk := generatePrivateKey()
		peerID, err := peer.IDFromPublicKey(convertPublicKeyToP2PPublicKey(pk))
		require.Nil(t, err)

		err = signer.Verify(payload, core.PeerID(peerID), []byte("sig"))
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createDefaultP2PSignerArgs()

		verifyWasCalled := false
		args.Signer = &mock.SingleSignerStub{
			VerifyCalled: func(public crypto.PublicKey, msg, sig []byte) error {
				verifyWasCalled = true
				return nil
			},
		}

		signer, _ := p2pCrypto.NewP2PSignerWrapper(args)

		_, pk := generatePrivateKey()
		peerID, err := peer.IDFromPublicKey(convertPublicKeyToP2PPublicKey(pk))
		require.Nil(t, err)

		err = signer.Verify(payload, core.PeerID(peerID), []byte("sig"))
		require.Nil(t, err)

		assert.True(t, verifyWasCalled)
	})
}

func TestP2PSigner_SignUsingPrivateKey(t *testing.T) {
	t.Parallel()

	payload := []byte("payload")
	pkBytes := []byte("private key bytes")

	t.Run("fail to convert private key", func(t *testing.T) {
		t.Parallel()

		args := createDefaultP2PSignerArgs()

		expectedErr := errors.New("expected error")
		args.KeyGen = &mock.KeyGenStub{
			PrivateKeyFromByteArrayStub: func(b []byte) (crypto.PrivateKey, error) {
				return nil, expectedErr
			},
		}

		signer, _ := p2pCrypto.NewP2PSignerWrapper(args)

		sig, err := signer.SignUsingPrivateKey(pkBytes, payload)
		assert.Nil(t, sig)
		assert.Equal(t, expectedErr, err)

	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createDefaultP2PSignerArgs()

		keyGenWasCalled := false
		args.KeyGen = &mock.KeyGenStub{
			PrivateKeyFromByteArrayStub: func(b []byte) (crypto.PrivateKey, error) {
				require.Equal(t, pkBytes, b)
				keyGenWasCalled = true
				return &mock.PrivateKeyStub{}, nil
			},
		}
		signerWasCalled := false
		args.Signer = &mock.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				signerWasCalled = true
				return []byte{}, nil
			},
		}

		signer, _ := p2pCrypto.NewP2PSignerWrapper(args)

		sig, err := signer.SignUsingPrivateKey(pkBytes, payload)
		assert.Nil(t, err)
		assert.NotNil(t, sig)

		assert.True(t, keyGenWasCalled)
		assert.True(t, signerWasCalled)
	})
}

func TestP2pSigner_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	numOps := 1000
	wg := sync.WaitGroup{}
	wg.Add(numOps)

	payload1 := []byte("payload1")
	payload2 := []byte("payload2")

	sk, pk := generatePrivateKey()
	args := createDefaultP2PSignerArgs()
	args.PrivateKey = sk

	signer, _ := p2pCrypto.NewP2PSignerWrapper(args)
	libp2pPid, _ := peer.IDFromPublicKey(convertPublicKeyToP2PPublicKey(pk))
	pid := core.PeerID(libp2pPid)

	sig1, _ := signer.Sign(payload1)

	for i := 0; i < numOps; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)

			switch idx {
			case 0:
				_, errSign := signer.Sign(payload2)
				assert.Nil(t, errSign)
			case 1:
				errVerify := signer.Verify(payload1, pid, sig1)
				assert.Nil(t, errVerify)
			case 2:
				errVerify := signer.Verify(payload1, pid, sig1)
				assert.Nil(t, errVerify)
			}

			wg.Done()
		}(i % 3)
	}

	wg.Wait()
}
