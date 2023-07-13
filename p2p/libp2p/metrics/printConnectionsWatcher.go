package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-storage-go/timecache"
	"github.com/multiversx/mx-chain-storage-go/types"
)

const minTimeToLive = time.Second

type printConnectionsWatcher struct {
	timeCacher      types.TimeCacher
	goRoutineClosed atomic.Flag
	timeToLive      time.Duration
	printHandler    func(pid core.PeerID, connection string, log p2p.Logger)
	cancel          func()
	log             p2p.Logger
}

// NewPrintConnectionsWatcher creates a new
func NewPrintConnectionsWatcher(timeToLive time.Duration, logger p2p.Logger) (*printConnectionsWatcher, error) {
	if timeToLive < minTimeToLive {
		return nil, fmt.Errorf("%w in NewPrintConnectionsWatcher, got: %d, minimum: %d", ErrInvalidValueForTimeToLiveParam, timeToLive, minTimeToLive)
	}
	if check.IfNil(logger) {
		return nil, p2p.ErrNilLogger
	}

	pcw := &printConnectionsWatcher{
		timeToLive:   timeToLive,
		timeCacher:   timecache.NewTimeCache(timeToLive),
		printHandler: logPrintHandler,
		log:          logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	pcw.cancel = cancel
	go pcw.doSweep(ctx)

	return pcw, nil
}

func (pcw *printConnectionsWatcher) doSweep(ctx context.Context) {
	timer := time.NewTimer(pcw.timeToLive)
	defer func() {
		timer.Stop()
		pcw.goRoutineClosed.SetValue(true)
	}()

	for {
		timer.Reset(pcw.timeToLive)

		select {
		case <-ctx.Done():
			pcw.log.Debug("printConnectionsWatcher's processing loop is closing...")
			return
		case <-timer.C:
		}

		pcw.timeCacher.Sweep()
	}
}

// NewKnownConnection will add the known connection to the cache, printing it as necessary
func (pcw *printConnectionsWatcher) NewKnownConnection(pid core.PeerID, connection string) {
	conn := strings.Trim(connection, " ")
	if len(conn) == 0 {
		return
	}

	has := pcw.timeCacher.Has(pid.Pretty())
	err := pcw.timeCacher.Upsert(pid.Pretty(), pcw.timeToLive)
	if err != nil {
		pcw.log.Warn("programming error in printConnectionsWatcher.NewKnownConnection", "error", err)
		return
	}
	if has {
		return
	}

	pcw.printHandler(pid, conn, pcw.log)
}

// Close will close any go routines opened by this instance
func (pcw *printConnectionsWatcher) Close() error {
	pcw.cancel()

	return nil
}

func logPrintHandler(pid core.PeerID, connection string, log p2p.Logger) {
	log.Debug("new known peer", "pid", pid.Pretty(), "connection", connection)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pcw *printConnectionsWatcher) IsInterfaceNil() bool {
	return pcw == nil
}
