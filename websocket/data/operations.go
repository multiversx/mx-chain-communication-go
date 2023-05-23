package data

const (
	// WSRoute is the route which data will be sent over websocket
	WSRoute = "/save"
	// ModeServer is a constant value that is used to indicate that the WebSocket host should start in server mode, meaning it will listen for incoming connections from clients and respond to them.
	ModeServer = "server"
	// ModeClient is a constant value that is used to indicate that the WebSocket host should start in client mode, meaning it will initiate connections to a remote server.
	ModeClient = "client"
)

// WebSocketConfig holds the configuration needed for instantiating a new web socket server
type WebSocketConfig struct {
	URL                        string
	Mode                       string
	RetryDurationInSec         int
	WithAcknowledge            bool
	BlockingAckOnError         bool
	DropMessagesIfNoConnection bool
}
