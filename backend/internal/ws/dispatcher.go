package ws

import "context"

type MessageDispatcher interface {
	Dispatch(ctx context.Context, envelope *Envelope) error
}
