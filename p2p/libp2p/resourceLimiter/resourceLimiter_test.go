package resourceLimiter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	"github.com/stretchr/testify/assert"
)

const testManualSystemMem = 3737
const testManualFD = 3737373

func TestCreateResourceLimiterOption(t *testing.T) {
	t.Parallel()

	t.Run("unknown type should error", func(t *testing.T) {
		t.Parallel()

		unknownType := "unknown-type"
		opt, err := CreateResourceLimiterOption(config.ResourceLimiterConfig{
			Type:                   unknownType,
			ManualSystemMemoryInMB: testManualSystemMem,
			ManualMaximumFD:        testManualFD,
		})

		assert.Nil(t, opt)
		assert.ErrorIs(t, err, p2p.ErrUnknownResourceLimiterType)
		assert.Contains(t, err.Error(), unknownType)
	})
	t.Run("default autoscale should work", func(t *testing.T) {
		opt, err := CreateResourceLimiterOption(config.ResourceLimiterConfig{
			Type:                   p2p.DefaultAutoscaleResourceLimiter,
			ManualSystemMemoryInMB: testManualSystemMem,
			ManualMaximumFD:        testManualFD,
		})

		assert.Nil(t, err)

		cfg := &libp2p.Config{}
		err = opt(cfg) // apply the option
		assert.Nil(t, err)
		assert.NotNil(t, cfg.ResourceManager)
		// we can not assert the limits values as it depends on the machine this test is run on
		// we are assuming that the defaults are working
	})
	t.Run("infinite should work", func(t *testing.T) {
		opt, err := CreateResourceLimiterOption(config.ResourceLimiterConfig{
			Type:                   p2p.InfiniteResourceLimiter,
			ManualSystemMemoryInMB: testManualSystemMem,
			ManualMaximumFD:        testManualFD,
		})

		assert.Nil(t, err)

		cfg := &libp2p.Config{}
		err = opt(cfg) // apply the option
		assert.Nil(t, err)
		assert.NotNil(t, cfg.ResourceManager)

		limits, err := extractSystemLimitValuesFromResourceManager(cfg.ResourceManager)
		assert.Nil(t, err)

		expectedSystemLimits := rcmgr.BaseLimit{
			Streams:         math.MaxInt,
			StreamsInbound:  math.MaxInt,
			StreamsOutbound: math.MaxInt,
			Conns:           math.MaxInt,
			ConnsInbound:    math.MaxInt,
			ConnsOutbound:   math.MaxInt,
			FD:              math.MaxInt,
			Memory:          math.MaxInt64,
		}

		assert.Equal(t, expectedSystemLimits, limits)
	})
	t.Run("default with manual should work", func(t *testing.T) {
		opt, err := CreateResourceLimiterOption(config.ResourceLimiterConfig{
			Type:                   p2p.DefaultWithScaleResourceLimiter,
			ManualSystemMemoryInMB: testManualSystemMem,
			ManualMaximumFD:        testManualFD,
		})

		assert.Nil(t, err)

		cfg := &libp2p.Config{}
		err = opt(cfg) // apply the option
		assert.Nil(t, err)
		assert.NotNil(t, cfg.ResourceManager)

		limits, err := extractSystemLimitValuesFromResourceManager(cfg.ResourceManager)
		assert.Nil(t, err)

		expectedSystemLimits := rcmgr.BaseLimit{
			Streams:         9522,
			StreamsInbound:  4761,
			StreamsOutbound: 9522,
			Conns:           595,
			ConnsInbound:    297,
			ConnsOutbound:   595,
			FD:              testManualFD,
			Memory:          4052746240,
		}

		assert.Equal(t, expectedSystemLimits, limits)
	})
}

// we need this function in order to extract the Limits provided to the ResourceManager. As the libp2p code is
// written now, this is the only possibility to check that the limits were provided correctly.
func extractSystemLimitValuesFromResourceManager(resourceMan network.ResourceManager) (rcmgr.BaseLimit, error) {
	str := spew.Sdump(resourceMan)

	var err error

	isLimitsFound := false
	isConcreteLimitConfigFound := false
	isSystemFound := false
	baseLimit := rcmgr.BaseLimit{}

	splt := strings.Split(str, "\n")
	for _, stringLine := range splt {
		if strings.Contains(stringLine, "limits: ") {
			isLimitsFound = true
			continue
		}
		if strings.Contains(stringLine, "ConcreteLimitConfig: ") {
			isConcreteLimitConfigFound = true
			continue
		}
		if strings.Contains(stringLine, "system: ") {
			isSystemFound = true
			continue
		}
		if isLimitsFound && isConcreteLimitConfigFound && isSystemFound {
			if strings.Contains(stringLine, "},") {
				return baseLimit, nil
			}

			if strings.Contains(stringLine, "Streams") {
				baseLimit.Streams, err = extractValue(stringLine)
				if err != nil {
					return baseLimit, err
				}
			}

			if strings.Contains(stringLine, "StreamsInbound") {
				baseLimit.StreamsInbound, err = extractValue(stringLine)
				if err != nil {
					return baseLimit, err
				}
			}

			if strings.Contains(stringLine, "StreamsOutbound") {
				baseLimit.StreamsOutbound, err = extractValue(stringLine)
				if err != nil {
					return baseLimit, err
				}
			}

			if strings.Contains(stringLine, "Conns") {
				baseLimit.Conns, err = extractValue(stringLine)
				if err != nil {
					return baseLimit, err
				}
			}

			if strings.Contains(stringLine, "ConnsInbound") {
				baseLimit.ConnsInbound, err = extractValue(stringLine)
				if err != nil {
					return baseLimit, err
				}
			}

			if strings.Contains(stringLine, "ConnsOutbound") {
				baseLimit.ConnsOutbound, err = extractValue(stringLine)
				if err != nil {
					return baseLimit, err
				}
			}

			if strings.Contains(stringLine, "FD") {
				baseLimit.FD, err = extractValue(stringLine)
				if err != nil {
					return baseLimit, err
				}
			}

			if strings.Contains(stringLine, "Memory") {
				var intVal int
				intVal, err = extractValue(stringLine)
				if err != nil {
					return baseLimit, err
				}

				baseLimit.Memory = int64(intVal)
			}
		}
	}

	return baseLimit, fmt.Errorf("something is wrong while parsing this struct. "+
		"It seems the provided resource manager does not contain the limits struct %s", spew.Sdump(resourceMan))
}

func extractValue(stringLine string) (int, error) {
	filtered := strings.Trim(stringLine, " ,")
	splt := strings.Split(filtered, " ")

	val, err := strconv.Atoi(splt[2])
	if err != nil {
		return 0, fmt.Errorf("%w for line %s", err, stringLine)
	}

	return val, nil
}
