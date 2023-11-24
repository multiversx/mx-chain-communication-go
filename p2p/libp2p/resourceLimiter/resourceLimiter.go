package resourceLimiter

import (
	"fmt"

	"github.com/libp2p/go-libp2p"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
)

const oneMegabyteInBytes = 1024 * 1024

// CreateResourceLimiterOption will create a new resource limiter option
func CreateResourceLimiterOption(cfg config.ResourceLimiterConfig) (libp2p.Option, error) {
	switch cfg.Type {
	case p2p.DefaultAutoscaleResourceLimiter:
		return libp2p.DefaultResourceManager, nil
	case p2p.InfiniteResourceLimiter:
		resourceManager, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
		if err != nil {
			return nil, err
		}

		return libp2p.ResourceManager(resourceManager), nil
	case p2p.DefaultWithScaleResourceLimiter:
		limits := rcmgr.DefaultLimits
		memoryInBytes := oneMegabyteInBytes * cfg.ManualSystemMemoryInMB
		resourceManager, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(limits.Scale(memoryInBytes, cfg.ManualMaximumFD)))
		if err != nil {
			return nil, err
		}

		return libp2p.ResourceManager(resourceManager), nil
	default:
		return nil, fmt.Errorf("%w for resourceLimiterType %s", p2p.ErrUnknownResourceLimiterType, cfg.Type)
	}
}
