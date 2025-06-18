package p2p

import (
	"errors"
)

// ErrNilContext signals that a nil context was provided
var ErrNilContext = errors.New("nil context")

// ErrNilMockNet signals that a nil mocknet was provided. Should occur only in testing!!!
var ErrNilMockNet = errors.New("nil mocknet provided")

// ErrNilTopic signals that a nil topic has been provided
var ErrNilTopic = errors.New("nil topic")

// ErrChannelDoesNotExist signals that a requested channel does not exist
var ErrChannelDoesNotExist = errors.New("channel does not exist")

// ErrChannelCanNotBeDeleted signals that a channel can not be deleted (might be the default channel)
var ErrChannelCanNotBeDeleted = errors.New("channel can not be deleted")

// ErrChannelCanNotBeReAdded signals that a channel can not be re added as it is the default channel
var ErrChannelCanNotBeReAdded = errors.New("channel can not be re added")

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

// ErrAlreadySeenMessage signals that the message has already been seen
var ErrAlreadySeenMessage = errors.New("already seen this message")

// ErrMessageTooNew signals that a message has a timestamp that is in the future relative to self
var ErrMessageTooNew = errors.New("message is too new")

// ErrMessageTooOld signals that a message has a timestamp that is in the past relative to self
var ErrMessageTooOld = errors.New("message is too old")

// ErrNilDirectSendMessageHandler signals that the message handler for new message has not been wired
var ErrNilDirectSendMessageHandler = errors.New("nil direct sender message handler")

// ErrPeerNotDirectlyConnected signals that the peer is not directly connected to self
var ErrPeerNotDirectlyConnected = errors.New("peer is not directly connected")

// ErrNilHost signals that a nil host has been provided
var ErrNilHost = errors.New("nil host")

// ErrNilValidator signals that a validator hasn't been set for the required topic
var ErrNilValidator = errors.New("no validator has been set for this topic")

// ErrPeerDiscoveryProcessAlreadyStarted signals that a peer discovery is already turned on
var ErrPeerDiscoveryProcessAlreadyStarted = errors.New("peer discovery is already turned on")

// ErrMessageTooLarge signals that the message provided is too large
var ErrMessageTooLarge = errors.New("buffer too large")

// ErrEmptyBufferToSend signals that an empty buffer was provided for sending to other peers
var ErrEmptyBufferToSend = errors.New("empty buffer to send")

// ErrInvalidDurationProvided signals that an invalid time.Duration has been provided
var ErrInvalidDurationProvided = errors.New("invalid time.Duration provided")

// ErrTooManyGoroutines is raised when the number of goroutines has exceeded a threshold
var ErrTooManyGoroutines = errors.New(" number of goroutines exceeded")

// ErrInvalidValue signals that an invalid value has been provided
var ErrInvalidValue = errors.New("invalid value")

// ErrInvalidPortValue signals that an invalid port value has been provided
var ErrInvalidPortValue = errors.New("invalid port value")

// ErrInvalidPortsRangeString signals that an invalid ports range string has been provided
var ErrInvalidPortsRangeString = errors.New("invalid ports range string")

// ErrInvalidStartingPortValue signals that an invalid starting port value has been provided
var ErrInvalidStartingPortValue = errors.New("invalid starting port value")

// ErrInvalidEndingPortValue signals that an invalid ending port value has been provided
var ErrInvalidEndingPortValue = errors.New("invalid ending port value")

// ErrEndPortIsSmallerThanStartPort signals that the ending port value is smaller than the starting port value
var ErrEndPortIsSmallerThanStartPort = errors.New("ending port value is smaller than the starting port value")

// ErrNoFreePortInRange signals that no free port was found from provided range
var ErrNoFreePortInRange = errors.New("no free port in range")

// ErrNilSharder signals that the provided sharder is nil
var ErrNilSharder = errors.New("nil sharder")

// ErrNilPeerShardResolver signals that the peer shard resolver provided is nil
var ErrNilPeerShardResolver = errors.New("nil PeerShardResolver")

// ErrNilMarshaller signals that an operation has been attempted to or with a nil marshaller implementation
var ErrNilMarshaller = errors.New("nil marshaller")

// ErrWrongTypeAssertion signals that a wrong type assertion occurred
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilReconnecter signals that a nil reconnecter has been provided
var ErrNilReconnecter = errors.New("nil reconnecter")

// ErrUnwantedPeer signals that the provided peer has a longer kademlia distance in respect with the already connected
// peers and a connection to this peer will result in an immediate disconnection
var ErrUnwantedPeer = errors.New("unwanted peer: will not initiate connection as it will get disconnected")

// ErrNilPeerDenialEvaluator signals that a nil peer denial evaluator was provided
var ErrNilPeerDenialEvaluator = errors.New("nil peer denial evaluator")

// ErrMessageUnmarshalError signals that an invalid message was received from a peer. There is no way to communicate
// with such a peer as it does not respect the protocol
var ErrMessageUnmarshalError = errors.New("message unmarshal error")

// ErrUnsupportedFields signals that unsupported fields are provided
var ErrUnsupportedFields = errors.New("unsupported fields")

