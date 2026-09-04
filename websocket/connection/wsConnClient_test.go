package connection

import (
	"errors"
	"net"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func filterAddress(originalURL string) string {
	if strings.Contains(originalURL, "://") {
		originalURL = strings.Split(originalURL, "://")[1]
	}

	return originalURL
}

func createConnectionURLForTestServer(server *httptest.Server) string {
	u := url.URL{
		Scheme: "ws",
		Host:   filterAddress(server.URL),
		Path:   "/echo",
	}

	return u.String()
}

func TestWsConnClient_OpenCloseConnectionShouldWork(t *testing.T) {
	t.Parallel()

	testServer := testscommon.NewHttpTestEchoHandler()
	defer testServer.Close()

	conClient := NewWSConnClient(time.Second)
	connectionURL := createConnectionURLForTestServer(testServer)
	err := conClient.OpenConnection(connectionURL)
	require.Nil(t, err)

	err = conClient.Close()
	require.Nil(t, err)
}

func TestWsConnClient_WriteAndReadMessageShouldWork(t *testing.T) {
	t.Parallel()

	testServer := testscommon.NewHttpTestEchoHandler()
	defer testServer.Close()

	conClient := NewWSConnClient(time.Second)
	connectionURL := createConnectionURLForTestServer(testServer)
	_ = conClient.OpenConnection(connectionURL)
	defer func() {
		_ = conClient.Close()
	}()

	message := "TEST"
	err := conClient.WriteMessage(websocket.TextMessage, []byte(message))
	require.Nil(t, err)

	messageType, receivedMessage, err := conClient.ReadMessage()
	require.Nil(t, err)
	assert.Equal(t, websocket.TextMessage, messageType)
	assert.Equal(t, "ECHO: "+message, string(receivedMessage))
}

func TestWsConnClient_WorkingWithANonOpenedConnectionShouldNotPanic(t *testing.T) {
	t.Parallel()

	conClient := NewWSConnClient(time.Second)
	assert.NotPanics(t, func() {
		err := conClient.Close()
		assert.Equal(t, data.ErrConnectionNotOpen, err)
	})
	assert.NotPanics(t, func() {
		err := conClient.WriteMessage(websocket.TextMessage, []byte("TEST"))
		assert.Equal(t, data.ErrConnectionNotOpen, err)
	})
	assert.NotPanics(t, func() {
		messageType, message, err := conClient.ReadMessage()
		assert.Equal(t, data.ErrConnectionNotOpen, err)
		assert.Equal(t, 0, messageType)
		assert.Nil(t, message)
	})
}

func TestWsConnClient_WorkingWithAClosedConnectionShouldNotPanic(t *testing.T) {
	t.Parallel()

	testServer := testscommon.NewHttpTestEchoHandler()
	defer testServer.Close()

	conClient := NewWSConnClient(time.Second)
	connectionURL := createConnectionURLForTestServer(testServer)
	_ = conClient.OpenConnection(connectionURL)
	_ = conClient.Close()

	assert.NotPanics(t, func() {
		err := conClient.Close()
		assert.Equal(t, data.ErrConnectionNotOpen, err)
	})
	assert.NotPanics(t, func() {
		err := conClient.WriteMessage(websocket.TextMessage, []byte("TEST"))
		assert.Equal(t, data.ErrConnectionNotOpen, err)
	})
	assert.NotPanics(t, func() {
		messageType, message, err := conClient.ReadMessage()
		assert.Equal(t, data.ErrConnectionNotOpen, err)
		assert.Equal(t, 0, messageType)
		assert.Nil(t, message)
	})
}

func TestWsConnClient_ReOpenConnectionAfterCloseShouldWork(t *testing.T) {
	t.Parallel()

	testServer := testscommon.NewHttpTestEchoHandler()
	defer testServer.Close()

	conClient := NewWSConnClient(time.Second)
	connectionURL := createConnectionURLForTestServer(testServer)
	err := conClient.OpenConnection(connectionURL)
	require.Nil(t, err)
	err = conClient.Close()
	require.Nil(t, err)

	err = conClient.OpenConnection(connectionURL)
	require.Nil(t, err)

	message := "TEST"
	err = conClient.WriteMessage(websocket.TextMessage, []byte(message))
	require.Nil(t, err)

	messageType, receivedMessage, err := conClient.ReadMessage()
	require.Nil(t, err)
	assert.Equal(t, websocket.TextMessage, messageType)
	assert.Equal(t, "ECHO: "+message, string(receivedMessage))

	err = conClient.Close()
	require.Nil(t, err)
}

func TestWsConnClient_ReOpenAlreadyOpenedConnectionShouldError(t *testing.T) {
	t.Parallel()

	testServer := testscommon.NewHttpTestEchoHandler()
	defer testServer.Close()

	conClient := NewWSConnClient(time.Second)
	connectionURL := createConnectionURLForTestServer(testServer)
	err := conClient.OpenConnection(connectionURL)
	require.Nil(t, err)

	err = conClient.OpenConnection(connectionURL)
	assert.Equal(t, data.ErrConnectionAlreadyOpen, err)

	_ = conClient.Close()
}

func TestWsConnClient_IsOpen(t *testing.T) {
	testServer := testscommon.NewHttpTestEchoHandler()
	defer testServer.Close()

	conClient := NewWSConnClient(time.Second)
	connectionURL := createConnectionURLForTestServer(testServer)
	err := conClient.OpenConnection(connectionURL)
	require.Nil(t, err)

	open := conClient.IsOpen()
	require.True(t, open)

	_ = conClient.Close()
}

type closeErrConn struct {
	net.Conn
	err error
}

func (c *closeErrConn) Close() error {
	_ = c.Conn.Close()
	return c.err
}

func TestWsConnClient_CloseWithErrorShouldSetConToNil(t *testing.T) {
	t.Parallel()

	testServer := testscommon.NewHttpTestEchoHandler()
	defer testServer.Close()

	connectionURL := createConnectionURLForTestServer(testServer)
	u, err := url.Parse(connectionURL)
	require.NoError(t, err)

	rawConn, err := net.Dial("tcp", u.Host)
	require.NoError(t, err)

	closeErr := errors.New(data.ClosedConnectionMessage)
	wrappedConn := &closeErrConn{
		Conn: rawConn,
		err:  closeErr,
	}

	d := websocket.Dialer{
		ReadBufferSize:  0,
		WriteBufferSize: 0,
		NetDial: func(net, addr string) (net.Conn, error) {
			return wrappedConn, nil
		},
	}
	wsConn, _, err := d.Dial(u.String(), nil)
	if err != nil {
		_ = rawConn.Close()
	}
	require.Nil(t, err)

	conClient := NewWSConnClientWithConn(wsConn, time.Second)
	err = conClient.Close()
	require.Nil(t, err)
	require.Nil(t, conClient.conn)
	require.False(t, conClient.IsOpen())

	err = conClient.Close()
	require.Equal(t, data.ErrConnectionNotOpen, err)
}
