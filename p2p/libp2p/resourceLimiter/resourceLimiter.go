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
	ipv4ConnLimit := parseConnLimit(cfg.Ipv4ConnLimit)
	ipv6ConnLimit := parseConnLimit(cfg.Ipv6ConnLimit)

	switch cfg.Type {
	case p2p.DefaultAutoscaleResourceLimiter:
		limits := rcmgr.DefaultLimits
		libp2p.SetDefaultServiceLimits(&limits)
		resourceManager, err := rcmgr.NewResourceManager(
			rcmgr.NewFixedLimiter(limits.AutoScale()),
			rcmgr.WithLimitPeersPerCIDR(ipv4ConnLimit, ipv6ConnLimit),
		)
		if err != nil {
			return nil, err
		}

		return libp2p.ResourceManager(resourceManager), nil
	case p2p.InfiniteResourceLimiter:
		resourceManager, err := rcmgr.NewResourceManager(
			rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits),
			rcmgr.WithLimitPeersPerCIDR(ipv4ConnLimit, ipv6ConnLimit),
		)
		if err != nil {
			return nil, err
		}

		return libp2p.ResourceManager(resourceManager), nil
	case p2p.DefaultWithScaleResourceLimiter:
		limits := rcmgr.DefaultLimits
		memoryInBytes := oneMegabyteInBytes * cfg.ManualSystemMemoryInMB
		resourceManager, err := rcmgr.NewResourceManager(
			rcmgr.NewFixedLimiter(limits.Scale(memoryInBytes, cfg.ManualMaximumFD)),
			rcmgr.WithLimitPeersPerCIDR(ipv4ConnLimit, ipv6ConnLimit),
		)
		if err != nil {
			return nil, err
		}

		return libp2p.ResourceManager(resourceManager), nil
	default:
		return nil, fmt.Errorf("%w for resourceLimiterType %s", p2p.ErrUnknownResourceLimiterType, cfg.Type)
	}
}

func parseConnLimit(cfgConnLimit []config.ConnLimitConfig) []rcmgr.ConnLimitPerCIDR {
	// if config not provided, return nil so default values will be used
	if len(cfgConnLimit) == 0 {
		return nil
	}

	libp2pConnLimit := make([]rcmgr.ConnLimitPerCIDR, 0, len(cfgConnLimit))
	for _, connLimit := range cfgConnLimit {
		libp2pConnLimit = append(libp2pConnLimit, rcmgr.ConnLimitPerCIDR{
			BitMask:   connLimit.BitMask,
			ConnCount: connLimit.ConnCount,
		})
	}

	return libp2pConnLimit
}