// ErrUnsupportedMessageVersion signals that an unsupported message version was detected
var ErrUnsupportedMessageVersion = errors.New("unsupported message version")

// ErrNilSyncTimer signals that a nil sync timer was provided
var ErrNilSyncTimer = errors.New("nil sync timer")

// ErrNilPreferredPeersHolder signals that a nil preferred peers holder was provided
var ErrNilPreferredPeersHolder = errors.New("nil peers holder")

// ErrInvalidSeedersReconnectionInterval signals that an invalid seeders reconnection interval error occurred
var ErrInvalidSeedersReconnectionInterval = errors.New("invalid seeders reconnection interval")

// ErrMessageProcessorAlreadyDefined signals that a message processor was already defined on the provided topic and identifier
var ErrMessageProcessorAlreadyDefined = errors.New("message processor already defined")

// ErrMessageProcessorDoesNotExists signals that a message processor does not exist on the provided topic and identifier
var ErrMessageProcessorDoesNotExists = errors.New("message processor does not exists")

// ErrWrongTypeAssertions signals that a wrong type assertion occurred
var ErrWrongTypeAssertions = errors.New("wrong type assertion")

// ErrNilConnectionsWatcher signals that a nil connections watcher has been provided
var ErrNilConnectionsWatcher = errors.New("nil connections watcher")

// ErrNilPeersRatingHandler signals that a nil peers rating handler has been provided
var ErrNilPeersRatingHandler = errors.New("nil peers rating handler")

// ErrNilP2pPrivateKey signals that a nil p2p private key has been provided
var ErrNilP2pPrivateKey = errors.New("nil p2p private key")

// ErrNilP2pSingleSigner signals that a nil p2p single signer has been provided
var ErrNilP2pSingleSigner = errors.New("nil p2p single signer")

// ErrNilP2pKeyGenerator signals that a nil p2p key generator has been provided
var ErrNilP2pKeyGenerator = errors.New("nil p2p key generator")

// ErrNilCacher signals that a nil cacher has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilP2PSigner signals that a nil p2p signer has been provided
var ErrNilP2PSigner = errors.New("nil p2p signer")

// ErrNilPubSub signals that a nil pubSub has been provided
var ErrNilPubSub = errors.New("nil pubSub")

// ErrNoPubSub signals that no pubSub was provided
var ErrNoPubSub = errors.New("no pubSub")

// ErrNilDirectSender signals that a nil direct sender has been provided
var ErrNilDirectSender = errors.New("nil direct sender")

// ErrNilThrottler signals that a nil throttler has been provided
var ErrNilThrottler = errors.New("nil throttler")

// ErrNilChannelLoadBalancer signals that a nil channel load balancer has been provided
var ErrNilChannelLoadBalancer = errors.New("nil channel load balancer")

// ErrNilDebugger signals that a nil debugger has been provided
var ErrNilDebugger = errors.New("nil debugger")

// ErrInvalidConfig signals that an invalid config has been provided
var ErrInvalidConfig = errors.New("invalid config")

// ErrNilPeersOnChannel signals that a nil peers on channel has been provided
var ErrNilPeersOnChannel = errors.New("nil peers on channel")

// ErrNilConnectionMonitor signals that a nil connections monitor has been provided
var ErrNilConnectionMonitor = errors.New("nil connections monitor")

// ErrNilNetwork signals that a nil network has been provided
var ErrNilNetwork = errors.New("nil network")

// ErrNilPeerDiscoverer signals that a nil peer discoverer has been provided
var ErrNilPeerDiscoverer = errors.New("nil peer discoverer")

// ErrNilConnectionsMetric signals that a nil connections metric has been provided
var ErrNilConnectionsMetric = errors.New("nil connections metric")

// ErrNilConnectionsHandler signals that a nil connections handler has been provided
var ErrNilConnectionsHandler = errors.New("nil connections handler")

// ErrNilLogger signals that a nil logger has been provided
var ErrNilLogger = errors.New("nil logger")

// ErrInvalidTCPAddress signals that an invalid TCP address was used
var ErrInvalidTCPAddress = errors.New("invalid TCP address")

// ErrInvalidQUICAddress signals that an invalid QUIC address was used
var ErrInvalidQUICAddress = errors.New("invalid QUIC address")

// ErrInvalidWSAddress signals that an invalid WebSocket address was used
var ErrInvalidWSAddress = errors.New("invalid WebSocket address")

// ErrInvalidWebTransportAddress signals that an invalid WebTransport address was used
var ErrInvalidWebTransportAddress = errors.New("invalid WebTransport address")

// ErrNoTransportsDefined signals that no transports were defined
var ErrNoTransportsDefined = errors.New("no transports defined")

// ErrUnknownResourceLimiterType signals that an unknown resource limiter type was provided
var ErrUnknownResourceLimiterType = errors.New("unknown resource limiter type")

// ErrNilPubSubsHolder signals that a nil pubSubs holder has been provided
var ErrNilPubSubsHolder = errors.New("nil pubSubs holder")

// ErrNilNetworkTopicsHolder signals that a nil network topics holder has been provided
var ErrNilNetworkTopicsHolder = errors.New("nil network topics holder")
