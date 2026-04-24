package snowflake

import "testing"

func TestNextIDIsIncreasing(t *testing.T) {
	node, err := NewNode(1)
	if err != nil {
		t.Fatalf("NewNode() error = %v", err)
	}

	previous := node.NextID()
	for i := 0; i < 1000; i++ {
		current := node.NextID()
		if current <= previous {
			t.Fatalf("NextID() = %d, want greater than %d", current, previous)
		}
		previous = current
	}
}

func TestNewNodeRejectsInvalidID(t *testing.T) {
	if _, err := NewNode(-1); err == nil {
		t.Fatal("NewNode(-1) error = nil, want error")
	}
	if _, err := NewNode(maxNodeID + 1); err == nil {
		t.Fatal("NewNode(maxNodeID + 1) error = nil, want error")
	}
}
