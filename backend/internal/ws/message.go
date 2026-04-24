package ws

import "encoding/json"

type Envelope struct {
	Seq       string          `json:"seq"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
}
