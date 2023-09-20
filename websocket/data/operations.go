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
	URL                        string // The WebSocket URL to connect to.
	Mode                       string // The host operation mode: 'client' or 'server'.
	RetryDurationInSec         int    // The duration in seconds to wait before retrying the connection in case of failure.
	WithAcknowledge            bool   // Set to `true` to enable message acknowledgment mechanism.
	AcknowledgeTimeoutInSec    int    // The duration in seconds to wait for an acknowledgement message
	BlockingAckOnError         bool   // Set to `true` to send the acknowledgment message only if the processing part of a message succeeds. If an error occurs during processing, the acknowledgment will not be sent.
	DropMessagesIfNoConnection bool   // Set to `true` to drop messages if there is no active WebSocket connection to send to.
	Version                    uint32 // Defines the payload version.
}
