package idgenerator

import (
	"sync"
	"time"
)

var _ IDGeneratorInterface = (*SnowflakeGenerator)(nil)

type SnowflakeGenerator struct {
	mu            sync.Mutex
	lastTimestamp int64
	sequenceNo    int64
	machineID     int64
}

func NewSnowflakeGenerator() *SnowflakeGenerator {
	return &SnowflakeGenerator{
		lastTimestamp: time.Now().UnixNano(),
		sequenceNo:    0,
		machineID:     0,
	}
}

func (g *SnowflakeGenerator) GenerateShortCode() (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()

	if now == g.lastTimestamp {
		// same millisecond, increment sequenceID
		g.sequenceNo++
		if g.sequenceNo > 4095 { // 12 bits
			// wait for next millisecond
			for now <= g.lastTimestamp {
				now = time.Now().UnixMilli()
			}
			g.sequenceNo = 0
		}
	} else {
		g.sequenceNo = 0
	}

	g.lastTimestamp = now

	// Combine timestamp, machine ID, and sequence
	id := (now << 22) | (g.machineID << 12) | g.sequenceNo

	return base62Encode(id), nil
}

// base62Encode converts a number to base62
func base62Encode(n int64) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	if n == 0 {
		return "0"
	}

	var result []byte
	base := int64(len(charset))

	for n > 0 {
		result = append([]byte{charset[n%base]}, result...)
		n /= base
	}

	return string(result)
}
