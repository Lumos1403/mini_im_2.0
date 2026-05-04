package agentclient

import (
	"strings"
	"testing"
)

func TestParseTolerantSSEPreservesContinuationLines(t *testing.T) {
	input := strings.Join([]string{
		"data: 第一行",
		"第二行",
		"",
		"第三行",
		"",
		"data: ![chart](https://mermaid.ink/img/abc123)",
		"",
		"data: [DONE]",
		"",
	}, "\n")

	var events []StreamEvent
	err := parseTolerantSSE(strings.NewReader(input), func(event StreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil {
		t.Fatalf("parseTolerantSSE returned error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].Content != "第一行\n第二行\n\n第三行" {
		t.Fatalf("unexpected first content: %q", events[0].Content)
	}
	if events[1].Content != "![chart](https://mermaid.ink/img/abc123)" {
		t.Fatalf("unexpected second content: %q", events[1].Content)
	}
	if !events[2].Done {
		t.Fatalf("expected done event")
	}
}
