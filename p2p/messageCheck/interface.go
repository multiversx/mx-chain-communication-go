package messagecheck

import "github.com/multiversx/mx-chain-core-go/core"

type p2pSigner interface {
	Verify(payload []byte, pid core.PeerID, signature []byte) error
}
