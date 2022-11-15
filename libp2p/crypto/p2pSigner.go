package crypto

import (
	"crypto/sha256"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// ArgsP2pSignerWrapper defines the arguments needed to create a p2p signer wrapper
type ArgsP2pSignerWrapper struct {
	PrivateKey crypto.PrivateKey
	Signer     crypto.SingleSigner
	KeyGen     crypto.KeyGenerator
}

type p2pSignerWrapper struct {
	privateKey crypto.PrivateKey
	signer     crypto.SingleSigner
	keyGen     crypto.KeyGenerator
}

// NewP2PSignerWrapper creates a new p2pSigner instance
func NewP2PSignerWrapper(args ArgsP2pSignerWrapper) (*p2pSignerWrapper, error) {
	if check.IfNilReflect(args.PrivateKey) {
		return nil, ErrNilPrivateKey
	}
	if check.IfNilReflect(args.Signer) {
		return nil, ErrNilSingleSigner
	}
	if check.IfNilReflect(args.KeyGen) {
		return nil, ErrNilKeyGenerator
	}

	return &p2pSignerWrapper{
		privateKey: args.PrivateKey,
		signer:     args.Signer,
		keyGen:     args.KeyGen,
	}, nil
}

// Sign will sign a payload with the internal private key
func (psw *p2pSignerWrapper) Sign(payload []byte) ([]byte, error) {
	hash := sha256.Sum256(payload)
	return psw.signer.Sign(psw.privateKey, hash[:])
}

// Verify will check that the (payload, peer ID, signature) tuple is valid or not
func (psw *p2pSignerWrapper) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	pubKey, err := ConvertPeerIDToPublicKey(psw.keyGen, pid)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(payload)
	err = psw.signer.Verify(pubKey, hash[:], signature)
	if err != nil {
		return err
	}

	return nil
}

// SignUsingPrivateKey will sign the payload with provided private key bytes
func (psw *p2pSignerWrapper) SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error) {
	sk, err := psw.keyGen.PrivateKeyFromByteArray(skBytes)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(payload)
	return psw.signer.Sign(sk, hash[:])
}
