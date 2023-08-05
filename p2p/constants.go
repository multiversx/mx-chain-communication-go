package p2p

import "time"

// NetworkType defines the type of the network a messenger is running on
type NetworkType string

const (
	// LocalHostListenAddrWithIp4AndTcp defines the local host listening ip v.4 address and TCP
	LocalHostListenAddrWithIp4AndTcp = "/ip4/127.0.0.1/tcp/%d"

	displayLastPidChars = 12

	// ListsSharder is the variant that uses lists
	ListsSharder = "ListsSharder"
	// OneListSharder is the variant that is shard agnostic and uses one list
	OneListSharder = "OneListSharder"
	// NilListSharder is the variant that will not do connection trimming
	NilListSharder = "NilListSharder"

	// ConnectionWatcherTypePrint - new connection found will be printed in the log file
	ConnectionWatcherTypePrint = "print"
	// ConnectionWatcherTypeDisabled - no connection watching should be made
	ConnectionWatcherTypeDisabled = "disabled"
	// ConnectionWatcherTypeEmpty - not set, no connection watching should be made
	ConnectionWatcherTypeEmpty = ""

	// WrongP2PMessageBlacklistDuration represents the time to keep a peer id in the blacklist if it sends a message that
	// do not follow this protocol
	WrongP2PMessageBlacklistDuration = time.Second * 7200
)

// BroadcastMethod defines the broadcast method of the message
type BroadcastMethod string

// Direct defines a direct message
const Direct BroadcastMethod = "Direct"

// Broadcast defines a broadcast message
const Broadcast BroadcastMethod = "Broadcast"
