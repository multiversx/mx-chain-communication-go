package facade

import (
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/mock"
	"github.com/stretchr/testify/require"
)

func TestNewNetworkMessengersFacade(t *testing.T) {
	t.Parallel()

	t.Run("no messenger should error", func(t *testing.T) {
		t.Parallel()

		facade, err := NewNetworkMessengersFacade()
		require.Equal(t, p2p.ErrEmptyMessengersList, err)
		require.Nil(t, facade)
	})
	t.Run("nil messenger should error", func(t *testing.T) {
		t.Parallel()

		facade, err := NewNetworkMessengersFacade(&mock.MessengerStub{}, nil)
		require.Equal(t, p2p.ErrNilMessenger, err)
		require.Nil(t, facade)
	})
	t.Run("should work with one messenger", func(t *testing.T) {
		t.Parallel()

		facade, err := NewNetworkMessengersFacade(&mock.MessengerStub{})
		require.NoError(t, err)
		require.NotNil(t, facade)
	})
	t.Run("should work multiple messengers", func(t *testing.T) {
		t.Parallel()

		facade, err := NewNetworkMessengersFacade(&mock.MessengerStub{}, &mock.MessengerStub{}, &mock.MessengerStub{})
		require.NoError(t, err)
		require.NotNil(t, facade)
	})
}
