package p2p

import "time"

const (
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

// NetworkMessengerType defines the type of the network messenger
type NetworkMessengerType string

// RegularMessenger is the regular p2p messenger
const RegularMessenger NetworkMessengerType = "Regular"

// FullArchiveMessenger is the p2p messenger that serves a full archive node
const FullArchiveMessenger NetworkMessengerType = "FullArchive"
