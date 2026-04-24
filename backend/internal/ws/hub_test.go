package ws

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeOnlineStatus struct {
	onlineCount  int
	refreshCount int
	offlineCount int
}

func (f *fakeOnlineStatus) MarkOnline(context.Context, int64, time.Time) error {
	f.onlineCount++
	return nil
}

func (f *fakeOnlineStatus) RefreshOnline(context.Context, int64) error {
	f.refreshCount++
	return nil
}

func (f *fakeOnlineStatus) MarkOffline(context.Context, int64) error {
	f.offlineCount++
	return nil
}

func TestHubKeepsOnlineStatusUntilLastConnectionUnregisters(t *testing.T) {
	online := &fakeOnlineStatus{}
	hub := NewHub(online)
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(runCtx)

	clientA := &Client{UserID: 123, Send: make(chan []byte, 1)}
	clientB := &Client{UserID: 123, Send: make(chan []byte, 1)}
	ctx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()

	if err := hub.Register(ctx, clientA); err != nil {
		t.Fatalf("register client A: %v", err)
	}
	if err := hub.Register(ctx, clientB); err != nil {
		t.Fatalf("register client B: %v", err)
	}
	if online.onlineCount != 2 {
		t.Fatalf("online writes = %d, want 2", online.onlineCount)
	}

	if err := hub.Unregister(ctx, clientA); err != nil {
		t.Fatalf("unregister client A: %v", err)
	}
	if online.offlineCount != 0 {
		t.Fatalf("offline writes after first unregister = %d, want 0", online.offlineCount)
	}
	if _, ok := <-clientA.Send; ok {
		t.Fatal("client A send channel should be closed")
	}

	if err := hub.Unregister(ctx, clientB); err != nil {
		t.Fatalf("unregister client B: %v", err)
	}
	if online.offlineCount != 1 {
		t.Fatalf("offline writes after last unregister = %d, want 1", online.offlineCount)
	}
}

func TestHubClosesSlowConnectionWhenSendQueueIsFull(t *testing.T) {
	hub := NewHub(&fakeOnlineStatus{})
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(runCtx)

	client := &Client{UserID: 123, Send: make(chan []byte, 1)}
	ctx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()

	if err := hub.Register(ctx, client); err != nil {
		t.Fatalf("register client: %v", err)
	}
	client.Send <- []byte("queued")

	if err := hub.SendToUser(ctx, 123, []byte("overflow")); err != nil {
		t.Fatalf("send to user: %v", err)
	}
	if _, ok := <-client.Send; !ok {
		t.Fatal("first queued payload should still be readable")
	}
	if _, ok := <-client.Send; ok {
		t.Fatal("slow client send channel should be closed after queue overflow")
	}
}

func TestMarshalEnvelope(t *testing.T) {
	payload, err := MarshalEnvelope("tmp-1", EventPong, map[string]string{})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	var envelope Envelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if envelope.Seq != "tmp-1" || envelope.Type != EventPong {
		t.Fatalf("unexpected envelope: %+v", envelope)
	}
	if envelope.Timestamp <= 0 {
		t.Fatal("timestamp should be set")
	}
	if string(envelope.Data) != "{}" {
		t.Fatalf("data = %s, want {}", string(envelope.Data))
	}
}

func TestUnknownEventEnqueuesErrorEnvelope(t *testing.T) {
	client := &Client{
		UserID: 123,
		Send:   make(chan []byte, 1),
	}

	continued := client.handleEnvelope(&Envelope{
		Seq:  "tmp-unknown",
		Type: "unknown.event",
	})
	if !continued {
		t.Fatal("unknown event with available queue should not close connection")
	}

	var envelope Envelope
	if err := json.Unmarshal(<-client.Send, &envelope); err != nil {
		t.Fatalf("unmarshal error envelope: %v", err)
	}
	if envelope.Seq != "tmp-unknown" || envelope.Type != EventError {
		t.Fatalf("unexpected envelope: %+v", envelope)
	}

	var data ErrorData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("unmarshal error data: %v", err)
	}
	if data.Code != "unsupported_event" || data.EventType != "unknown.event" {
		t.Fatalf("unexpected error data: %+v", data)
	}
}

func TestCheckOriginAllowsConfiguredOrigin(t *testing.T) {
	check := checkOrigin(Options{AllowedOrigins: []string{"https://example.com"}})
	req := httptest.NewRequest("GET", "http://api.example.com/ws", nil)
	req.Header.Set("Origin", "https://example.com")

	if !check(req) {
		t.Fatal("configured origin should be allowed")
	}
}

func TestCheckOriginRejectsUnconfiguredProductionOrigin(t *testing.T) {
	check := checkOrigin(Options{})
	req := httptest.NewRequest("GET", "http://api.example.com/ws", nil)
	req.Header.Set("Origin", "https://evil.example.com")

	if check(req) {
		t.Fatal("unconfigured origin should be rejected")
	}
}
