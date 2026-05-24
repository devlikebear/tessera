package observe

import (
	"context"
	"sync"

	"github.com/devlikebear/tessera/pkg/run"
)

type Sink interface {
	OnEvent(context.Context, run.Event) error
}

type MultiSink []Sink

func (s MultiSink) OnEvent(ctx context.Context, event run.Event) error {
	for _, sink := range s {
		if sink == nil {
			continue
		}
		if err := sink.OnEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

type MemorySink struct {
	mu     sync.Mutex
	events []run.Event
}

func (s *MemorySink) OnEvent(_ context.Context, event run.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *MemorySink) Events() []run.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]run.Event, len(s.events))
	copy(out, s.events)
	return out
}
