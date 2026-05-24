package observe

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/devlikebear/tessera/pkg/run"
)

type JSONLinesSink struct {
	mu      sync.Mutex
	encoder *json.Encoder
	file    *os.File
}

func NewJSONLinesSink(w io.Writer) *JSONLinesSink {
	return &JSONLinesSink{encoder: json.NewEncoder(w)}
}

func NewJSONLinesFile(path string) (*JSONLinesSink, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &JSONLinesSink{encoder: json.NewEncoder(file), file: file}, nil
}

func (s *JSONLinesSink) OnEvent(_ context.Context, event run.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.encoder.Encode(event)
}

func (s *JSONLinesSink) Close() error {
	if s.file == nil {
		return nil
	}
	return s.file.Close()
}
