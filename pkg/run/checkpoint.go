package run

import (
	"encoding/json"
	"io"

	"github.com/devlikebear/tessera/pkg/queue"
)

type Checkpoint struct {
	RunID   string       `json:"run_id"`
	Status  Status       `json:"status"`
	Closure ClosureKind  `json:"closure"`
	Tasks   []queue.Task `json:"tasks"`
	Events  []Event      `json:"events"`
}

func NewCheckpoint(r *Run, tasks []queue.Task) Checkpoint {
	return Checkpoint{
		RunID:   r.ID,
		Status:  r.Status,
		Closure: r.Closure,
		Tasks:   cloneTasks(tasks),
		Events:  r.Events.Events(),
	}
}

func WriteCheckpoint(w io.Writer, cp Checkpoint) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(cp)
}

func ReadCheckpoint(r io.Reader) (Checkpoint, error) {
	var cp Checkpoint
	if err := json.NewDecoder(r).Decode(&cp); err != nil {
		return Checkpoint{}, err
	}
	return cp, nil
}
