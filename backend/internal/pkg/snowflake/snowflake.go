package snowflake

import (
	"errors"
	"sync"
	"time"
)

const (
	epochMillis  int64 = 1704067200000
	nodeBits           = 10
	sequenceBits       = 12
	maxNodeID    int64 = -1 ^ (-1 << nodeBits)
	maxSequence  int64 = -1 ^ (-1 << sequenceBits)
	nodeShift          = sequenceBits
	timeShift          = sequenceBits + nodeBits
)

type Node struct {
	mu            sync.Mutex
	nodeID        int64
	lastTimestamp int64
	sequence      int64
}

func NewNode(nodeID int64) (*Node, error) {
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, errors.New("snowflake node id out of range")
	}

	return &Node{
		nodeID:        nodeID,
		lastTimestamp: -1,
	}, nil
}

func (n *Node) NextID() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := currentMillis()
	if now < n.lastTimestamp {
		now = n.lastTimestamp
	}

	if now == n.lastTimestamp {
		n.sequence = (n.sequence + 1) & maxSequence
		if n.sequence == 0 {
			for now <= n.lastTimestamp {
				now = currentMillis()
			}
		}
	} else {
		n.sequence = 0
	}

	n.lastTimestamp = now
	return ((now - epochMillis) << timeShift) | (n.nodeID << nodeShift) | n.sequence
}

func currentMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
