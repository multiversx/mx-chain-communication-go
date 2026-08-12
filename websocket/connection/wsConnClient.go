package connection

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("connection")

type wsConnClient struct {
	mut          sync.RWMutex
	conn         *websocket.Conn
	clientID     string
	writeTimeout time.Duration
}

// NewWSConnClient creates a new wrapper over a websocket connection
func NewWSConnClient(writeTimeout time.Duration) *wsConnClient {
	return &wsConnClient{
		writeTimeout: writeTimeout,
	}
}

// NewWSConnClientWithConn creates a new wrapper over a provided websocket connection
func NewWSConnClientWithConn(conn *websocket.Conn, writeTimeout time.Duration) *wsConnClient {
	wsc := &wsConnClient{
		conn:         conn,
		writeTimeout: writeTimeout,
	}
	wsc.clientID = fmt.Sprintf("%p", wsc)

	return wsc
}

// OpenConnection will open a new client with a background context
func (wsc *wsConnClient) OpenConnection(url string) error {
	wsc.mut.Lock()
	defer wsc.mut.Unlock()

	if wsc.conn != nil {
		return data.ErrConnectionAlreadyOpen
	}

	var err error
	wsc.conn, _, err = websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	return nil
}

// ReadMessage calls the underlying reading message ws connection func
func (wsc *wsConnClient) ReadMessage() (messageType int, p []byte, err error) {
	conn, err := wsc.getConn()
	if err != nil {
		return 0, nil, err
	}

	return conn.ReadMessage()
}

func (wsc *wsConnClient) setWriteDeadline() {
	if wsc.writeTimeout == 0 {
		return
	}

	err := wsc.conn.SetWriteDeadline(time.Now().Add(wsc.writeTimeout))
	if err != nil {
		log.Trace("cannot set write deadline", "error", err)
	}
}

// WriteMessage calls the underlying write message ws connection func
func (wsc *wsConnClient) WriteMessage(messageType int, payload []byte) error {
	wsc.mut.Lock()
	defer wsc.mut.Unlock()

	if wsc.conn == nil {
		return data.ErrConnectionNotOpen
	}

	wsc.setWriteDeadline()

	return wsc.conn.WriteMessage(messageType, payload)
}

// IsOpen will return true if the connection is open, false otherwise
func (wsc *wsConnClient) IsOpen() bool {
	wsc.mut.RLock()
	defer wsc.mut.RUnlock()

	return wsc.conn != nil
}

func (wsc *wsConnClient) getConn() (*websocket.Conn, error) {
	wsc.mut.RLock()
	defer wsc.mut.RUnlock()

	if wsc.conn == nil {
		return nil, data.ErrConnectionNotOpen
	}

	conn := wsc.conn

	return conn, nil
}

// GetID will return the unique id of the client
func (wsc *wsConnClient) GetID() string {
	return wsc.clientID
}

// Close will try to cleanly close the connection, if possible
func (wsc *wsConnClient) Close() error {
	// critical section
	wsc.mut.Lock()
	defer wsc.mut.Unlock()

	if wsc.conn == nil {
		return data.ErrConnectionNotOpen
	}

	log.Debug("closing ws connection...")

	conn := wsc.conn
	wsc.conn = nil

	if wsc.writeTimeout > 0 {
		_ = conn.SetWriteDeadline(time.Now().Add(wsc.writeTimeout))
	}

	//Cleanly close the connection by sending a close message and then
	//waiting (with timeout) for the server to close the connection.
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Trace("cannot send close message", "error", err)
	}

	err = conn.Close()
	if err != nil && !strings.Contains(err.Error(), data.ClosedConnectionMessage) {
		return err
	}

	return nil
}

// IsInterfaceNil -
func (wsc *wsConnClient) IsInterfaceNil() bool {
	return wsc == nil
}
